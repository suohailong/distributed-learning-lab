package wal

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"distributed-learning-lab/util/log"
	"distributed-learning-lab/wal"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type KvData struct {
	Key     []byte
	Value   []byte
	Cmdtype int
}

const separator = "|"

func (kd *KvData) Marshal() ([]byte, error) {
	// Convert the struct fields to bytes
	keyBytes := []byte(kd.Key)
	valueBytes := []byte(kd.Value)
	cmdtypeBytes := []byte{byte(kd.Cmdtype)}
	// Concatenate the bytes with separator
	result := bytes.Join([][]byte{keyBytes, valueBytes, cmdtypeBytes}, []byte(separator))

	return result, nil
}

func (kd *KvData) UnMarshal(data []byte) error {
	parts := bytes.Split(data, []byte(separator))
	if len(parts) != 3 {
		return errors.New("invalid data format")
	}

	kd.Key = parts[0]
	kd.Value = parts[1]
	kd.Cmdtype = int(parts[2][0])
	return nil
}

const (
	CmdPut = iota
	cmdDel
)

type KVStore struct {
	wal  *wal.Wal
	data sync.Map
}

func NewKvStore(dir string) (*KVStore, error) {
	w, err := wal.CreateWal(dir, 512)
	if err != nil {
		return nil, err
	}
	kv := &KVStore{
		wal: w,
	}
	err = kv.applyLog()
	if err != nil {
		return nil, err
	}
	return kv, nil
}

func (kv *KVStore) Stop() error {
	return kv.wal.Close()
}

func (kv *KVStore) applyLog() error {
	entities, err := kv.wal.ReadAll(0)
	if err != nil {
		log.Errorf("read all failed: %v", err)
		return err
	}
	for _, entity := range entities {
		kd := &KvData{}
		err := kd.UnMarshal(entity)
		if err != nil {
			log.Errorf("unmarshal entity failed: %v", err)
			return err
		}
		switch kd.Cmdtype {
		case CmdPut:
			kv.data.Store(string(kd.Key), string(kd.Value))
		}

	}
	return nil
}

func (kv *KVStore) appendLog(key, value string, cmdtype int) error {
	kd := &KvData{
		Key:     []byte(key),
		Value:   []byte(value),
		Cmdtype: cmdtype,
	}
	data, err := kd.Marshal()
	if err != nil {
		return err
	}
	return kv.wal.Write([][]byte{data})
}

func (kv *KVStore) Get(key string) string {
	data, ok := kv.data.Load(key)
	if !ok {
		return ""
	}
	return data.(string)
}

func (kv *KVStore) Put(key string, value string) error {
	err := kv.appendLog(key, value, CmdPut)
	if err != nil {
		log.Errorf("append log failed: %v", err)
		return err
	}
	kv.data.Store(key, value)
	return nil
}

func TestKvStore(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	kv, err := NewKvStore("testdata/waltest.log")
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		err := kv.Put(key, value)
		assert.NoError(t, err)
	}
	kv.Stop()

	kv1, err := NewKvStore("testdata/waltest.log")
	assert.NoError(t, err)
	defer kv1.Stop()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := kv1.Get(key)
		log.Debugf("key: %s, value: %s", key, value)
	}
}
