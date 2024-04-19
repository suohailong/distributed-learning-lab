package consistent_hash

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strconv"
	"sync"
)

type ConsistentHash interface {
	AddNode(node string)
	RemoveNode(node string)
	// 通过数据的key获取node
	GetNode(key string) string
	// 获取可以存储key的所有node列表
	GetNodes(key string, count int) []string
}

type defaultConsistentHash struct {
	sync.RWMutex
	nodes        []uint32
	nodesMap     map[uint32]string
	virtualNodes int
	nodelength   int
}

func (d *defaultConsistentHash) hash(key string) uint32 {
	hash := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint32(hash[:4])
}

func New(virtualNodes int) ConsistentHash {
	return &defaultConsistentHash{
		nodes:        []uint32{},
		nodesMap:     make(map[uint32]string),
		virtualNodes: virtualNodes,
		nodelength:   0,
	}
}

func (d *defaultConsistentHash) nodeLen() int {
	d.RLock()
	defer d.RUnlock()
	return d.nodelength
}

func (d *defaultConsistentHash) AddNode(node string) {
	d.Lock()
	defer d.Unlock()
	for i := 0; i < d.virtualNodes; i++ {
		virtualNodeId := node + "#" + strconv.Itoa(i)
		virtualNodeHash := d.hash(virtualNodeId)
		d.nodesMap[virtualNodeHash] = node
		d.nodes = append(d.nodes, virtualNodeHash)
	}
	sort.Slice(d.nodes, func(i int, j int) bool {
		return d.nodes[i] < d.nodes[j]
	})
	d.nodelength++
}

func (d *defaultConsistentHash) RemoveNode(node string) {
	d.Lock()
	defer d.Unlock()
	deleteflag := 0
	for i := 0; i < d.virtualNodes; i++ {
		virtualNodeId := node + "#" + strconv.Itoa(i)
		hash := d.hash(virtualNodeId)
		index := d.searchIndex(hash)
		if index >= 0 {
			d.nodes = append(d.nodes[:index], d.nodes[index+1:]...)
			delete(d.nodesMap, hash)
			deleteflag = 1
		}
	}
	if deleteflag > 0 {
		d.nodelength--
	}
}

func (d *defaultConsistentHash) searchIndex(hash uint32) int {
	for i, nodeHash := range d.nodes {
		if hash == nodeHash {
			return i
		}
	}
	return -1
}

func (d *defaultConsistentHash) GetNode(key string) string {
	d.RLock()
	defer d.RUnlock()
	keyHash := d.hash(key)

	index := d.searchInsertIndex(keyHash)

	return d.nodesMap[d.nodes[index%len(d.nodes)]]
}

func (d *defaultConsistentHash) GetNodes(key string, count int) (nodeIDs []string) {
	d.RLock()
	defer d.RUnlock()
	hash := d.hash(key)
	index := d.searchInsertIndex(hash)

	nodes := make(map[string]struct{})
	for len(nodes) < count && len(nodes) < len(d.nodes) {
		nodeID := d.nodesMap[d.nodes[index%len(d.nodes)]]
		nodes[nodeID] = struct{}{}
		index++
	}

	for nodeID := range nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	return
}

func (d *defaultConsistentHash) searchInsertIndex(hash uint32) int {
	index := sort.Search(len(d.nodes), func(i int) bool {
		return d.nodes[i] >= hash
	})

	return index
}
