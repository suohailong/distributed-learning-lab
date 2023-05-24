package cluster

import (
	"crypto/sha256"
	"distributed-learning-lab/harmoniakv/node"
	"encoding/binary"
	"sort"
	"strconv"
	"sync"
)

/*
*******
需要实现的功能:
数据分片
数据复制
一致性
不一致的解决方案
失效处理
********
*/
type Cluster interface {
	Start()
	HandleGossipMessage(*GossipStateMessage)

	AddNode(*node.Node)
	RemoveNode(*node.Node)
	GetReplicas([]byte, int) map[string]struct{}
	GetNode([]byte) *node.Node
	NodeLen() int

	GetNodeByID(id string)
}

type defaultCluster struct {
	sync.RWMutex
	//当前node
	currentNode *node.Node

	vnodesMap map[uint32]*node.Node
	// hash环
	vnodes []uint32
	// 物理节点
	pnodes []*node.Node
	// 随机发送信息的node个数
	randomNode int
	// 虚拟节点个数
	vCount int
	// 副本数
	replica int
}

func (c *defaultCluster) SetCurrentNode(node *node.Node) {
	c.Lock()
	defer c.Unlock()
}

func (c *defaultCluster) AddNode(n *node.Node) {
	c.Lock()
	defer c.Unlock()
	c.pnodes = append(c.pnodes, n)
	for i := 0; i < c.vCount; i++ {
		vn := n.GetId() + "-" + strconv.Itoa(i)
		hash := c.hash(vn)
		c.vnodesMap[hash] = n
		c.vnodes = append(c.vnodes, hash)
	}
	sort.Slice(c.vnodes, func(i, j int) bool {
		// 从小到大排列
		return c.vnodes[i] < c.vnodes[j]
	})
	sort.Slice(c.pnodes, func(i, j int) bool {
		// 从小到大排列
		return c.pnodes[i].GetId() < c.pnodes[j].GetId()
	})
}

func (c *defaultCluster) RemoveNode(n *node.Node) {
	c.Lock()
	defer c.Unlock()
	for i := 0; i < c.vCount; i++ {
		vn := n.GetId() + "-" + strconv.Itoa(i)
		targetHash := c.hash(vn)
		delete(c.vnodesMap, targetHash)
		for index, hash := range c.vnodes {
			if hash == targetHash {
				c.vnodes = append(c.vnodes[:index], c.vnodes[index+1:]...)
			}
		}
	}
	// 要求待搜索的切片已经按照升序排列
	index := sort.Search(len(c.pnodes), func(i int) bool {
		// 找到第一个大于等于node的节点
		return c.pnodes[i].GetId() >= n.GetId()
	})
	c.pnodes = append(c.pnodes[:index], c.pnodes[index+1:]...)
}

func (c *defaultCluster) GetReplicas(key []byte, count int) (target map[string]struct{}) {
	c.RLock()
	defer c.RUnlock()
	hash := c.hash(string(key))
	index := sort.Search(len(c.vnodes), func(i int) bool {
		return c.vnodes[i] >= hash
	})

	for _, hash := range c.vnodes[index:] {
		target[c.vnodesMap[hash].GetId()] = struct{}{}
		// TODO: 跳过相同的物理机器
		// TODO: 跳过不可达的节点
		if len(target) == count {
			break
		}
	}
	return
}

func (c *defaultCluster) GetNode(key []byte) *node.Node {
	c.RLock()
	defer c.RUnlock()
	if c.NodeLen() == 0 {
		return nil
	}
	hash := c.hash(string(key))
	index := sort.Search(len(c.vnodes), func(i int) bool {
		return c.vnodes[i] >= hash
	})
	targetHash := c.vnodes[index]
	return c.vnodesMap[targetHash]
}

func (c *defaultCluster) NodeLen() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.pnodes)
}

func (d *defaultCluster) hash(key string) uint32 {
	hash := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint32(hash[:4])
}

func (d *defaultCluster) GetNodeByID(id string) {

}

// param:
// vnodeNumber: 虚拟节点个数
// randomNode: gossip 随机发送信息的node个数
// replicas: 副本的个数
func New(vnodeCount, randomNode, replicas int, n *node.Node) Cluster {
	co := &defaultCluster{
		vnodesMap:  make(map[uint32]*node.Node),
		vnodes:     make([]uint32, 0),
		pnodes:     make([]*node.Node, 0),
		randomNode: randomNode,
		vCount:     vnodeCount,
		replica:    replicas,
	}

	co.currentNode = n
	co.AddNode(n)
	return co
}
