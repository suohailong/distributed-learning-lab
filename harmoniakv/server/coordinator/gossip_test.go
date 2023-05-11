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
	co := &coordinator{}
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
	g.co.randomSet = 2
	g.co.currentNode = node1
	g.co.AddNode(node1)
	g.co.AddNode(node2)
	g.co.AddNode(node3)

	nodes := g.co.getRandomNode()
	g.Assert().Equal(2, len(nodes))
	for _, n := range nodes {
		g.Assert().Contains([]string{"node1", "node2"}, n.ID)
	}

}
