package keycloak

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	responses "github.com/syseleven/terraform-provider-sys11iam/internal/http-responses"
)

type RestClientKeystoneTestSuite struct {
	suite.Suite
}

func (suite *RestClientKeystoneTestSuite) TestLoginSuccess() {
	sampleResponse := `{
		"access_token": "token"
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/").
			WithBody([]byte(`client_id=pytest&client_secret=YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK&grant_type=password&password=pytest&scope=pytest&username=pytest`)).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithClientConfig("pytest", "YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK", "pytest", "pytest", "pytest")

	id, err := client.Login()
	suite.NoError(err)
	suite.Equal(id, "token")
	mockServer.HasExpectedRequests()
}

func (suite *RestClientKeystoneTestSuite) TestLoginError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/").
			WithBody([]byte(`client_id=pytest&client_secret=YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK&grant_type=password&password=pytest&scope=pytest&username=pytest`)).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithClientConfig("pytest", "YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK", "pytest", "pytest", "pytest")

	id, err := client.Login()
	suite.Error(err) //TODO: check error message
	suite.Equal(id, "")
	mockServer.HasExpectedRequests()
}

func TestRestClientKeystoneTestSuite(t *testing.T) {
	suite.Run(t, new(RestClientKeystoneTestSuite))
}
