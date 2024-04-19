package consistent_hash

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConsistentHashTestSuite struct {
	suite.Suite
	ch *defaultConsistentHash
}

func (suite *ConsistentHashTestSuite) SetupTest() {
	suite.ch = &defaultConsistentHash{
		nodes:        []uint32{},
		nodesMap:     make(map[uint32]string),
		virtualNodes: 100,
	}
}

func (suite *ConsistentHashTestSuite) TestAddNode() {
	suite.ch.AddNode("node1")
	suite.ch.AddNode("node2")

	// assert.Len(suite.T(), suite.ch.Nodes(), 2, "Expected 2 nodes")
	assert.Equal(suite.T(), 2, suite.ch.nodeLen())
}

func (suite *ConsistentHashTestSuite) TestRemoveNode() {
	suite.ch.AddNode("node1")
	suite.ch.AddNode("node2")
	suite.ch.RemoveNode("node1")

	assert.Equal(suite.T(), 1, suite.ch.nodeLen())
}

func (suite *ConsistentHashTestSuite) TestGetNodes() {
	suite.ch.AddNode("node1")
	suite.ch.AddNode("node2")

	key := "test-key"
	nodes := suite.ch.GetNodes(key, 1)

	require.NotEmpty(suite.T(), nodes, "Expected at least 1node")
}

func (suite *ConsistentHashTestSuite) concurrencyOperation(fn func(nodeID string)) {
	wg := sync.WaitGroup{}
	wg.Add(5)
	for i := 1; i <= 5; i++ {
		go func(nodeID string) {
			defer wg.Done()
			// suite.ch.AddNode(nodeID)
			fn(nodeID)
		}(strconv.Itoa(i))
	}
	wg.Wait()
}

func (suite *ConsistentHashTestSuite) TestConcurrencySafety() {
	fmt.Println(suite.ch.nodesMap)
	suite.concurrencyOperation(func(nodeID string) {
		suite.ch.AddNode(nodeID)
	})
	assert.Equal(suite.T(), 5, suite.ch.nodeLen())

	suite.concurrencyOperation(func(nodeID string) {
		suite.ch.RemoveNode(nodeID)
	})
	assert.Equal(suite.T(), 0, suite.ch.nodeLen())
}

func TestConsistentHashTestSuite(t *testing.T) {
	suite.Run(t, new(ConsistentHashTestSuite))
}
