package consistenthash_test

import (
	hash "distributed-learning-lab/go-consistent-hash"
	"distributed-learning-lab/go-consistent-hash/test/node"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConsistentHashTestSuite struct {
	suite.Suite
	ch hash.ConsistentHash
}

func (suite *ConsistentHashTestSuite) SetupTest() {
	suite.ch = hash.New(100)
}

// 测试节点分布均匀性
func (suite *ConsistentHashTestSuite) TestUniformDistribution() {
	// Implement a test to check if the keys are uniformly distributed among nodes
	nodePool := node.NewPool(3)
	nodes := nodePool.GetNodes()

	// 300是最佳
	cluster := hash.New(300)
	for host := range nodes {
		cluster.AddNode(host)
	}
	for i := 0; i < 10000000; i++ {
		key := "key" + strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		host := cluster.GetNode(key)
		node := nodePool.GetNode(host)
		node.Set(key, value)
	}

	for _, node := range nodes {
		fmt.Println(node.NodeName(), node.Len())
	}
}

func TestConsistentHashTestSuite(t *testing.T) {
	suite.Run(t, new(ConsistentHashTestSuite))
}
