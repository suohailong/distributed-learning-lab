package transport

import (
	"distributed-learning-lab/harmoniakv/config"
	"distributed-learning-lab/harmoniakv/node"
	"distributed-learning-lab/harmoniakv/node/version"
	"testing"

	"github.com/cch123/supermonkey"
	"github.com/panjf2000/ants/v2"
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

// GOARCH=amd64 go test -test.run="TestTran/TestTransport" -v -ldflags="-s=false" -gcflags="-l"  ./harmoniakv/transport/
func (g *transTestSuite) TestTransport() {
	// logger, hook := test.NewNullLogger() // logger.Error("Helloerror")

	// assert.Equal(t, 1, len(hook.Entries))
	// assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	// assert.Equal(t, "Helloerror", hook.LastEntry().Message)

	// hook.Reset()
	// assert.Nil(t, hook.LastEntry())
	node1 := &node.Node{ID: "node1"}
	node2 := &node.Node{ID: "node2"}
	node3 := &node.Node{ID: "node3"}

	//这里mac m1 需要以下设置GOARCH=amd64 go test -test.run="TestTran/TestTransport" -v -ldflags="-s=false" -gcflags="-l"  ./harmoniakv/transport/
	p := supermonkey.Patch(config.LocalId, func() string {
		logrus.Debug("patch local id")
		return "node1"
	})
	defer p.Unpatch()
	pool, _ := ants.NewPool(10)
	transfer := &defaultTransporter{
		ret:  make(chan interface{}),
		pool: pool,
	}
	cmd := &node.KvCommand{
		Command: node.PUT,
		Key:     []byte("key"),
		Value: version.Value{
			Key:   []byte("key"),
			Value: []byte("value"),
			VersionVector: &version.Vector{
				Versions: map[string]uint64{
					"node1": 1,
					"node2": 1,
					"node3": 1,
				},
			},
		},
	}
	ret, err := transfer.send(node1, cmd)
	collectResp(ret, err)
	transfer.asend(node2, cmd)
	transfer.asend(node3, cmd)
	transfer.waitResp(2)
}
