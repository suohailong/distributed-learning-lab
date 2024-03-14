package wal

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"distributed-learning-lab/util/log"

	"github.com/sirupsen/logrus"
)

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

func TestKvStore(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	kv := NewKvStore("testdata/waltest.log")

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		// kv.Put(key, value)
		v := kv.Get(key)
		fmt.Println("value: ", v)
		if v != value {
			t.Fatal("kv store get failed")
		}
	}
	time.Sleep(2 * time.Second)
}
