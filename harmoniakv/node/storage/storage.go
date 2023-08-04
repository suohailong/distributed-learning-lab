package storage

import (
	"distributed-learning-lab/harmoniakv/node/version"

	badger "github.com/dgraph-io/badger/v3"
)

type Storage interface {
	Put(key []byte, value version.Value)
	Get(key []byte) []*version.Value
}

type defaultStorage struct {
	db *badger.DB
}

func (d *defaultStorage) Put(key []byte, value version.Value) error {
	return d.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, []byte("42"))
		err := txn.SetEntry(e)
		return err
	})
}

func (d *defaultStorage) Get(key []byte) []*version.Value {
	panic("not implemented") // TODO: Implement
}
