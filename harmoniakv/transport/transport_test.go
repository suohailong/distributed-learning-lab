package transport

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type transTestSuite struct {
	suite.Suite
}

func (g *transTestSuite) SetupTest() {
}

func (g *transTestSuite) TearDownTest() {
}

func TestGossip(t *testing.T) {
	suite.Run(t, new(transTestSuite))
}

func TestTransport(t *testing.T) {

	nodes := map[string]struct{}{
		"1": struct{}{},
		"2": struct{}{},
		"3": struct{}{},
	}

	Send(nodes, []byte{"aa"})

	WaitResp(2)
}
