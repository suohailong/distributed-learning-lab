package coordinator

import (
	"math/rand"
	"time"
)

type Gossip interface {
	Start()
	Stop()
	Join(seedNodeAddress string)
	UpdateState(key string, value interface{})
	GetState(key string) (interface{}, bool)
	RegisterEventHandler(handler EventHandler)
}

type EventHandler func(event Event)

type Event struct {
	Type   string
	Key    string
	Value  interface{}
	NodeID string
}
type Node struct {
	Id      string
	Address string
	//TODO: 这里会一直增长吗
	Counter   uint64
	heartTime int64
}

func (n *Node) Send(message interface{}) {

}

type gossip struct {
	// Implementation details
	currentNode uint32
	// node 关系列表
	memberlist map[uint32]*Node
	// node的排序列表
	sortNodeList []*Node
	// 随机发送信息的node个数
	randomSet int

	// heartThreshold
}

func (g *gossip) getRandomNode() []*Node {
	nodes := make([]*Node, 0)
	for i := 0; i < g.randomSet; i++ {
		rand.Seed(time.Now().UnixNano())
		min := 0
		max := len(g.sortNodeList)
		index := rand.Intn(max-min+1) + min
		nodes = append(nodes, g.sortNodeList[index])
	}

	return nodes
}

func (g *gossip) Start() {
	// 更新heart, 发送本地信息
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			node := g.memberlist[g.currentNode]
			node.heartTime = time.Now().Unix()
			node.Counter++

			nodes := g.getRandomNode()
			for _, n := range nodes {
				n.Send(g.memberlist)
			}
		}
	}()
}

func (g *gossip) Stop() {
	panic("not implemented") // TODO: Implement
}

func (g *gossip) Join(seedNodeAddress string) {
	panic("not implemented") // TODO: Implement
}

func (g *gossip) UpdateState(members interface{}) {
	memberlist := members.(map[uint32]*Node)

}

func (g *gossip) GetState(key string) (interface{}, bool) {
	panic("not implemented") // TODO: Implement
}

func (g *gossip) RegisterEventHandler(handler EventHandler) {
	panic("not implemented") // TODO: Implement
}
