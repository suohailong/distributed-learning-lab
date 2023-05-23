package server

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	server *server
}

func (suite *ServerTestSuite) SetupTest() {
	suite.server = &server{}
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) TestServerWrite() {
	
}
