package coordinator

import (
	"distributed-learning-lab/HarmoniaKV/server/gossip"
	hash "distributed-learning-lab/go-consistent-hash"
)

// 管理shard,管理replicas
// node 容错
// consistency
type Coordinator interface {
	hash.ConsistentHash
	gossip.Gossip
}
