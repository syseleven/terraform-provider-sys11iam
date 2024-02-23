package iam

import (
	"net/http"
	"testing"

	responses "gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/http-responses"

	"github.com/stretchr/testify/suite"
)

type RestClientIAMTestSuite struct {
	suite.Suite
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationSuccess() {
	sampleResponse := `{
		"name": "sample-org",
		"description": "sample-org",
		"id": "1",
		"created_at": "date",
		"updated_at": "date",
		"tags": ["sample-tag"],
		"is_active": true,
		"organization_id": "1"
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v1/orgs/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.GetOrganization("1")
	suite.NoError(err)
	iamOrg := IAMOrganization(IAMOrganization{ID: "1", OrganizationId: "1", Name: "sample-org", Description: "sample-org", Tags: []string{"sample-tag"}, CreatedAt: "date", IsActive: true, UpdatedAt: "date"})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationSuccess() {
	sampleResponse := `{
		"name": "sample-org",
		"description": "sample-org",
		"id": "1",
		"created_at": "date",
		"updated_at": "date",
		"tags": ["sample-tag"],
		"is_active": true,
		"organization_id": "1"
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v1/orgs").
			WithBody([]byte(`{"description":"sample-org","name":"sample-org","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateOrganization("sample-org", "sample-org", []string{"sample-tag"})
	suite.NoError(err)
	iamOrg := IAMOrganization(IAMOrganization{ID: "1", OrganizationId: "1", Name: "sample-org", Description: "sample-org", Tags: []string{"sample-tag"}, CreatedAt: "date", IsActive: true, UpdatedAt: "date"})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v1/orgs").
			WithBody([]byte(`{"description":"sample-org","name":"sample-org","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateOrganization("sample-org", "sample-org", []string{"sample-tag"})
	suite.Error(err) //TODO: check error message
	iamOrg := IAMOrganization(IAMOrganization{ID: "", OrganizationId: "", Name: "", Description: "", Tags: []string(nil), CreatedAt: "", IsActive: false, UpdatedAt: ""})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationSuccess() {
	sampleResponse := `{
		"name": "sample-org",
		"description": "sample-org",
		"id": "1",
		"created_at": "date",
		"updated_at": "date",
		"tags": ["sample-tag"],
		"is_active": true,
		"organization_id": "1"
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPut, "/v1/orgs/1").
			WithBody([]byte(`{"description":"sample-org","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.UpdateOrganization("1", "sample-org", []string{"sample-tag"})
	suite.NoError(err)
	iamOrg := IAMOrganization(IAMOrganization{ID: "1", OrganizationId: "1", Name: "sample-org", Description: "sample-org", Tags: []string{"sample-tag"}, CreatedAt: "date", IsActive: true, UpdatedAt: "date"})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPut, "/v1/orgs/1").
			WithBody([]byte(`{"description":"sample-org","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.UpdateOrganization("1", "sample-org", []string{"sample-tag"})
	suite.Error(err) //TODO: check error message
	iamOrg := IAMOrganization(IAMOrganization{ID: "", OrganizationId: "", Name: "", Description: "", Tags: []string(nil), CreatedAt: "", IsActive: false, UpdatedAt: ""})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationSuccess() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v1/orgs/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusNoContent).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteOrganization("1")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v1/orgs/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusBadRequest).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteOrganization("1")
	suite.Error(err)
	mockServer.HasExpectedRequests()
}

func TestRestClientIAMTestSuite(t *testing.T) {
	suite.Run(t, new(RestClientIAMTestSuite))
}
