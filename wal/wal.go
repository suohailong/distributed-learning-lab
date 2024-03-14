package wal

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"os"
	"sync"
	"time"

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

type EntryType uint8

const (
	EntryTypePut    EntryType = iota
	EntryTypeUpdate EntryType = iota
	EntryTypeDelete EntryType = iota
)

type WalEntry struct {
	EntryIndex uint64    `json:"entryIndex"`
	Key        []byte    `json:"key"`
	Value      []byte    `json:"value"`
	EntryType  EntryType `json:"entryType"`
	TimeStamp  uint64    `json:"timeStamp"`
}

// TODO: 什么时候刷新到磁盘, etcd相当于小批量刷新
// TODO: 如果日志文件损坏了，怎么办
// TODO: 日志文件是append only， 如果客户端由于网络等原因失败后重试， 需要append 提供幂等性
// TODO: writeAheadLog 日志低于低水位线的可以删掉
type WriteAheadLog struct {
	sync.RWMutex
	index    int
	entries  []*WalEntry
	file     *os.File
	ds       Serializer
	crcTable *crc32.Table
}

func NewWriteAheadLog(logName string) *WriteAheadLog {
	f, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		log.Fatalf("open file %s failed: %v", logName, err)
	}
	log.Debugf("open file: %s successfully", logName)

	wal := &WriteAheadLog{
		index:    0,
		entries:  make([]*WalEntry, 0),
		file:     f,
		ds:       &defaultSerializer{},
		crcTable: crc32.MakeTable(crc32.Castagnoli),
	}

	// 定时刷新到磁盘
	go func() {
		for {
			time.Sleep(1 * time.Second)
			wal.SaveToDisk()
		}
	}()

	return wal
}

func (wal *WriteAheadLog) ReadLogAll() ([]*WalEntry, error) {
	wal.Lock()
	defer wal.Unlock()

	stat, err := wal.file.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, stat.Size())
	_, err = wal.file.Read(buf)
	if err != nil {
		return nil, err
	}

	entries := make([]*WalEntry, 0)
	for _, line := range bytes.Split(buf, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		entry := &WalEntry{}
		err = wal.ds.Unmarshal(line, entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	// wal.entries = entries
	wal.index = len(entries)
	return entries, nil
}

func (wal *WriteAheadLog) WriteEntry(key, value string, entryType EntryType) {
	wal.Lock()
	defer wal.Unlock()
	entry := &WalEntry{
		EntryIndex: uint64(wal.index),
		TimeStamp:  uint64(time.Now().Unix()),
		Key:        []byte(key),
		Value:      []byte(value),
		EntryType:  entryType,
	}
	wal.entries = append(wal.entries, entry)
	wal.index++
}

func (wal *WriteAheadLog) SaveToDisk() error {
	wal.RLock()
	defer wal.RUnlock()

	if len(wal.entries) == 0 {
		return nil
	}

	for _, entry := range wal.entries {
		byteData, err := wal.ds.Marshal(entry)
		if err != nil {
			log.Fatalf("Failed to marshal entry: %v", err)
		}
		entryCrc := crc32.Checksum(byteData, wal.crcTable)
		_, err = wal.file.Write(byteData)
		if err != nil {
			log.Fatalf("Failed to write entry to file: %v", err)
		}

		binary.Write(wal.file, binary.LittleEndian, entryCrc)
	}
	// 这里同步完就删掉当前的日志
	wal.entries = make([]*WalEntry, 0)

	err := wal.file.Sync()
	if err != nil {
		return err
	}

	return nil
}
