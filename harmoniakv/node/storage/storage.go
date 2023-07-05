package storage

import "distributed-learning-lab/harmoniakv/node/version"

type Storage interface {
	Put(key []byte, value version.Value)
}
