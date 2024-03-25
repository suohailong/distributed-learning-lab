package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"distributed-learning-lab/util/log"
)

// WAL接口
type WAL interface {
	// 写入数据
	Save(entities [][]byte) error

	// 读取数据
	Read(offset int64, length int) ([]byte, error)

	// 读取所有数据
	ReadAll() ([]byte, error)

	// 同步WAL日志到磁盘
	Sync() error

	// 获取最新LSN
	GetLastLSN() int64

	// 获取最近一次检查点的LSN
	GetCheckpointLSN() int64
}

var (
	ErrCRCMismatch     = errors.New("wal: crc mismatch")
	ErrFileNameInvalid = errors.New("wal name is invalid")
)

const (
	MAX_SEGMENT_SIZE = "max_segment_size"
)

type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type record struct {
	data []byte
	crc  uint32
}

func (r *record) UnMarshal(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("invalid data length")
	}
	r.crc = binary.LittleEndian.Uint32(data[:4])
	r.data = data[4:]
	return nil
}

func (r *record) Marshal() ([]byte, error) {
	dataLen := len(r.data)
	buf := make([]byte, 4+dataLen)
	binary.LittleEndian.PutUint32(buf[:4], r.crc)
	copy(buf[4:], r.data)
	return buf, nil
}

// 使用场景：wal解决的事单机事务ACID中的A和D。

// 功能
// 已完成:
// 什么时候刷新到磁盘, 立即刷新, 因为是wal，不立即刷新到磁盘就没有持久化可言
// 如果日志文件损坏了，怎么办, crc
// 日志分段
// Wal 日志低于低水位线的可以删掉, 低水位就是一个阈值可以是一个索引，也可以是一个时间值

// 优化
// TODO:
// 1. 文件锁定
// 2. 存储优化， 写入的时候，按页写入， 能防止不完整的写入，并且减少磁盘碎片
// 3. 写入的时候保证一帧的数据是8字节或4字节的倍数， 以便于提高性能， 防止数据撕裂

// 质量要求
type Wal struct {
	sync.RWMutex
	dir string
	// 当前段
	file     *os.File
	crcTable *crc32.Table
	// 总偏移
	totalOffset int64
	// 之前所有的段文件
	segmentFiles []*os.File
	// 每个段的最大大小
	maxSegmentSize int64
}

func CreateWal(dir string, segmentSize int64) (*Wal, error) {
	w := &Wal{
		dir:            dir,
		totalOffset:    0,
		crcTable:       crc32.MakeTable(crc32.Castagnoli),
		maxSegmentSize: segmentSize,
	}
	// TODO: 这里需要从文件索引恢复totalOffset
	if err := w.OpenSegment(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Wal) Close() error {
	if err := w.sync(); err != nil {
		return err
	}
	return w.file.Close()
}

func (w *Wal) readRecord(f io.Reader) ([]byte, error) {
	readInt64 := func(r io.Reader) (uint64, error) {
		var n uint64
		err := binary.Read(r, binary.LittleEndian, &n)
		return n, err
	}
	n64, err := readInt64(f)
	if err != nil {
		return []byte{}, err
	}
	buf := make([]byte, n64)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return []byte{}, err
	}
	r := &record{}
	err = r.UnMarshal(buf)
	if err != nil {
		return []byte{}, err
	}
	crc := crc32.Checksum(r.data, w.crcTable)
	if crc != r.crc {
		return []byte{}, ErrCRCMismatch
	}

	return r.data, nil
}

func (w *Wal) getSegmentNames() ([]string, error) {
	fns, err := os.ReadDir(w.dir)
	fileNames := make([]string, 0)
	if err != nil {
		return fileNames, err
	}
	sort.Slice(fns, func(i, j int) bool {
		return fns[i].Name() < fns[j].Name()
	})
	for _, fn := range fns {
		fileName := fmt.Sprintf("%s/%s", w.dir, fn.Name())
		fileNames = append(fileNames, fileName)
	}
	return fileNames, nil
}

func (w *Wal) selectFiles(names []string, offset int64) ([]string, int64, error) {
	if offset < 0 {
		return names, 0, nil
	}
	index := -1
	seq := int64(0)
	for i := len(names) - 1; i >= 0; i-- {
		// var dir, segStr string
		var segNo int64
		name := strings.Split(names[i], "/")
		_, err := fmt.Sscanf(name[1], "segment_%d.wal", &segNo)
		if err != nil {
			return []string{}, 0, err
		}
		if offset < segNo {
			continue
		} else {
			index = i
			seq = segNo
			break
		}
	}
	if index < 0 {
		return []string{}, 0, nil
	}

	return names[index:], seq, nil
}

func (w *Wal) CleanWal(offset int64) error {
	fns, err := w.getSegmentNames()
	if err != nil {
		return err
	}
	_, seq, err := w.selectFiles(fns, offset)
	if err != nil {
		return err
	}
	for _, name := range fns {
		var segNo int64
		n := strings.Split(name, "/")
		_, err := fmt.Sscanf(n[1], "segment_%d.wal", &segNo)
		if err != nil {
			return err
		}
		if segNo < seq {
			// TODO: 如果当前文件在segmentFiles中，先关闭segmentFiles中的之后再清除这些文件
			err = os.Remove(name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Wal) ReadAll(offset int64) ([][]byte, error) {
	// 打开所有的文件
	records := make([][]byte, 0)
	fns, err := w.getSegmentNames()
	if err != nil {
		return records, err
	}
	sFiles, seq, err := w.selectFiles(fns, offset)
	if err != nil {
		return records, err
	}
	openFiles := []*os.File{}
	defer func() {
		for _, f := range openFiles {
			f.Close()
		}
	}()

	for i, name := range sFiles {
		f, err := os.OpenFile(name, os.O_RDONLY, 0o644)
		if err != nil {
			return records, err
		}
		openFiles = append(openFiles, f)
		// 如果是第一个文件，offset可能是中间位置， 所以直接定位到offset
		if i == 0 && offset > seq {
			_, err := f.Seek(offset, io.SeekStart)
			if err != nil {
				return records, err
			}
		}
		for {
			record, err := w.readRecord(f)
			if err == io.EOF {
				log.Debugf("read all records from wal")
				break
			}
			if err != nil {
				log.Errorf("read record from wal failed: %v", err)
				return nil, err
			}
			records = append(records, record)
		}
	}

	return records, nil
}

func (w *Wal) sync() error {
	return w.file.Sync()
}

// TEST:
func (w *Wal) Save(entities [][]byte) error {
	if len(entities) == 0 {
		return nil
	}
	writeUint64 := func(w io.Writer, n uint64, buf []byte) error {
		binary.LittleEndian.PutUint64(buf, n)
		_, err := w.Write(buf)
		return err
	}
	for _, entity := range entities {
		rec := &record{
			data: entity,
		}
		rec.crc = crc32.Checksum(entity, w.crcTable)
		d, err := rec.Marshal()
		if err != nil {
			log.Errorf("marshal record failed: %v", err)
			return err
		}
		err = writeUint64(w.file, uint64(len(d)), make([]byte, 8))
		if err != nil {
			log.Errorf("write record length failed: %v", err)
			return err
		}
		_, err = w.file.Write(d)
		if err != nil {
			log.Errorf("write record failed: %v", err)
			return err
		}
		w.totalOffset += int64(len(d)) + 8
		w.maybeRoll()
	}
	return w.sync()
}

func (w *Wal) getSegmentName() string {
	return fmt.Sprintf("%s/segment_%d.wal", w.dir, w.totalOffset)
}

func (w *Wal) OpenSegment() error {
	sn := w.getSegmentName()
	f, err := os.OpenFile(sn, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	log.Debugf("open file: %s successfully", sn)
	w.file = f
	w.segmentFiles = append(w.segmentFiles, f)
	return nil
}

// 日志分段
func (w *Wal) maybeRoll() error {
	// FIXME: 这里有个问题， w.writeOffset是实际写入的大小(包含了crc和数据长度), maxSegmentSize是实际数据的大小(不包含crc和数据长度)
	curoff, err := w.file.Seek(0, io.SeekCurrent)
	fmt.Println(curoff)
	if err != nil {
		return err
	}
	if curoff > w.maxSegmentSize {
		currOff, err := w.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		// fmt.Println(curoff, w.maxSegmentSize)
		err = w.file.Truncate(currOff)
		if err != nil {
			return err
		}

		err = w.sync()
		if err != nil {
			log.Errorf("")
			return err
		}

		err = w.OpenSegment()
		if err != nil {
			log.Errorf("")
			return err

		}
	}
	return nil
}
