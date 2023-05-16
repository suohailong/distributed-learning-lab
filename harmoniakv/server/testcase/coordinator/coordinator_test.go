package coordinator

import (
	"distributed-learning-lab/harmoniakv/server/coordinator"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CoordinatorTestSuite struct {
	suite.Suite
	c coordinator.Coordinator
}

func (suite *CoordinatorTestSuite) SetupTest() {
	fmt.Println("SetupTest", suite.T().Name())
	suite.c = coordinator.New(100, 5, 3)
}

func (suite *CoordinatorTestSuite) TearDownTest() {
	fmt.Println("DownTest", suite.T().Name())
	suite.c = nil
}

func (suite *CoordinatorTestSuite) TestAddNode() {
	node := &coordinator.Node{ID: "node1"}
	suite.c.AddNode(node)

	suite.Assert().Equal(1, suite.c.NodeLen())
	suite.Assert().Equal(node, suite.c.GetNode("some_key"))
}

func (suite *CoordinatorTestSuite) TestRemoveNode() {
	node := &coordinator.Node{ID: "node1"}
	suite.c.AddNode(node)
	suite.c.RemoveNode(node)

	suite.Assert().Equal(0, suite.c.NodeLen())
	suite.Assert().Nil(suite.c.GetNode("some_key"))
}

func (suite *CoordinatorTestSuite) TestGetNode() {
	node1 := &coordinator.Node{ID: "node1"}
	node2 := &coordinator.Node{ID: "node2"}
	suite.c.AddNode(node1)
	suite.c.AddNode(node2)

	suite.Assert().Equal(2, suite.c.NodeLen())
	suite.Assert().Contains([]*coordinator.Node{node1, node2}, suite.c.GetNode("some_key"))
}

func (suite *CoordinatorTestSuite) TestGetReplicas() {
	for i := 0; i < 10; i++ {
		node := &coordinator.Node{ID: "node" + strconv.Itoa(i), State: coordinator.ONLINE}
		suite.c.AddNode(node)
	}

	nodes1 := suite.c.GetReplicas("some_key", 3)
	nodes2 := suite.c.GetReplicas("some_key", 3)

	suite.Assert().Equal(len(nodes1), len(nodes2))
	for _, n := range nodes2 {
		suite.Assert().Contains(nodes1, n)
	}
	suite.T().Log(nodes1, nodes2)

	suite.c.RemoveNode(&coordinator.Node{ID: "node2"})
	nodes3 := suite.c.GetReplicas("some_key", 3)
	suite.T().Log(nodes1, nodes3)

}

func (suite *CoordinatorTestSuite) TestGossip() {

}

func TestCoordinatorTestSuite(t *testing.T) {
	suite.Run(t, new(CoordinatorTestSuite))
}
