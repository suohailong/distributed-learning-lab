package main

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
)

type Node struct {
	Id     int
	Weight int
	data   map[string]interface{}
}

type Ring []*Node

func (r Ring) Len() int {
	return len(r)
}

func (r Ring) Less(i, j int) bool {
	return r[i].Weight < r[j].Weight
}

func (r Ring) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type ConsistentHash struct {
	Nodes Ring
}

func NewConsistentHash() *ConsistentHash {
	return &ConsistentHash{Nodes: Ring{}}
}

func (c *ConsistentHash) AddNode(node *Node) {
	c.Nodes = append(c.Nodes, node)
	sort.Sort(c.Nodes)
}

func (c *ConsistentHash) RemoveNode(node *Node) {
	for i, n := range c.Nodes {
		if n.Id == node.Id {
			c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
		}
	}

}

func (c *ConsistentHash) Get(key string) *Node {
	if len(c.Nodes) == 0 {
		return nil
	}

	hash := int(crc32.ChecksumIEEE([]byte(key)))
	idx := sort.Search(len(c.Nodes), func(i int) bool {
		return c.Nodes[i].Weight >= hash
	})

	if idx == len(c.Nodes) {
		idx = 0
	}

	return c.Nodes[idx]
}

func (c *ConsistentHash) GetReplicas(key string, numReplicas int) []*Node {
	if len(c.Nodes) == 0 || numReplicas <= 0 {
		return nil
	}

	hash := int(crc32.ChecksumIEEE([]byte(key)))
	idx := sort.Search(len(c.Nodes), func(i int) bool {
		return c.Nodes[i].Weight >= hash
	})

	replicas := make([]*Node, 0, numReplicas)
	for i := 0; i < numReplicas; i++ {
		replicas = append(replicas, c.Nodes[idx])

		idx++
		if idx == len(c.Nodes) {
			idx = 0
		}
	}

	return replicas
}

func main() {
	ch := NewConsistentHash()

	for i := 1; i <= 10; i++ {
		node := &Node{
			Id:     i,
			Weight: int(crc32.ChecksumIEEE([]byte(strconv.Itoa(i)))),
		}
		ch.AddNode(node)
	}

	key := "my-key"
	numReplicas := 3

	node := ch.Get(key)
	fmt.Printf("Key '%s' is mapped to Node %d\n", key, node.Id)

	replicas := ch.GetReplicas(key, numReplicas)
	fmt.Printf("Key '%s' has %d replicas in the following nodes:\n", key, numReplicas)
	for _, replica := range replicas {
		fmt.Printf("Node %d\n", replica.Id)
	}

	ch.RemoveNode(&Node{Id: 2})

	replicas = ch.GetReplicas(key, numReplicas)
	fmt.Printf("Key '%s' has %d replicas in the following nodes:\n", key, numReplicas)
	for _, replica := range replicas {
		fmt.Printf("after remove node 2,  Node %d\n", replica.Id)
	}

	ch.AddNode(&Node{
		Id:     2,
		Weight: int(crc32.ChecksumIEEE([]byte(strconv.Itoa(2)))),
	})
	replicas = ch.GetReplicas(key, numReplicas)
	fmt.Printf("Key '%s' has %d replicas in the following nodes:\n", key, numReplicas)
	for _, replica := range replicas {
		fmt.Printf("after add node 2,  Node %d\n", replica.Id)
	}

	// 失败检测
	// go func() {
	// 假设节点2挂了
	// 从hash环上摘掉
	// 新来的请求,调用GetReplicas 获取节点就好.
	// 这时会向下再选一个节点作为临时副本
	// 故障期间的数据都会落到这个临时副本上
	// 当节点2恢复后, 会再次调用GetReplicas, 重新分配数据

	// }()

}
