package coordinator

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
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

func (n *Node) Send(message interface{}) {
	fmt.Println("发送信息", message)
}
func (n *Node) GetID() string {
	return n.ID
}

type HeartBeat struct {
	Source     *Node
	Memberlist map[string]*Node
}

func (c *coordinator) getRandomNode() []*Node {
	nodes := make([]*Node, 0)
	rand.Seed(time.Now().UnixNano())
	useIndex := make(map[string]struct{}, 0)
	for i := 0; i < c.randomNode; i++ {
		index := rand.Intn(len(c.pnodes))
		nodeId := c.pnodes[index].GetID()
		if _, ok := useIndex[nodeId]; ok {
			continue
		}
		useIndex[nodeId] = struct{}{}
		nodes = append(nodes, c.pnodes[index])
	}
	return nodes
}

func (c *coordinator) runGossip() {

	// 更新heartBeat
	node := c.currentNode
	node.HeartTime = time.Now().Unix()
	node.Counter++
	log.Debugf("update node %s, heartCount: %d\n", node.GetID(), node.Counter)

	// 检查成员列表中node的状态
	scannedNodes := make(map[string]struct{}, 0)
	for _, n := range c.pnodes {
		if _, ok := scannedNodes[n.GetID()]; ok {
			continue
		}
		scannedNodes[n.GetID()] = struct{}{}
		t := time.Unix(n.HeartTime, 0)
		// 如果时间超过1m没有心跳，则认为机器已经失效
		if time.Since(t) > 1*time.Minute {
			n.State = OFFLINE
			log.Debugf("update node %s, state: %d\n", n.GetID(), n.State)
		}
	}

	// 随机发送本地消息
	nodes := c.getRandomNode()
	for _, n := range nodes {
		n.Send(&HeartBeat{})
		log.Debugf("send heart beat to node %s \n", n.GetID())
	}
}

func (c *coordinator) Start() {
	// 更新heart, 发送本地信息
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
		}
	}()
}

// 处理心跳
func (c *coordinator) HandleHeartBeat(heartBeeat *HeartBeat) {
	// 随机发送本地消息
	nodes := g.getRandomNode()
	for _, n := range nodes {
		fmt.Printf("send heart beat to %s\n", n.GetID())
		n.Send(heartBeeat)
	}
	// 更新本地
	for _, n := range heartBeeat.Memberlist {
		index := sort.Search(len(c.pnodes), func(i int) bool {
			return c.pnodes[i].GetID() >= n.GetID()
		})
		me := c.pnodes[index]
		if index < len(c.pnodes) {
			// 如果本地的counter小于远程的counter，则更新本地的信息
			if me.Counter <= n.Counter {
				me.HeartTime = time.Now().Unix()
				me.Counter = n.Counter
			}
		} else {
			n.HeartTime = time.Now().Unix()
			c.AddNode(n)
		}

	}

}
