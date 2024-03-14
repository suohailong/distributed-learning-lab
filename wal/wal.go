package wal

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"os"
	"sync"

	"distributed-learning-lab/util/log"
)

type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type defaultSerializer struct{}

func (d *defaultSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (d *defaultSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
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

func (w *Wal) save(entities [][]byte) error {
	w.Lock()
	defer w.Lock()
	if len(entities) == 0 {
		return nil
	}
	for _, entity := range entities {
		entryCrc := crc32.Checksum(entity, w.crcTable)
		_, err := w.file.Write(entity)
		if err != nil {
			return err
		}
		err = binary.Write(w.file, binary.LittleEndian, entryCrc)
		if err != nil {
			return err
		}
	}
	return w.sync()
}
