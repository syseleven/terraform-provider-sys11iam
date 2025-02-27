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
		"is_active": true
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
	iamOrg := IAMOrganization(IAMOrganization{ID: "1", Name: "sample-org", Description: "sample-org", Tags: []string{"sample-tag"}, CreatedAt: "date", IsActive: true, UpdatedAt: "date"})
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
		"is_active": true
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v1/orgs").
			WithBody([]byte(`{"company_info":{"accepted_tos":true,"city":"testcity","company_name":"testcompany","country":"testland","phone":"+49123456789","preferred_billing_method":"SEPA","street":"teststreet","street_number":"1","vat_id":"42069","zip_code":"12345"},"description":"sample-org","name":"sample-org","tags":["sample-tag"]}`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
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
	suite.NoError(err)
	iamOrg := IAMOrganization(IAMOrganization{ID: "1", Name: "sample-org", Description: "sample-org", Tags: []string{"sample-tag"}, CreatedAt: "date", IsActive: true, UpdatedAt: "date"})
	suite.Equal(id, iamOrg)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v1/orgs").
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
	sampleResponse := `{
		"name": "sample-org",
		"description": "sample-org",
		"id": "1",
		"created_at": "date",
		"updated_at": "date",
		"tags": ["sample-tag"],
		"is_active": true
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
	suite.NoError(err)
	iamOrg := IAMOrganization(IAMOrganization{ID: "1", Name: "sample-org", Description: "sample-org", Tags: []string{"sample-tag"}, CreatedAt: "date", IsActive: true, UpdatedAt: "date"})
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

func (suite *RestClientIAMTestSuite) TestGetProjectSuccess() {
	sampleResponse := `{
		"name": "sample-project",
		"description": "sample-project",
		"id": "1",
		"tags": ["sample-tag"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v1/orgs/1/projects/1").
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
		responses.Expect(http.MethodPost, "/v1/orgs/1/projects").
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
		responses.Expect(http.MethodPost, "/v1/orgs/1/projects").
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
		responses.Expect(http.MethodPut, "/v1/orgs/1/projects/1").
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
		responses.Expect(http.MethodPut, "/v1/orgs/1/projects/1").
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
		responses.Expect(http.MethodDelete, "/v1/orgs/1/projects/1").
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
		responses.Expect(http.MethodDelete, "/v1/orgs/1/projects/1").
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
	sampleResponse := `{
		"organization": {
			"id": "1",
			"name": "syseleven"
		},
		"user": {"id": "1", "email": "test@syseleven.net"},
		"editable_permissions": ["can_do"]
	  }`
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

	id, err := client.GetOrganizationMembership("1", "1")
	suite.NoError(err)
	iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "1", Name: "syseleven"}, User: IAMOrganisationUser{ID: "1", Email: "test@syseleven.net"}, Permissions: []string{"can_do"}})
	suite.Equal(id, iamOrgMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationMembershipSuccess() {
	sampleResponse := `{
		"organization": {
			"id": "1",
			"name": "syseleven"
		},
		"user": {"id": "1", "email": "test@syseleven.net"},
		"editable_permissions": ["can_do"]
	  }`
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/memberships/1/permissions").
			WithBody([]byte(`["can_do"]`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody([]byte(sampleResponse)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	iom, err := client.CreateOrganizationMembership("1", "1", []string{"can_do"})
	suite.NoError(err)
	iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "1", Name: "syseleven"}, User: IAMOrganisationUser{ID: "1", Email: "test@syseleven.net"}, Permissions: []string{"can_do"}})
	suite.Equal(iom, iamOrgMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestCreateOrganizationMembershipError() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPost, "/v2/orgs/1/memberships/1/permissions").
			WithBody([]byte(`["can_do"]`)).
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusConflict).
			ReturnWithBody([]byte(`{}`)),
	)
	defer mockServer.Close()
	client := NewClient(mockServer.URL, 0).WithBearerToken("testtoken")

	id, err := client.CreateOrganizationMembership("1", "1", []string{"can_do"})
	suite.Error(err) //TODO: check error message
	iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "", Name: ""}, Permissions: []string(nil)})
	suite.Equal(id, iamOrgMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestUpdateOrganizationMembershipSuccess() {
	sampleResponse := `{
		"membership_type": "service_account",
        "affiliation": "member",
		"editable_permissions": ["can_do"]
	  }`
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

	id, err := client.UpdateOrganizationMembership("1", "1", []string{"can_do"})
	suite.NoError(err)
	iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "", Name: ""}, User: IAMOrganisationUser{ID: "", Email: ""}, Affiliation: "member", MembershipType: "service_account", Permissions: []string{"can_do"}})
	suite.Equal(id, iamOrgMembership)
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

	id, err := client.UpdateOrganizationMembership("1", "1", []string{"can_do"})
	suite.Error(err)
	iamOrgMembership := IAMOrganizationMembership(IAMOrganizationMembership{Organisation: IAMOrganization{ID: "", Name: ""}, Permissions: []string(nil)})
	suite.Equal(id, iamOrgMembership)
	mockServer.HasExpectedRequests()
}

func (suite *RestClientIAMTestSuite) TestDeleteOrganizationMembershipSuccess() {
	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodDelete, "/v2/orgs/1/memberships/1").
			WithHeaders(map[string]string{
				"Authorization": "Bearer testtoken",
			}).
			ReturnWithCode(http.StatusNoContent).
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

func TestRestClientIAMTestSuite(t *testing.T) {
	suite.Run(t, new(RestClientIAMTestSuite))
}
