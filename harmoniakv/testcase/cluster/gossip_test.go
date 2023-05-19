package cluster

import (
	"distributed-learning-lab/harmoniakv/cluster"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var nodes = []*cluster.Node{}

type gossipTestSuite struct {
	suite.Suite
	co cluster.Cluster
}

func (g *gossipTestSuite) SetupTest() {
	for i := 1; i < 11; i++ {
		nodes = append(nodes, cluster.NewNode("node"+strconv.Itoa(i), ""))
	}
}

func (g *gossipTestSuite) TearDownTest() {
	nodes = []*cluster.Node{}
}

func TestGossip(t *testing.T) {
	suite.Run(t, new(gossipTestSuite))
}

func (g *gossipTestSuite) TestGossipStart() {
	node1, node2, node3 := nodes[0], nodes[1], nodes[2]

	co := cluster.New(100, 2, 3, node1)
	co.AddNode(node2)
	co.AddNode(node3)

	tm := time.Now()
	co.Start()
	for time.Since(tm) > 1*time.Second {
		// 检查currentNode是否被更新
		g.Assert().Equal(1, node1.Counter)
		// 检查超时节点及超时节点是否被更新
		node2.HeartTime = time.Now().Add(-1 * time.Minute).Unix()
		g.Assert().Equal(cluster.OFFLINE, node2.State)
		// 检查是否发送了消息, 且发送节点不包含本地节点
		g.Assert().NotEqual(0, node1.SendTime)
		g.Assert().Less(time.Now().Sub(time.Unix(node1.SendTime, 0)), 1*time.Second)
		g.Assert().NotEqual(0, node2.SendTime)
		g.Assert().Less(time.Now().Sub(time.Unix(node2.SendTime, 0)), 1*time.Second)
	}
}

func (g *gossipTestSuite) TestHandleGossipMessage() {
	node1, node2, node3 := nodes[0], nodes[1], nodes[2]
	node1.Counter = 1
	node2.Counter = 2
	node3.Counter = 3

	co := cluster.New(100, 2, 3, node2)
	co.AddNode(node1)
	co.AddNode(node3)

	// 这种场景是 心跳中的node没有本地node更新
	testCase := &cluster.GossipStateMessage{
		Source: node1,
		NodeStates: map[string]*cluster.Node{
			node1.GetID(): cluster.NewNode("node1", ""),
			node2.GetID(): cluster.NewNode("node2", ""),
			node3.GetID(): cluster.NewNode("node3", ""),
		},
	}
	co.HandleGossipMessage(testCase)
	// // 检查是否发送了消息, 且发送节点不包含本地节点
	// g.Assert().NotEqual(0, node1.SendTime)
	// g.Assert().Less(time.Now().Sub(time.Unix(node1.SendTime, 0)), 1*time.Second)
	// g.Assert().NotEqual(0, node3.SendTime)
	// g.Assert().Less(time.Now().Sub(time.Unix(node3.SendTime, 0)), 1*time.Second)

	// 检查currentNode是否被更新
	g.Assert().Equal(uint64(1), node1.Counter)
	g.Assert().Equal(uint64(2), node2.Counter)
	g.Assert().Equal(uint64(3), node3.Counter)

	// 这种场景是 心跳中的node比本地node更新
	hertNode1 := cluster.NewNode("node1", "")
	hertNode2 := cluster.NewNode("node2", "")
	hertNode3 := cluster.NewNode("node3", "")
	hertNode1.Counter = 3
	hertNode2.Counter = 4
	testCase = &cluster.GossipStateMessage{
		Source: node1,
		NodeStates: map[string]*cluster.Node{
			node1.GetID(): hertNode1,
			node2.GetID(): hertNode2,
			node3.GetID(): hertNode3,
		},
	}
	co.HandleGossipMessage(testCase)
	// 检查currentNode是否被更新
	g.Assert().Equal(uint64(3), node1.Counter)
	g.Assert().Equal(uint64(4), node2.Counter)

	// 这种场景是 心跳中的node比本地node多
	hertNode1 = cluster.NewNode("node1", "")
	hertNode2 = cluster.NewNode("node2", "")
	hertNode3 = cluster.NewNode("node3", "")
	hertNode4 := cluster.NewNode("node4", "")
	testCase = &cluster.GossipStateMessage{
		Source: node1,
		NodeStates: map[string]*cluster.Node{
			node1.GetID():     hertNode1,
			node2.GetID():     hertNode2,
			node3.GetID():     hertNode3,
			hertNode4.GetID(): hertNode4,
		},
	}
	co.HandleGossipMessage(testCase)
	// 检查currentNode是否被更新
	g.Assert().Equal(4, co.NodeLen())
}
