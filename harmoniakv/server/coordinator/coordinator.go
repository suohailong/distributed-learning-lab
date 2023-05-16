package coordinator

import (
	"crypto/sha256"
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
type Coordinator interface {
	AddNode(*Node)
	RemoveNode(*Node)
	GetReplicas(string, int) []string
	GetNode(string) *Node
	NodeLen() int
}

type coordinator struct {
	sync.RWMutex
	//当前node
	currentNode *Node

	vnodesMap map[uint32]*Node
	// hash环
	vnodes []uint32
	// 物理节点
	pnodes []*Node
	// 随机发送信息的node个数
	randomNode int
	// 虚拟节点个数
	vCount int
	// 副本数
	replica int
}

func (c *coordinator) AddNode(node *Node) {
	c.Lock()
	defer c.Unlock()
	c.pnodes = append(c.pnodes, node)
	for i := 0; i < c.vCount; i++ {
		vn := node.GetID() + "-" + strconv.Itoa(i)
		hash := c.hash(vn)
		c.vnodesMap[hash] = node
		c.vnodes = append(c.vnodes, hash)
	}
	sort.Slice(c.vnodes, func(i, j int) bool {
		// 从小到大排列
		return c.vnodes[i] < c.vnodes[j]
	})
	sort.Slice(c.pnodes, func(i, j int) bool {
		// 从小到大排列
		return c.pnodes[i].GetID() < c.pnodes[j].GetID()
	})
}

func (c *coordinator) RemoveNode(node *Node) {
	c.Lock()
	defer c.Unlock()
	for i := 0; i < c.vCount; i++ {
		vn := node.GetID() + "-" + strconv.Itoa(i)
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
		return c.pnodes[i].GetID() >= node.GetID()
	})
	c.pnodes = append(c.pnodes[:index], c.pnodes[index+1:]...)
}

func (c *coordinator) GetReplicas(key string, count int) []string {
	c.RLock()
	defer c.RUnlock()
	hash := c.hash(key)
	targetNode := make([]string, 0)
	index := sort.Search(len(c.vnodes), func(i int) bool {
		return c.vnodes[i] >= hash
	})

	for _, hash := range c.vnodes[index:] {
		targetNode = append(targetNode, c.vnodesMap[hash].GetID())
		if len(targetNode) == count {
			break
		}
	}

	return targetNode
}

func (c *coordinator) GetNode(key string) *Node {
	c.RLock()
	defer c.RUnlock()
	if c.NodeLen() == 0 {
		return nil
	}
	hash := c.hash(key)
	index := sort.Search(len(c.vnodes), func(i int) bool {
		return c.vnodes[i] >= hash
	})
	targetHash := c.vnodes[index]
	return c.vnodesMap[targetHash]
}

func (c *coordinator) NodeLen() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.pnodes)
}

func (d *coordinator) hash(key string) uint32 {
	hash := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint32(hash[:4])
}

// param:
// vnodeNumber: 虚拟节点个数
// randomNode: gossip 随机发送信息的node个数
// replicas: 副本的个数
func New(vnodeCount, randomNode, replicas int) Coordinator {

	return &coordinator{
		vnodesMap:  make(map[uint32]*Node),
		vnodes:     make([]uint32, 0),
		pnodes:     make([]*Node, 0),
		randomNode: randomNode,
		vCount:     vnodeCount,
		replica:    replicas,
	}
}
