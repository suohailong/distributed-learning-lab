package wal

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"distributed-learning-lab/util/log"
	"distributed-learning-lab/wal"

	"github.com/sirupsen/logrus"
)

type KVStore struct {
	wal  *wal.Wal
	data sync.Map
}

func NewKvStore(dir string) *KVStore {
	kv := &KVStore{
		wal: wal.CreateWal(dir),
	}
	kv.applyLog()
	return kv
}

func (kv *KVStore) applyLog() error {
	entries, err := kv.wal.ReadAll()
	if err != nil {
		return err
	}
	log.Debugf("read entries from wal: %v", entries)
	// TODO: 恢复数据
	return nil
}

func (kv *KVStore) appendLog(key, value string, cmdtype int) {
	kv.wal.Save([][]byte{})
}

func (kv *KVStore) Get(key string) string {
	data, ok := kv.data.Load(key)
	if !ok {
		return ""
	}
	return data.(string)
}

func (kv *KVStore) Put(key string, value string) {
	kv.appendLog(key, value, 0)
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
