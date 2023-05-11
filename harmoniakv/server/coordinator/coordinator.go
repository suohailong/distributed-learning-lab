package coordinator

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strconv"
	"sync"
)

/********
需要实现的功能:
数据分片
数据复制
一致性
不一致的解决方案
失效处理
*********/
type Coordinator interface {
}

type coordinator struct {
	sync.RWMutex
	// 物理节点列表
	nodes []*Node
	// 物理节点id到物理节点的映射
	nodeMap map[string]*Node
	// 虚拟节点列表
	virtualNodes []uint32
	// 虚拟id到物理节点的映射
	vnodesMap map[uint32]*int
	// 虚拟节点个数
	vnodeLen int
	// 物理节点个数
	pnodeLen int
	// Implementation details
	currentNode *Node
	// 随机发送信息的node个数
	randomSet int
}

func (d *coordinator) hash(key string) uint32 {
	hash := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint32(hash[:4])
}

func New(virtualNodes int) Coordinator {
	return &coordinator{
		// uint32 代表虚拟hash值
		nodes: []uint32{},
		//hash 对应的真实节点
		nodesMap:     make(map[uint32]*Node),
		virtualNodes: virtualNodes,
		nodelength:   0,
	}
}

func (d *coordinator) nodeLen() int {
	d.RLock()
	defer d.RUnlock()
	return d.nodelength
}

func (d *coordinator) AddNode(node *Node) {
	d.Lock()
	defer d.Unlock()
	for i := 0; i < d.virtualNodes; i++ {
		virtualNodeId := node.GetID() + "#" + strconv.Itoa(i)
		virtualNodeHash := d.hash(virtualNodeId)
		d.nodesMap[virtualNodeHash] = node
		d.nodes = append(d.nodes, virtualNodeHash)
	}
	sort.Slice(d.nodes, func(i int, j int) bool {
		return d.nodes[i] < d.nodes[j]
	})
	d.nodelength++
}

func (d *coordinator) RemoveNode(node *Node) {
	d.Lock()
	defer d.Unlock()
	deleteflag := 0
	for i := 0; i < d.virtualNodes; i++ {
		virtualNodeId := node.GetID() + "#" + strconv.Itoa(i)
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

func (d *coordinator) GetNode(key string) *Node {
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
		node := d.nodesMap[d.nodes[index%len(d.nodes)]]
		nodes[node.GetID()] = struct{}{}
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
