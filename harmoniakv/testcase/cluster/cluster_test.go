package cluster

import (
	"distributed-learning-lab/harmoniakv/cluster"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClusterTestSuite struct {
	suite.Suite
	c cluster.Cluster
}

func (suite *ClusterTestSuite) SetupTest() {
	fmt.Println("SetupTest", suite.T().Name())
	suite.c = cluster.New(100, 5, 3, &cluster.Node{})
}

func (suite *ClusterTestSuite) TearDownTest() {
	fmt.Println("DownTest", suite.T().Name())
	suite.c = nil
}

func (suite *ClusterTestSuite) TestAddNode() {
	node := &cluster.Node{ID: "node1"}
	suite.c.AddNode(node)

	suite.Assert().Equal(1, suite.c.NodeLen())
	suite.Assert().Equal(node, suite.c.GetNode("some_key"))
}

func (suite *ClusterTestSuite) TestRemoveNode() {
	node := &cluster.Node{ID: "node1"}
	suite.c.AddNode(node)
	suite.c.RemoveNode(node)

	suite.Assert().Equal(0, suite.c.NodeLen())
	suite.Assert().Nil(suite.c.GetNode("some_key"))
}

func (suite *ClusterTestSuite) TestGetNode() {
	node1 := &cluster.Node{ID: "node1"}
	node2 := &cluster.Node{ID: "node2"}
	suite.c.AddNode(node1)
	suite.c.AddNode(node2)

	suite.Assert().Equal(2, suite.c.NodeLen())
	suite.Assert().Contains([]*cluster.Node{node1, node2}, suite.c.GetNode("some_key"))
}

func (suite *ClusterTestSuite) TestGetReplicas() {
	for i := 0; i < 10; i++ {
		node := &cluster.Node{ID: "node" + strconv.Itoa(i), State: cluster.ONLINE}
		suite.c.AddNode(node)
	}

	nodes1 := suite.c.GetReplicas("some_key", 3)
	nodes2 := suite.c.GetReplicas("some_key", 3)

	suite.Assert().Equal(len(nodes1), len(nodes2))
	for _, n := range nodes2 {
		suite.Assert().Contains(nodes1, n)
	}
	suite.T().Log(nodes1, nodes2)

	suite.c.RemoveNode(&cluster.Node{ID: "node2"})
	nodes3 := suite.c.GetReplicas("some_key", 3)
	suite.T().Log(nodes1, nodes3)

}

func (suite *ClusterTestSuite) TestGossip() {

}

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, new(ClusterTestSuite))
}
