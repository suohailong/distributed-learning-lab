package memory

import (
	"distributed-learning-lab/harmoniakv/node/version"
	herror "distributed-learning-lab/harmoniakv/util/error"
	"sync"
)

type memoryStore struct {
	sync.Map
}

func (m *memoryStore) Put(key []byte, value version.Value) (e error) {
	m.Store(key, value)
	if _, ok := m.Load(key); !ok {
		return herror.New(herror.ErrPutFailed, "put failed")
	}
	return
}

func (m *memoryStore) Get(key []byte) (_ []*version.Value) {
	panic("not implemented") // TODO: Implement
}
