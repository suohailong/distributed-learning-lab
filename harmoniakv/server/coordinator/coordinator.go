package coordinator

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strconv"
	"sync"
)

// 管理shard,管理replicas
// node 容错
// consistency
type Coordinator interface {
}

type coordinator struct {
	sync.RWMutex
	nodes        []uint32
	nodesMap     map[uint32]string
	virtualNodes int
	nodelength   int
}

func (d *coordinator) hash(key string) uint32 {
	hash := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint32(hash[:4])
}

func New(virtualNodes int) Coordinator {
	return &coordinator{
		nodes:        []uint32{},
		nodesMap:     make(map[uint32]string),
		virtualNodes: virtualNodes,
		nodelength:   0,
	}
}

func (d *coordinator) nodeLen() int {
	d.RLock()
	defer d.RUnlock()
	return d.nodelength
}

func (d *coordinator) AddNode(node string) {
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

func (d *coordinator) RemoveNode(node string) {
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

func (d *coordinator) searchIndex(hash uint32) int {
	for i, nodeHash := range d.nodes {
		if hash == nodeHash {
			return i
		}
	}
	return -1
}

func (d *coordinator) GetNode(key string) string {
	d.RLock()
	defer d.RUnlock()
	keyHash := d.hash(key)

	index := d.searchInsertIndex(keyHash)

	return d.nodesMap[d.nodes[index%len(d.nodes)]]
}

func (d *coordinator) GetNodes(key string, count int) (nodeIDs []string) {
	d.RLock()
	defer d.RUnlock()
	hash := d.hash(key)
	index := d.searchInsertIndex(hash)

	nodes := make(map[string]struct{})
	for len(nodes) < count && len(nodes) < len(d.nodes) {
		// 这里是顺着环再往下找,直到到了count个
		// TODO: 这里还得判断一下尽量避免选出的节点都是不同的节点
		nodeID := d.nodesMap[d.nodes[index%len(d.nodes)]]
		nodes[nodeID] = struct{}{}
		index++
	}

	for nodeID := range nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	return

}

func (d *coordinator) searchInsertIndex(hash uint32) int {
	index := sort.Search(len(d.nodes), func(i int) bool {
		return d.nodes[i] >= hash
	})

	return index
}
