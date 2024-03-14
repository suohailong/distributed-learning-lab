package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"

	"distributed-learning-lab/util/log"

	"github.com/golang/protobuf/proto"
)

type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type record struct {
	data []byte
	crc  uint32
}

func (r *record) Reset() {
	r.data = nil
	r.crc = 0
}

func (r *record) String() string {
	return fmt.Sprintf("Record{data: %v, crc: %v}", r.data, r.crc)
}

func (r *record) ProtoMessage() {}

func (r *record) Descriptor() ([]byte, []int) {
	return nil, nil
}

func (r *record) XXX_MessageName() string {
	return ""
}

func (r *record) Marshal() ([]byte, error) {
	return proto.Marshal(r)
}

func (r *record) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, r)
}

func (r *record) MarshalTo(data []byte) (int, error) {
	return 0, nil
}

func (r *record) Size() int {
	return len(r.data) + binary.Size(r.crc)
}

func (r *record) XXX_DiscardUnknown() {}

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
	crcTable *crc32.Table
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

func (w *Wal) ReadAll() ([]byte, error) {
	w.Lock()
	defer w.Unlock()

	stat, err := w.file.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, stat.Size())
	_, err = w.file.Read(buf)
	if err != nil {
		return nil, err
	}

	// entries := make([]byte, 0)
	// for _, line := range bytes.Split(buf, []byte("\n")) {
	// 	if len(line) == 0 {
	// 		continue
	// 	}
	// 	entries = append(entries, entry...)
	// }

	// w.entries = entries
	// w.index = len(entries)
	return buf, nil
}

func (w *Wal) sync() error {
	return w.file.Sync()
}

func (w *Wal) Save(entities [][]byte) error {
	w.Lock()
	defer w.Lock()
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
			return err
		}
		err = writeUint64(w.file, uint64(len(d)), make([]byte, 8))
		if err != nil {
			return err
		}
		_, err = w.file.Write(d)
		if err != nil {
			return err
		}
	}
	return w.sync()
}
