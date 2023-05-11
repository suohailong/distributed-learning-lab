package coordinator

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	ONLINE  = 1 // 0
	OFFLINE = -1
)

type Gossip interface {
	Start()
	UpdateState(key string, value interface{})
}

type Node struct {
	ID      string
	Address string
	//TODO: 这里会一直增长吗
	Counter   uint64
	HeartTime int64
	State     int8
}

func (n *Node) Send(message interface{}) {}
func (n *Node) GetID() string {
	return n.ID
}

type HeartBeat struct {
	Source     *Node
	Memberlist map[string]*Node
}

func (g *coordinator) getRandomNode() []*Node {
	nodes := make([]*Node, 0)
	rand.Seed(time.Now().UnixNano())
	useIndex := make(map[string]struct{}, 0)
	for i := 0; i < g.randomSet; i++ {
		min := 0
		max := len(g.nodes) - 1
		index := rand.Intn(max-min+1) + min
		nodeID := g.nodesMap[g.nodes[index]].GetID()
		for _, ok := useIndex[nodeID]; ok || nodeID == g.currentNode.GetID(); _, ok = useIndex[nodeID] {
			index = rand.Intn(max-min+1) + min
			nodeID = g.nodesMap[g.nodes[index]].GetID()
		}
		useIndex[nodeID] = struct{}{}
		nodes = append(nodes, g.nodesMap[g.nodes[index]])
	}
	return nodes
}

func (g *coordinator) Start() {
	// 更新heart, 发送本地信息
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			// 更新heartBeat
			node := g.currentNode
			node.HeartTime = time.Now().Unix()
			node.Counter++

			// 检查成员列表中node的状态
			scannedNodes := make(map[string]struct{}, 0)
			for _, n := range g.nodesMap {
				if _, ok := scannedNodes[n.GetID()]; ok {
					continue
				}
				scannedNodes[n.GetID()] = struct{}{}
				t := time.Unix(n.HeartTime, 0)
				// 如果时间超过1m没有心跳，则认为机器已经失效
				if time.Since(t) > 1*time.Minute {
					n.State = OFFLINE
				}
			}

			// 随机发送本地消息
			nodes := g.getRandomNode()
			for _, n := range nodes {
				n.Send(&HeartBeat{})
			}
		}
	}()
}

// 处理心跳
func (g *coordinator) HandleHeartBeat(heartBeeat *HeartBeat) {
	// 随机发送本地消息
	nodes := g.getRandomNode()
	for _, n := range nodes {
		fmt.Printf("send heart beat to %d\n", n.GetID())
		n.Send(heartBeeat)
	}
	// 更新本地
	for _, n := range heartBeeat.Memberlist {
		if me, ok := g.[n.Id]; ok {
			// 如果本地的counter小于远程的counter，则更新本地的信息
			if me.Counter <= n.Counter {
				n.HeartTime = time.Now().Unix()
				g.memberlist[n.Id] = n
			}
		} else {
			// 如果本地没有，则添加
			n.HeartTime = time.Now().Unix()
			g.memberlist[n.Id] = n
		}
	}

}
