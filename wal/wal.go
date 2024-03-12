package wal

import (
	"bytes"
	"encoding/json"
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

type WriteAheadLog struct {
	sync.RWMutex
	index   int
	entries []*WalEntry
	file    *os.File
	ds      Serializer
}

func NewWriteAheadLog(logName string) *WriteAheadLog {
	f, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		log.Fatalf("open file %s failed: %v", logName, err)
	}
	log.Debugf("open file: %s successfully", logName)

	wal := &WriteAheadLog{
		index:   0,
		entries: make([]*WalEntry, 0),
		file:    f,
		ds:      &defaultSerializer{},
	}

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
	wal.entries = append(wal.entries, &WalEntry{
		EntryIndex: uint64(wal.index),
		TimeStamp:  uint64(time.Now().Unix()),
		Key:        []byte(key),
		Value:      []byte(value),
		EntryType:  entryType,
	})
	wal.index++
}

func (wal *WriteAheadLog) SaveToDisk() error {
	wal.RLock()
	defer wal.RUnlock()

	if len(wal.entries) == 0 {
		return nil
	}

	for _, entry := range wal.entries {
		str, err := wal.ds.Marshal(entry)
		if err != nil {
			log.Fatalf("Failed to marshal entry: %v", err)
		}
		str = append(str, []byte("\n")...)
		_, err = wal.file.Write(str)
		if err != nil {
			log.Fatalf("Failed to write entry to file: %v", err)
		}
	}

	err := wal.file.Sync()
	if err != nil {
		return err
	}

	return nil
}

type KVStore struct {
	wal  *WriteAheadLog
	data sync.Map
}

func NewKvStore(wal string) *KVStore {
	kv := &KVStore{
		wal: NewWriteAheadLog(wal),
	}
	kv.applyLog()
	return kv
}

func (kv *KVStore) applyLog() error {
	entries, err := kv.wal.ReadLogAll()
	if err != nil {
		return err
	}
	log.Debugf("read entries from wal: %v", entries)
	for _, entry := range entries {
		switch entry.EntryType {
		case EntryTypePut:
			kv.data.Store(string(entry.Key), string(entry.Value))
		}
	}
	return nil
}

func (kv *KVStore) appendLog(key, value string, cmdtype EntryType) {
	kv.wal.WriteEntry(key, value, cmdtype)
}

func (kv *KVStore) Get(key string) string {
	data, ok := kv.data.Load(key)
	if !ok {
		return ""
	}
	return data.(string)
}

func (kv *KVStore) Put(key string, value string) {
	kv.appendLog(key, value, EntryTypePut)
	kv.data.Store(key, value)
}
