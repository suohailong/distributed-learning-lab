package node

import (
	"strconv"
	"sync"
)

type Node struct {
	sync.RWMutex
	host string
	data map[string]string
}

func New(hostOrIp string) *Node {
	return &Node{
		host: hostOrIp,
		data: make(map[string]string),
	}
}

func (n *Node) Set(key, value string) {
	n.Lock()
	defer n.Unlock()
	n.data[key] = value
}

func (n *Node) Get(key string) string {
	n.RLock()
	defer n.RUnlock()
	return n.data[key]
}

func (n *Node) Len() int {
	n.RLock()
	defer n.RUnlock()
	return len(n.data)
}

func (n *Node) NodeName() string {
	return n.host
}

type NodePool struct {
	nodes map[string]*Node
}

func NewPool(number int) *NodePool {
	nodes := make(map[string]*Node, 0)
	for i := 0; i < number; i++ {
		nodes["node"+strconv.Itoa(i)] = New("node" + strconv.Itoa(i))
	}
	return &NodePool{
		nodes: nodes,
	}
}

func (p *NodePool) GetNodes() map[string]*Node {
	return p.nodes
}
func (p *NodePool) GetNode(hostOrIp string) *Node {
	return p.nodes[hostOrIp]
}
