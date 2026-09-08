package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/suite"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	responses "github.com/syseleven/terraform-provider-sys11iam/internal/http-responses"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_team"
)

type OrganizationTeamResourceTestSuite struct {
	suite.Suite
}

func TestOrganizationTeamResource(t *testing.T) {
	suite.Run(t, new(OrganizationTeamResourceTestSuite))
}

const (
	testOrgID    = "org-1"
	testTeamID   = "team-1"
	testTeamName = "SaaS"
)

func teamAttributeTypes() map[string]tftypes.Type {
	listType := tftypes.List{ElementType: tftypes.String}
	return map[string]tftypes.Type{
		"editable_permissions": listType,
		"description":          tftypes.String,
		"id":                   tftypes.String,
		"name":                 tftypes.String,
		"tags":                 listType,
		"organization_id":      tftypes.String,
	}
}

func teamRaw(id tftypes.Value, name tftypes.Value, tags tftypes.Value) tftypes.Value {
	listType := tftypes.List{ElementType: tftypes.String}
	return tftypes.NewValue(tftypes.Object{AttributeTypes: teamAttributeTypes()}, map[string]tftypes.Value{
		"editable_permissions": tftypes.NewValue(listType, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "can_read_members_in_org"),
		}),
		"description":     tftypes.NewValue(tftypes.String, "Members of the Managed Solutions Oncall"),
		"id":              id,
		"name":            name,
		"tags":            tags,
		"organization_id": tftypes.NewValue(tftypes.String, testOrgID),
	})
}

func (suite *OrganizationTeamResourceTestSuite) orgBody() []byte {
	orgResponse, err := json.Marshal(iam.IAMOrganization{
		ID:        testOrgID,
		Name:      "org",
		Tags:      []string{},
		IsActive:  true,
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	})
	suite.Require().NoError(err)
	return orgResponse
}

func (suite *OrganizationTeamResourceTestSuite) teamBody() []byte {
	teamResponse, err := json.Marshal(iam.IAMOrganizationTeam{
		ID:          testTeamID,
		Name:        testTeamName,
		Description: "Members of the Managed Solutions Oncall",
		Tags:        []string{"saas-tag"},
	})
	suite.Require().NoError(err)
	return teamResponse
}

func (suite *OrganizationTeamResourceTestSuite) configureResource(res resource.ResourceWithConfigure, server *responses.MockServer) {
	resp := resource.ConfigureResponse{}
	res.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: iam.NewClient(server.URL, 0).WithBearerToken("testtoken"),
	}, &resp)
	suite.Require().False(resp.Diagnostics.HasError())
}

// TestCreateWithUnknownTags reproduces the "Value Conversion Error" failure:
// tags is optional+computed and not set in config, so the plan value is
// unknown. Create must not error and must send an empty tags list to the API.
func (suite *OrganizationTeamResourceTestSuite) TestCreateWithUnknownTags() {
	ctx := context.Background()
	schema := resource_organization_team.OrganizationTeamResourceSchema(ctx)
	listType := tftypes.List{ElementType: tftypes.String}

	unknownTags := tftypes.NewValue(listType, tftypes.UnknownValue)
	unknownID := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	planRaw := teamRaw(unknownID, tftypes.NewValue(tftypes.String, testTeamName), unknownTags)

	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v1/orgs/"+testOrgID).
			WithHeaders(map[string]string{"Authorization": "Bearer testtoken"}).
			ReturnWithCode(http.StatusOK).
			ReturnWithBody(suite.orgBody()),
		responses.Expect(http.MethodPost, "/v2/orgs/"+testOrgID+"/teams").
			WithHeaders(map[string]string{"Authorization": "Bearer testtoken"}).
			WithJSONParameters(map[string]interface{}{
				"name":        testTeamName,
				"description": "Members of the Managed Solutions Oncall",
				"tags":        []string{},
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody(suite.teamBody()),
	)
	defer mockServer.Close()

	res := &OrganizationTeamResource{}
	suite.configureResource(res, mockServer)

	req := resource.CreateRequest{
		Config: tfsdk.Config{Raw: planRaw, Schema: schema},
		Plan:   tfsdk.Plan{Raw: planRaw, Schema: schema},
	}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	res.Create(ctx, req, &resp)

	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	mockServer.HasExpectedRequests()

	var data resource_organization_team.OrganizationTeamModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &data)...)
	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	suite.Equal(testTeamID, data.Id.ValueString())
	suite.False(data.Tags.IsNull(), "tags must not be null in state")
	var tags []string
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	suite.Equal([]string{"saas-tag"}, tags)
}

// TestCreateWithKnownTags locks in the unchanged behavior when tags are set
// in config: the configured tags must be sent to the API.
func (suite *OrganizationTeamResourceTestSuite) TestCreateWithKnownTags() {
	ctx := context.Background()
	schema := resource_organization_team.OrganizationTeamResourceSchema(ctx)
	listType := tftypes.List{ElementType: tftypes.String}

	knownTags := tftypes.NewValue(listType, []tftypes.Value{
		tftypes.NewValue(tftypes.String, "oncall"),
		tftypes.NewValue(tftypes.String, "saas"),
	})
	unknownID := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	planRaw := teamRaw(unknownID, tftypes.NewValue(tftypes.String, testTeamName), knownTags)

	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodGet, "/v1/orgs/"+testOrgID).
			WithHeaders(map[string]string{"Authorization": "Bearer testtoken"}).
			ReturnWithCode(http.StatusOK).
			ReturnWithBody(suite.orgBody()),
		responses.Expect(http.MethodPost, "/v2/orgs/"+testOrgID+"/teams").
			WithHeaders(map[string]string{"Authorization": "Bearer testtoken"}).
			WithJSONParameters(map[string]interface{}{
				"name":        testTeamName,
				"description": "Members of the Managed Solutions Oncall",
				"tags":        []string{"oncall", "saas"},
			}).
			ReturnWithCode(http.StatusCreated).
			ReturnWithBody(suite.teamBody()),
	)
	defer mockServer.Close()

	res := &OrganizationTeamResource{}
	suite.configureResource(res, mockServer)

	req := resource.CreateRequest{
		Config: tfsdk.Config{Raw: planRaw, Schema: schema},
		Plan:   tfsdk.Plan{Raw: planRaw, Schema: schema},
	}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	res.Create(ctx, req, &resp)

	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	mockServer.HasExpectedRequests()

	var data resource_organization_team.OrganizationTeamModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &data)...)
	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	suite.Equal(testTeamID, data.Id.ValueString())
	var tags []string
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	suite.Equal([]string{"saas-tag"}, tags)
}

// TestUpdateWithUnknownTagsAfterNullState covers the update path: prior state
// holds null tags (API returned no tags) and the plan value is unknown.
// Update must not error and must not write unknown tags into state.
func (suite *OrganizationTeamResourceTestSuite) TestUpdateWithUnknownTagsAfterNullState() {
	ctx := context.Background()
	schema := resource_organization_team.OrganizationTeamResourceSchema(ctx)
	listType := tftypes.List{ElementType: tftypes.String}

	nullTags := tftypes.NewValue(listType, nil)
	unknownTags := tftypes.NewValue(listType, tftypes.UnknownValue)
	knownID := tftypes.NewValue(tftypes.String, testTeamID)
	knownName := tftypes.NewValue(tftypes.String, testTeamName)

	stateRaw := teamRaw(knownID, knownName, nullTags)
	planRaw := teamRaw(knownID, knownName, unknownTags)

	mockServer := responses.NewMockServer(
		&suite.Suite,
		responses.Expect(http.MethodPut, "/v2/orgs/"+testOrgID+"/teams/"+testTeamID).
			WithHeaders(map[string]string{"Authorization": "Bearer testtoken"}).
			WithJSONParameters(map[string]interface{}{
				"name":        testTeamName,
				"description": "Members of the Managed Solutions Oncall",
				"tags":        []string{},
			}).
			ReturnWithCode(http.StatusOK).
			ReturnWithBody(suite.teamBody()),
	)
	defer mockServer.Close()

	res := &OrganizationTeamResource{}
	suite.configureResource(res, mockServer)

	req := resource.UpdateRequest{
		State: tfsdk.State{Raw: stateRaw, Schema: schema},
		Plan:  tfsdk.Plan{Raw: planRaw, Schema: schema},
	}
	resp := resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	res.Update(ctx, req, &resp)

	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	mockServer.HasExpectedRequests()

	var data resource_organization_team.OrganizationTeamModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &data)...)
	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	suite.Equal(testTeamID, data.Id.ValueString())
	suite.False(data.Tags.IsUnknown(), "tags must not remain unknown in state")
	var tags []string
	resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
	suite.Require().False(resp.Diagnostics.HasError(), "expected no error diagnostics, got %d", resp.Diagnostics.ErrorsCount())
	suite.Equal([]string{"saas-tag"}, tags)
}
