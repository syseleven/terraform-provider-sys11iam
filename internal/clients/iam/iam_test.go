package iam

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	responses "github.com/syseleven/terraform-provider-sys11iam/internal/http-responses"
)

type RestClientIAMTestSuite struct {
	suite.Suite
}

var examplestring = "1"
var exampleSlicedstring = []string{"can_do"}

var exampleIAMOrganizationCompanyInfo = IAMOrganizationCompanyInfo{Street: "street", StreetNumber: "1", ZipCode: "12345"}
var exampleIAMOrganization = IAMOrganization{ID: "1", Name: "sample-org", Description: "sample-org", Tags: []string{"sample-tag"}, CreatedAt: "date", IsActive: true, UpdatedAt: "date", CompanyInfo: exampleIAMOrganizationCompanyInfo}
var exampleIAMOrganizationUser = IAMOrganisationUser{ID: "1", Email: "test@syseleven.net"}
var exampleIAMOrganizationMembership = IAMOrganizationMembership{Organisation: exampleIAMOrganization, User: exampleIAMOrganizationUser, Affiliation: "member", MembershipType: "service_account", Permissions: []string{"can_do"}}
var exampleIAMOrganizationContact = IAMOrganizationContact{}
var exampleIAMOrganizationTeam = IAMOrganizationTeam{}

var exampleIAMOrganizationInvitation = IAMOrganizationInvitation{ID: "1", Email: "test@syseleven.net"}

var exampleIAMOrganizationTeamMembership = IAMOrganizationTeamMembership{Organisation: exampleIAMOrganization}

var exampleIAMProjectTeamMembership = IAMProjectTeamMembership{}
var exampleIAMProjectTeamPermissions = IAMProjectTeamPermissions{TeamPermissions: []string{"can_do"}, OrganizationId: "1", ProjectId: "1", TeamId: "1", AddedPermissions: []string{"can_do"}, RemovedPermissions: []string{"wont_do"}, UpdatedPermissions: []string{"will_do"}}

var exampleIAMProjectS3User = IAMProjectS3User{}

var exampleIAMProjectS3KeyUser = IAMProjectS3UserKey{}

func (suite *RestClientIAMTestSuite) TestGetOrganizationSuccess() {
	method := "GET"
	url := "/v2/orgs/1"
	status := http.StatusOK
	expected := IAMOrganization(exampleIAMOrganization)
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganization("1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationSuccess() {
	method := "POST"
	url := "/v2/orgs"
	status := http.StatusCreated
	expected := IAMOrganization(exampleIAMOrganization)
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateOrganization(exampleIAMOrganization)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs").
			WithBody([]byte(`{"company_info":{"accepted_tos":true,"city":"testcity","company_name":"testcompany","country":"testland","phone":"+49123456789","preferred_billing_method":"SEPA","street":"teststreet","street_number":"1","vat_id":"42069","zip_code":"12345"},"description":"sample-org","name":"sample-org","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	iAMOrganization := IAMOrganization{
		Name:        "sample-org",
		Description: "sample-org",
		Tags:        []string{"sample-tag"},
		CompanyInfo: IAMOrganizationCompanyInfo{
			Street:                 "teststreet",
			StreetNumber:           "1",
			ZipCode:                "12345",
			Country:                "testland",
			Phone:                  "+49123456789",
			City:                   "testcity",
			VatID:                  "42069",
			PreferredBillingMethod: "SEPA",
			AcceptedTos:            true,
			CompanyName:            "testcompany",
		},
	}
	id, err := client.CreateOrganization(iAMOrganization)
	suite.Error(err) //TODO: check error message
	iamOrg := IAMOrganization(IAMOrganization{ID: "", Name: "", Description: "", Tags: []string(nil), CreatedAt: "", IsActive: false, UpdatedAt: ""})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationSuccess() {
	method := "PUT"
	url := "/v2/orgs/1"
	status := http.StatusOK
	expected := IAMOrganization(exampleIAMOrganization)
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateOrganization("1", exampleIAMOrganization)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPut, "/v2/orgs/1").
			WithBody([]byte(`{"description":"sample-org","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	iAMOrganization := IAMOrganization{
		Name:        "sample-org",
		Description: "sample-org",
		Tags:        []string{"sample-tag"},
		CompanyInfo: IAMOrganizationCompanyInfo{
			Street:                 "teststreet",
			StreetNumber:           "1",
			ZipCode:                "12345",
			Country:                "testland",
			Phone:                  "+49123456789",
			City:                   "testcity",
			VatID:                  "42069",
			PreferredBillingMethod: "SEPA",
			AcceptedTos:            true,
			CompanyName:            "testcompany",
		},
	}

	id, err := client.UpdateOrganization("1", iAMOrganization)
	suite.Error(err) //TODO: check error message
	iamOrg := IAMOrganization(IAMOrganization{ID: "", Name: "", Description: "", Tags: []string(nil), CreatedAt: "", IsActive: false, UpdatedAt: ""})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationSuccess() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1").
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
		responses.Expect(http.MethodDelete, "/v2/orgs/1").
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

func (suite *RestClientIAMTestSuite) TestGetProjectSuccess() {
	sampleResponse := `{
		"name": "sample-project",
		"description": "sample-project",
		"id": "1",
		"tags": ["sample-tag"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v2/orgs/1/projects/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.GetProject("1", "1")
	suite.NoError(err)
	iamProject := IAMProject(IAMProject{ID: "1", Name: "sample-project", Description: "sample-project", Tags: []string{"sample-tag"}})
	suite.Equal(id, iamProject)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectSuccess() {
	sampleResponse := `{
		"name": "sample-project",
		"description": "sample-project",
		"id": "1",
		"tags": ["sample-tag"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/projects").
			WithBody([]byte(`{"description":"sample-project","name":"sample-project","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateProject("1", "sample-project", "sample-project", []string{"sample-tag"})
	suite.NoError(err)
	iamProject := IAMProject(IAMProject{ID: "1", Name: "sample-project", Description: "sample-project", Tags: []string{"sample-tag"}})
	suite.Equal(id, iamProject)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/projects").
			WithBody([]byte(`{"description":"sample-project","name":"sample-project","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateProject("1", "sample-project", "sample-project", []string{"sample-tag"})
	suite.Error(err) //TODO: check error message
	iamProject := IAMProject(IAMProject{ID: "", Name: "", Description: "", Tags: []string(nil)})
	suite.Equal(id, iamProject)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateProjectSuccess() {
	sampleResponse := `{
		"name": "sample-project",
		"description": "sample-project",
		"id": "1",
		"tags": ["sample-tag"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPut, "/v2/orgs/1/projects/1").
			WithBody([]byte(`{"description":"sample-project","name":"sample-project","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.UpdateProject("1", "1", "sample-project", "sample-project", []string{"sample-tag"})
	suite.NoError(err)
	iamProject := IAMProject(IAMProject{ID: "1", Name: "sample-project", Description: "sample-project", Tags: []string{"sample-tag"}})
	suite.Equal(id, iamProject)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateProjectError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPut, "/v2/orgs/1/projects/1").
			WithBody([]byte(`{"description":"sample-project","name":"sample-project","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.UpdateProject("1", "1", "sample-project", "sample-project", []string{"sample-tag"})
	suite.Error(err)
	iamProject := IAMProject(IAMProject{ID: "", Name: "", Description: "", Tags: []string(nil)})
	suite.Equal(id, iamProject)
	mockServer.HasExpectedRequests()
}
func (suite *RestClientIAMTestSuite) TestDeleteProjectSuccess() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1/projects/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusNoContent).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteProject("1", "1")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteProjectError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1/projects/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusBadRequest).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteProject("1", "1")
	suite.Error(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationMembershipSuccess() {
	expected := IAMOrganizationMembership(exampleIAMOrganizationMembership)
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v2/orgs/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusOK).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationMembership("1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationMembershipSuccess() {
	expected := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: exampleIAMOrganization, User: IAMOrganisationUser{ID: "", Email: ""}, Affiliation: "member", MembershipType: "service_account", Permissions: []string{"can_do"}})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v2/orgs/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusOK).
			ReturnWithBody([]byte(sampleResponse)),
		responses.Expect(http.MethodPatch, "/v2/orgs/1/memberships/1").
			WithBody([]byte(`{"affiliation":"member","editable_permissions":["can_do"],"membership_type":"service_account"}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateOrUpdateOrganizationMembership("1", "1", "member", []string{"can_do"})
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationMembershipError() {
	expected := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: exampleIAMOrganization, User: IAMOrganisationUser{ID: "", Email: ""}, Affiliation: "member", MembershipType: "service_account", Permissions: []string{"can_do"}})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v2/orgs/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusNotFound).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateOrUpdateOrganizationMembership("1", "1", "member", []string{"can_do"})
	suite.Error(err) //TODO: check error message
	iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "", Name: ""}, Permissions: []string(nil)})
	suite.Equal(id, iamOrgMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationMembershipSuccess() {
	expected := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: exampleIAMOrganization, User: IAMOrganisationUser{ID: "", Email: ""}, Affiliation: "member", MembershipType: "service_account", Permissions: []string{"can_do"}})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v2/orgs/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusOK).
			ReturnWithBody([]byte(sampleResponse)),
		responses.Expect(http.MethodPatch, "/v2/orgs/1/memberships/1").
			WithBody([]byte(`{"affiliation":"member","editable_permissions":["can_do"],"membership_type":"service_account"}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateOrganizationMembership("1", "1", "member", []string{"can_do"})
	suite.NoError(err)
	//iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "", Name: ""}, User: IAMOrganisationUser{ID: "", Email: ""}, Affiliation: "member", MembershipType: "service_account", Permissions: []string{"can_do"}})
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationMembershipError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v2/orgs/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusNotFound),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.UpdateOrganizationMembership("1", "1", "member", []string{"can_do"})
	suite.Error(err)
	iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "", Name: ""}, Permissions: []string(nil)})
	suite.Equal(id, iamOrgMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationMembershipSuccess() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1/service-accounts/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusNotFound).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteOrganizationMembership("1", "1")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationMembershipError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1/service-accounts/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusBadRequest).
			ReturnWithBody([]byte(``)),
		responses.Expect(http.MethodDelete, "/v2/orgs/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusBadRequest).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteOrganizationMembership("1", "1")
	suite.Error(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetProjectMembershipSuccess() {
	sampleResponse := `{
		"project": {
		"id": "1",
		"name": "syseleven"},
		"user": {"id": "1", "email": "test@syseleven.net"},
		"permissions": ["can_do"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v2/orgs/1/projects/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.GetProjectMembership("1", "1", "1")
	suite.NoError(err)
	iamProjectMembership := IAMProjectMembership(IAMProjectMembership{User: IAMOrganisationUser{ID: "1", Email: "test@syseleven.net"}, Permissions: []string{"can_do"}, Project: IAMProject{ID: "1", Name: "syseleven"}})
	suite.Equal(id, iamProjectMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectMembershipSuccess() {
	sampleResponse := `{
		"project": {
		"id": "1",
		"name": "syseleven"},
		"user": {"id": "1", "email": "test@syseleven.net"},
		"permissions": ["can_do"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/projects/1/memberships/1/permissions").
			WithBody([]byte(`["can_do"]`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateProjectMembership("1", "1", "1", []string{"can_do"})
	suite.NoError(err)
	iamProjectMembership := IAMProjectMembership(IAMProjectMembership{User: IAMOrganisationUser{ID: "1", Email: "test@syseleven.net"}, Permissions: []string{"can_do"}, Project: IAMProject{ID: "1", Name: "syseleven"}})
	suite.Equal(id, iamProjectMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectMembershipError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/projects/1/memberships/1/permissions").
			WithBody([]byte(`["can_do"]`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateProjectMembership("1", "1", "1", []string{"can_do"})
	suite.Error(err) //TODO: check error message
	iamProjectMembership := IAMProjectMembership(IAMProjectMembership{User: IAMOrganisationUser{ID: "", Email: ""}, Permissions: nil, Project: IAMProject{ID: "", Name: ""}})
	suite.Equal(id, iamProjectMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateProjectMembershipSuccess() {
	sampleResponse := `{
		"project": {
		"id": "1",
		"name": "syseleven"},
		"user": {"id": "1", "email": "test@syseleven.net"},
		"permissions": ["can_do"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/projects/1/memberships/1/permissions").
			WithBody([]byte(`["can_do"]`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ipm, err := client.UpdateProjectMembership("1", "1", "1", []string{"can_do"})
	suite.NoError(err)
	iamProjectMembership := IAMProjectMembership(IAMProjectMembership{User: IAMOrganisationUser{ID: "1", Email: "test@syseleven.net"}, Permissions: []string{"can_do"}, Project: IAMProject{ID: "1", Name: "syseleven"}})
	suite.Equal(ipm, iamProjectMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateProjectMembershipError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/projects/1/memberships/1/permissions").
			WithBody([]byte(`["can_do"]`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.UpdateProjectMembership("1", "1", "1", []string{"can_do"})
	suite.Error(err) //TODO: check error message
	iamProjectMembership := IAMProjectMembership(IAMProjectMembership{User: IAMOrganisationUser{ID: "", Email: ""}, Permissions: nil, Project: IAMProject{ID: "", Name: ""}})
	suite.Equal(id, iamProjectMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteProjectMembershipSuccess() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1/projects/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusNoContent).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteProjectMembership("1", "1", "1")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteProjectMembershipError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1/projects/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusBadRequest).
			ReturnWithBody([]byte(``)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteProjectMembership("1", "1", "1")
	suite.Error(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationMembershipByEmailSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/memberships"
	status := http.StatusOK
	expected := IAMOrganizationMembership(exampleIAMOrganizationMembership)
	var iamOrganizationMemberships []IAMOrganizationMembership
	iamOrganizationMemberships = append(iamOrganizationMemberships, expected)
	sampleResponse, err := json.Marshal(iamOrganizationMemberships)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationMembershipByEmail("1", "test@syseleven.net")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationInvitationByEmailSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/invitations"
	status := http.StatusOK
	expected := IAMOrganizationInvitation(IAMOrganizationInvitation{ID: "1", Email: "test@syseleven.net"})
	var iamOrganizationInvitations []IAMOrganizationInvitation
	iamOrganizationInvitations = append(iamOrganizationInvitations, expected)
	sampleResponse, err := json.Marshal(iamOrganizationInvitations)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationInvitationByEmail("1", "test@syseleven.net")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationTeamMembershipSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/teams/1/memberships/1"
	status := http.StatusOK
	expected := IAMOrganizationTeamMembership(IAMOrganizationTeamMembership{Organisation: exampleIAMOrganization})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationTeamMembership("1", "1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationServiceaccountSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/service-accounts/1"
	status := http.StatusOK
	expected := IAMOrganizationServiceaccount(IAMOrganizationServiceaccount{})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationServiceaccount("1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationContactSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/contacts/1"
	status := http.StatusOK
	expected := IAMOrganizationContact(IAMOrganizationContact{})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationContact("1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationByNameSuccess() {
	method := http.MethodGet
	url := "/v2/orgs"
	status := http.StatusOK
	expected := exampleIAMOrganization
	var iamOrganization []IAMOrganization
	iamOrganization = append(iamOrganization, expected)
	sampleResponse, err := json.Marshal(iamOrganization)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationByName("sample-org")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetOrganizationTeamSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/teams/1"
	status := http.StatusOK
	expected := IAMOrganizationTeam(IAMOrganizationTeam{})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetOrganizationTeam("1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationServiceaccountSuccess() {
	method := http.MethodPost
	url := "/v2/orgs/1/service-accounts"
	status := http.StatusCreated
	expected := IAMOrganizationServiceaccount(IAMOrganizationServiceaccount{ID: "1", OrganizationId: "1"})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateOrganizationServiceaccount("1", "test", "test")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetProjectMembershipByEmailSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/projects/1/memberships"
	status := http.StatusOK
	expected := IAMProjectMembership(IAMProjectMembership{User: IAMOrganisationUser{Email: "test@syseleven.net"}})
	var iamProjectMemberships []IAMProjectMembership
	iamProjectMemberships = append(iamProjectMemberships, expected)
	sampleResponse, err := json.Marshal(iamProjectMemberships)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetProjectMembershipByEmail("1", "1", "test@syseleven.net")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetProjectTeamPermissionsSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/projects/1/teams/1/permissions"
	status := http.StatusOK
	expected := []string{"can_do"}
	sampleResponse, err := json.Marshal([]string{"can_do"})
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetProjectTeamPermissions("1", "1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetProjectTeamMembershipSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/projects/1/teams/1/memberships/1"
	status := http.StatusOK
	expected := IAMProjectTeamMembership(IAMProjectTeamMembership{})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetProjectTeamMembership("1", "1", "1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetProjectS3UserSuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/projects/1/s3-users"
	status := http.StatusOK
	expected := IAMProjectS3User(IAMProjectS3User{ID: "1"})
	sampleResponse, err := json.Marshal([]IAMProjectS3User{expected})
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetProjectS3User("1", "1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestGetProjectS3UserKeySuccess() {
	method := http.MethodGet
	url := "/v2/orgs/1/projects/1/s3-users/1/ec2-credentials/1"
	status := http.StatusOK
	expected := IAMProjectS3UserKey(exampleIAMProjectS3KeyUser)
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.GetProjectS3UserKey("1", "1", "1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationServiceaccountSuccess() {
	method := "PUT"
	url := "/v2/orgs/1/service-accounts/1"
	status := http.StatusOK
	expected := IAMOrganizationServiceaccount(IAMOrganizationServiceaccount{ID: "1", Description: "desc", Name: "name", OrganizationId: "1"})
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateOrganizationServiceaccount("1", "1", "name", "desc")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationServiceaccountSuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/service-accounts/1"
	status := http.StatusNoContent
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err := client.DeleteOrganizationServiceaccount("1", "1")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationTeamMembershipSuccess() {
	method := "POST"
	url := "/v2/orgs/1/teams/1/memberships/1"
	status := http.StatusOK
	expected := IAMOrganizationTeamMembership(exampleIAMOrganizationTeamMembership)
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateOrganizationTeamMembership("1", "1", "1")
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectTeamPermissionsSuccess() {
	method := "POST"
	url := "/v2/orgs/1/projects/1/teams/1/permissions"
	status := http.StatusOK
	expected := IAMProjectTeamPermissions(exampleIAMProjectTeamPermissions)
	sampleResponse, err := json.Marshal(expected)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateProjectTeamPermissions("1", "1", "1", []string{"can_do"})
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationInvitationSuccess() {
	method := "POST"
	url := "/v2/orgs/1/invitations"
	status := http.StatusOK
	body := []IAMOrganizationInvitation{exampleIAMOrganizationInvitation}
	expected := IAMOrganizationInvitation(exampleIAMOrganizationInvitation)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateOrganizationInvitation(examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationContactSuccess() {
	method := "POST"
	url := "/v2/orgs/1/contacts"
	status := http.StatusOK
	body := IAMOrganizationContact(exampleIAMOrganizationContact)
	expected := IAMOrganizationContact(exampleIAMOrganizationContact)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateOrganizationContact(examplestring, examplestring, examplestring, examplestring, examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationTeamSuccess() {
	method := "POST"
	url := "/v2/orgs/1/teams"
	status := http.StatusOK
	body := IAMOrganizationTeam(exampleIAMOrganizationTeam)
	expected := IAMOrganizationTeam(exampleIAMOrganizationTeam)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateOrganizationTeam(examplestring, examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectTeamMembershipSuccess() {
	method := "POST"
	url := "/v2/orgs/1/projects/1/teams/1/memberships/1/permissions"
	status := http.StatusOK
	body := IAMProjectTeamMembership(exampleIAMProjectTeamMembership)
	expected := IAMProjectTeamMembership(exampleIAMProjectTeamMembership)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateProjectTeamMembership(examplestring, examplestring, examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectS3UserSuccess() {
	method := "POST"
	url := "/v2/orgs/1/projects/1/s3-users"
	status := http.StatusOK
	body := IAMProjectS3User(exampleIAMProjectS3User)
	expected := IAMProjectS3User(exampleIAMProjectS3User)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateProjectS3User(examplestring, examplestring, examplestring, examplestring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateProjectS3UserKeySuccess() {
	method := "POST"
	url := "/v2/orgs/1/projects/1/s3-users/1/ec2-credentials"
	status := http.StatusOK
	body := IAMProjectS3UserKey(exampleIAMProjectS3KeyUser)
	expected := IAMProjectS3UserKey(exampleIAMProjectS3KeyUser)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.CreateProjectS3UserKey(examplestring, examplestring, examplestring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationTeamMembershipSuccess() {
	method := "POST"
	url := "/v2/orgs/1/teams/1/memberships/1"
	status := http.StatusOK
	body := IAMOrganizationTeamMembership(exampleIAMOrganizationTeamMembership)
	expected := IAMOrganizationTeamMembership(exampleIAMOrganizationTeamMembership)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateOrganizationTeamMembership(examplestring, examplestring, examplestring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateProjectTeamPermissionsSuccess() {
	method := "POST"
	url := "/v2/orgs/1/projects/1/teams/1/permissions"
	status := http.StatusOK
	body := []string{"can_do"}
	expected := []string{"can_do"}
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateProjectTeamPermissions(examplestring, examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateProjectTeamMembershipSuccess() {
	method := "PATCH"
	url := "/v2/orgs/1/projects/1/teams/1/memberships/1/permissions"
	status := http.StatusOK
	body := IAMProjectTeamMembership(exampleIAMProjectTeamMembership)
	expected := IAMProjectTeamMembership(exampleIAMProjectTeamMembership)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateProjectTeamMembership(examplestring, examplestring, examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationContactSuccess() {
	method := "PUT"
	url := "/v2/orgs/1/contacts/1"
	status := http.StatusOK
	body := IAMOrganizationContact(exampleIAMOrganizationContact)
	expected := IAMOrganizationContact(exampleIAMOrganizationContact)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateOrganizationContact(examplestring, examplestring, examplestring, examplestring, examplestring, examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationTeamSuccess() {
	method := "PUT"
	url := "/v2/orgs/1/teams/1"
	status := http.StatusOK
	body := IAMOrganizationTeam(exampleIAMOrganizationTeam)
	expected := IAMOrganizationTeam(exampleIAMOrganizationTeam)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateOrganizationTeam(examplestring, examplestring, examplestring, examplestring, exampleSlicedstring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateProjectS3UserSuccess() {
	method := "PUT"
	url := "/v2/orgs/1/projects/1/s3-users/1"
	status := http.StatusOK
	body := IAMProjectS3User(exampleIAMProjectS3User)
	expected := IAMProjectS3User(exampleIAMProjectS3User)
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	ret, err := client.UpdateProjectS3User(examplestring, examplestring, examplestring, examplestring, examplestring)
	suite.NoError(err)
	suite.Equal(expected, ret)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationTeamMembershipSuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/teams/1/memberships/1"
	status := http.StatusOK
	body := ""
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteOrganizationTeamMembership(examplestring, examplestring, examplestring)
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteProjectTeamPermissionsSuccess() {
	method := "PATCH"
	url := "/v2/orgs/1/projects/1/teams/1/permissions"
	status := http.StatusOK
	body := "[]"
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteProjectTeamPermissions(examplestring, examplestring, examplestring)
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationInvitationSuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/invitations/1"
	status := http.StatusOK
	sampleResponse, err := json.Marshal([]IAMOrganizationInvitation{exampleIAMOrganizationInvitation})
	body := ""
	sampleResponse2, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect("GET", "/v2/orgs/1/invitations").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse2)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteOrganizationInvitation(examplestring, "test@syseleven.net")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteProjectTeamMembershipSuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/projects/1/teams/1/memberships/1/permissions"
	status := http.StatusOK
	body := ""
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteProjectTeamMembership(examplestring, examplestring, examplestring, examplestring)
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationContactSuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/contacts/1"
	status := http.StatusOK
	body := ""
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteOrganizationContact(examplestring, examplestring)
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationTeamSuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/teams/1"
	status := http.StatusOK
	body := ""
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteOrganizationTeam(examplestring, examplestring)
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteProjectS3UserSuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/projects/1/s3-users/1"
	status := http.StatusOK
	body := ""
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteProjectS3User(examplestring, examplestring, examplestring)
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteProjectS3UserKeySuccess() {
	method := "DELETE"
	url := "/v2/orgs/1/projects/1/s3-users/1/ec2-credentials/1"
	status := http.StatusOK
	body := ""
	sampleResponse, err := json.Marshal(body)
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(method, url).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(status).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	err = client.DeleteProjectS3UserKey(examplestring, examplestring, examplestring, examplestring)
	suite.NoError(err)
	mockServer.HasExpectedRequests()
}

func TestRestClientIAMTestSuite(t *testing.T) {
	suite.Run(t, new(RestClientIAMTestSuite))
}
