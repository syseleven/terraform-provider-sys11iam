package rest

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

func (suite *ClientTestSuite) TestURLFix() {
	client := NewClient("http://localhost")
	suite.Equal("http://localhost", client.url)

	client = NewClient("http://localhost/")
	suite.Equal("http://localhost", client.url)
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
