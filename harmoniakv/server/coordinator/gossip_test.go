package coordinator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type gossipTestSuite struct {
	suite.Suite
	co *coordinator
}

func (g *gossipTestSuite) SetupTest() {
	co := &coordinator{
		// uint32 代表虚拟hash值
		nodes:        make([]*Node, 0),
		nodeMap:      make(map[string]int),
		virtualNodes: make([]uint32, 0),
		vnodesMap:    make(map[uint32]int),
		vnodeLen:     100,
		pnodeLen:     0,
		randomNode:   2,
	}
	g.co = co

}

func TestGossip(t *testing.T) {
	suite.Run(t, new(gossipTestSuite))
}

func (g *gossipTestSuite) TestGetRandom() {
	var (
		node1 = &Node{
			ID:    "node1",
			State: ONLINE,
		}
		node2 = &Node{
			ID:    "node2",
			State: ONLINE,
		}
		node3 = &Node{
			ID:    "node3",
			State: ONLINE,
		}
	)
	g.co.randomNode = 2
	g.co.currentNode = node1
	g.co.AddNode(node1)
	g.co.AddNode(node2)
	g.co.AddNode(node3)

	nodes := g.co.getRandomNode()
	g.Assert().Equal(2, len(nodes))
	for _, n := range nodes {
		g.Assert().Contains([]string{"node2", "node3"}, n.ID)
	}
}
