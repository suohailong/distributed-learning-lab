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

// TODO: 什么时候刷新到磁盘, 立即刷新, 因为是wal，不立即刷新到磁盘就没有持久化可言
// TODO: 如果日志文件损坏了，怎么办
// TODO: 日志文件是append only， 如果客户端由于网络等原因失败后重试， 需要append 提供幂等性
// TODO: Wal 日志低于低水位线的可以删掉
type Wal struct {
	sync.RWMutex
	index int
	// entities []byte
	file *os.File
	// ds       Serializer
	crcTable    *crc32.Table
	writeOffset int64
}

func CreateWal(logName string) *Wal {
	f, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		log.Fatalf("open file %s failed: %v", logName, err)
	}
	log.Debugf("open file: %s successfully", logName)

	wal := &Wal{
		index:    0,
		file:     f,
		crcTable: crc32.MakeTable(crc32.Castagnoli),
	}

	return wal
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
	}
	return w.sync()
}
