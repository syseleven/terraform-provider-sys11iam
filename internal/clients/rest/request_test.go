package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RequestTestSuite struct {
	suite.Suite
}

func (suite *RequestTestSuite) TestDuplicateAuthenticationHeaders() {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Len(r.Header["Authorization"], 1)
		suite.Equal("Basic bXktdXNlcm5hbWU6bXktcGFzc3dvcmQ=", r.Header.Get("Authorization"))
	}))
	defer mockServer.Close()

	c := NewClient(mockServer.URL).
		WithLogin("my-username", "my-password")
	ctx := Context{
		"some-request-id",
		"some-auth-token",
	}
	_, err := c.UseContext(&ctx).NewRequest(http.MethodGet, "/foo").Do()
	suite.NoError(err)
}

func TestRequestTestSuite(t *testing.T) {
	suite.Run(t, new(RequestTestSuite))
}
