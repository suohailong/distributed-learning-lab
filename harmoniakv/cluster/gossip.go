package cluster

import (
	"math/rand"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

// gossip负责传播的信息包括, 成员信息 token metadata   schema version information
// 每个节点独自决定other node的up and down 不通过gossip 传播

type NodeId string

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
	SendTime  int64
}

func NewNode(id string, address string) *Node {
	return &Node{
		ID:        id,
		Address:   address,
		Counter:   0,
		State:     ONLINE,
		HeartTime: 0,
		SendTime:  0,
	}
}

func (n *Node) Send(message interface{}) {
	log.Infof("id:[%s], send heartbeat: <%s>", n.GetID(), message)
}
func (n *Node) GetID() string {
	return n.ID
}

type GossipStateMessage struct {
	Source     *Node
	NodeStates map[string]*Node
}

func (c *defaultCluster) getRandomNode() []*Node {
	nodes := make([]*Node, 0)
	rand.Seed(time.Now().UnixNano())
	useIndex := make(map[string]struct{}, 0)
	for {
		index := rand.Intn(len(c.pnodes))
		nodeId := c.pnodes[index].GetID()
		if _, ok := useIndex[nodeId]; ok || c.currentNode.GetID() == nodeId {
			continue
		}
		useIndex[nodeId] = struct{}{}
		nodes = append(nodes, c.pnodes[index])
		if len(nodes) == c.randomNode {
			break
		}
	}
	return nodes
}

func (c *defaultCluster) runGossip() {
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
		n.Send(&GossipStateMessage{})
		n.SendTime = time.Now().Unix()
		log.Debugf("send heart beat to node %s \n", n.GetID())
	}
}

func (c *defaultCluster) Start() {
	// 更新heart, 发送本地信息
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			c.runGossip()
		}
	}()
}

// 处理心跳
func (c *defaultCluster) HandleGossipMessage(gossipMessage *GossipStateMessage) {
	// nodes := c.getRandomNode()
	// for _, n := range nodes {
	// 	n.Send(heartBeeat)
	// 	n.SendTime = time.Now().Unix()
	// }
	// 更新本地
	diff := make(map[string]*Node, 0)
	for _, n := range gossipMessage.NodeStates {
		index := sort.Search(len(c.pnodes), func(i int) bool {
			return c.pnodes[i].GetID() >= n.GetID()
		})

		if index < len(c.pnodes) {
			me := c.pnodes[index]
			// 本地存在的,直接更新对应的counter
			if me.Counter <= n.Counter {
				me.HeartTime = time.Now().Unix()
				me.Counter = n.Counter
			} else {
				diff[me.GetID()] = me
			}
		} else {
			//本地不存在,添加记录到本地
			n.HeartTime = time.Now().Unix()
			c.AddNode(n)
		}
	}

	for _, n := range c.pnodes {
		if _, ok := gossipMessage.NodeStates[n.GetID()]; !ok {
			diff[n.GetID()] = n
		}
	}
	log.Printf("diff: %v\n", diff)

	source := gossipMessage.Source
	source.Send(&GossipStateMessage{
		Source:     c.currentNode,
		NodeStates: diff,
	})
}
