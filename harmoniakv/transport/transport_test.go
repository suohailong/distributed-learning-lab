package transport

import (
	"distributed-learning-lab/harmoniakv/config"
	"distributed-learning-lab/harmoniakv/node"
	"testing"

	"github.com/cch123/supermonkey"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type transTestSuite struct {
	suite.Suite
}

func (g *transTestSuite) SetupTest() {
	logrus.SetLevel(logrus.DebugLevel)
}

func (g *transTestSuite) TearDownTest() {
}

func TestTran(t *testing.T) {
	suite.Run(t, new(transTestSuite))
}

func (g *transTestSuite) TestTransport() {
	// logger, hook := test.NewNullLogger()
	// logger.Error("Helloerror")

	// assert.Equal(t, 1, len(hook.Entries))
	// assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	// assert.Equal(t, "Helloerror", hook.LastEntry().Message)

	// hook.Reset()
	// assert.Nil(t, hook.LastEntry())
	node1 := &node.Node{ID: "node1"}
	node2 := &node.Node{ID: "node2"}
	node3 := &node.Node{ID: "node3"}

	nodes := []*node.Node{
		node1,
		node2,
		node3,
	}
	//这里mac m1 需要将 GOARCH=amd64
	p := supermonkey.Patch(config.LocalId, func() string {
		return "node1"
	})
	defer p.Unpatch()

	transfer := &defaultTransporter{
		ret: make(chan struct{}),
	}

	transfer.send(nodes, []byte("aa"))

	transfer.waitResp(2)
}
