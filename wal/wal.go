package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"

	"distributed-learning-lab/util/log"
)

var ErrCRCMismatch = errors.New("wal: crc mismatch")

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

// 功能
// 已完成:
// NOTE: 什么时候刷新到磁盘, 立即刷新, 因为是wal，不立即刷新到磁盘就没有持久化可言
// NOTE: 如果日志文件损坏了，怎么办, crc

// 未完成(TODO:)
// 1.日志分段
// 2.Wal 日志低于低水位线的可以删掉
// 3.日志文件是append only， 如果客户端由于网络等原因失败后重试， 需要append 提供幂等性

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
	// 当前段写入偏移
	writeOffset int64
	// 总偏移
	totalOffset int64
	// 之前所有的段文件
	segmentFiles []*os.File
	// 每个段的最大大小
	maxSegmentSize int64
}

func CreateWal(dir string) (*Wal, error) {
	w := &Wal{
		crcTable: crc32.MakeTable(crc32.Castagnoli),
	}
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

func (w *Wal) readRecord() ([]byte, error) {
	readInt64 := func(r io.Reader) (uint64, error) {
		var n uint64
		err := binary.Read(r, binary.LittleEndian, &n)
		return n, err
	}
	n64, err := readInt64(w.file)
	if err != nil {
		return []byte{}, err
	}
	buf := make([]byte, n64)
	_, err = io.ReadFull(w.file, buf)
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

func (w *Wal) ReadAll() ([][]byte, error) {
	_, err := w.file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	records := make([][]byte, 0)
	for {
		record, err := w.readRecord()
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
	return records, nil
}

func (w *Wal) sync() error {
	return w.file.Sync()
}

func (w *Wal) Save(entities [][]byte) error {
	if len(entities) == 0 {
		return nil
	}
	writeUint64 := func(w io.Writer, n uint64, buf []byte) error {
		binary.LittleEndian.PutUint64(buf, n)
		_, err := w.Write(buf)
		return err
	}
	_, err := w.file.Seek(w.writeOffset, 0)
	if err != nil {
		log.Errorf("seek file failed: %v", err)
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
		w.writeOffset += int64(len(d)) + 8
		w.totalOffset += int64(len(d)) + 8
		w.maybeRoll()
	}
	return w.sync()
}

func (w *Wal) getSegmentName() string {
	return fmt.Sprintf("%s/%d.wal", w.dir, w.writeOffset)
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
	w.writeOffset = 0
	return nil
}

// 日志分段
func (w *Wal) maybeRoll() error {
	// FIXME: 这里有个问题， w.writeOffset是实际写入的大小(包含了crc和数据长度), maxSegmentSize是实际数据的大小(不包含crc和数据长度)
	if w.writeOffset > w.maxSegmentSize {

		currOff, err := w.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
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
