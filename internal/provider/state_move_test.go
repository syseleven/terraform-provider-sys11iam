package provider

import (
	"context"
	"encoding/json"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_membership"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_s3_user"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_s3_user_key"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_team"
)

type resourceWithMoveAndSchema interface {
	fwresource.ResourceWithMoveState
	Schema(context.Context, fwresource.SchemaRequest, *fwresource.SchemaResponse)
}

func moveStateForTest(t *testing.T, target resourceWithMoveAndSchema, sourceTypeName string, rawJSON string) tfsdk.State {
	t.Helper()

	ctx := context.Background()

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
	}

	resp := fwresource.MoveStateResponse{
		TargetState: state,
	}

	movers := target.MoveState(ctx)
	require.NotEmpty(t, movers)

	movers[0].StateMover(ctx, fwresource.MoveStateRequest{
		SourceProviderAddress: "sys11iam",
		SourceRawState:        &tfprotov6.RawState{JSON: []byte(rawJSON)},
		SourceSchemaVersion:   0,
		SourceTypeName:        sourceTypeName,
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	return resp.TargetState
}

func TestProjectResourceMoveStateFromProject(t *testing.T) {
	state := moveStateForTest(t, &ProjectResource{}, "sys11iam_project", `{
		"description": "old project",
		"id": "project-1",
		"name": "project-name",
		"organization_id": "org-1",
		"tags": ["one", "two"]
	}`)

	var data resource_organization_project.OrganizationProjectModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)

	require.Equal(t, "project-1", data.Id.ValueString())
	require.Equal(t, "project-1", data.ProjectId.ValueString())
	require.Equal(t, "org-1", data.OrgId.ValueString())
	require.Equal(t, "org-1", data.OrganizationId.ValueString())
	require.Equal(t, "project-name", data.Name.ValueString())
	require.Equal(t, "old project", data.Description.ValueString())

	var tags []string
	diags = data.Tags.ElementsAs(context.Background(), &tags, false)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, []string{"one", "two"}, tags)
}

func TestProjectS3UserResourceMoveStateFromProjectS3User(t *testing.T) {
	state := moveStateForTest(t, &ProjectS3UserResource{}, "sys11iam_project_s3user", `{
		"description": "old s3 user",
		"id": "s3-user-1",
		"name": "s3username",
		"organization_id": "org-1",
		"project_id": "project-1"
	}`)

	var data resource_organization_project_s3_user.OrganizationProjectS3UserModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)

	require.Equal(t, "s3-user-1", data.Id.ValueString())
	require.Equal(t, "s3-user-1", data.S3UserId.ValueString())
	require.Equal(t, "org-1", data.OrgId.ValueString())
	require.Equal(t, "project-1", data.ProjectId.ValueString())
	require.True(t, data.Keys.IsNull())
}

func TestProjectS3UserKeyResourceMoveStateFromProjectS3UserKey(t *testing.T) {
	state := moveStateForTest(t, &ProjectS3UserKeyResource{}, "sys11iam_project_s3user_key", `{
		"organization_id": "org-1",
		"project_id": "project-1",
		"s3_access_key": "legacy-access-key",
		"s3_user_id": "s3-user-1",
		"secret_key": "secret"
	}`)

	var data resource_organization_project_s3_user_key.OrganizationProjectS3UserKeyModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)

	require.Equal(t, "legacy-access-key", data.AccessKey.ValueString())
	require.Equal(t, "org-1", data.OrgId.ValueString())
	require.Equal(t, "project-1", data.ProjectId.ValueString())
	require.Equal(t, "s3-user-1", data.S3UserId.ValueString())
	require.Equal(t, "secret", data.SecretKey.ValueString())
}

func TestProjectMembershipResourceMoveStateFromProjectMembership(t *testing.T) {
	state := moveStateForTest(t, &ProjectMembershipResource{}, "sys11iam_project_membership", `{
		"email": "user@example.com",
		"id": "user-1",
		"organization_id": "org-1",
		"permissions": ["can_read", "can_write"],
		"project_id": "project-1"
	}`)

	var data resource_organization_project_membership.OrganizationProjectMembershipModel
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)

	require.Equal(t, "user-1", data.Id.ValueString())
	require.Equal(t, "org-1", data.OrgId.ValueString())
	require.Equal(t, "project-1", data.ProjectId.ValueString())
	require.False(t, data.Membership.IsNull())

	var membership resource_organization_project_membership.MembershipValue
	diags = data.Membership.As(context.Background(), &membership, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), diags)
	require.False(t, membership.UserMembership.IsNull())
	require.True(t, membership.ServiceAccountMembership.IsNull())

	var userMembership resource_organization_project_membership.UserMembershipValue
	diags = membership.UserMembership.As(context.Background(), &userMembership, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), diags)
	require.Equal(t, "user", userMembership.MembershipType.ValueString())

	var user resource_organization_project_membership.UserValue
	diags = userMembership.User.As(context.Background(), &user, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), diags)
	require.Equal(t, "user@example.com", user.Email.ValueString())
	require.Equal(t, "user-1", user.Id.ValueString())

	var permissions []string
	diags = userMembership.Permissions.ElementsAs(context.Background(), &permissions, false)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, []string{"can_read", "can_write"}, permissions)
}

func TestProjectMembershipResourceMoveStateRejectsStateWithoutEmail(t *testing.T) {
	ctx := context.Background()
	target := &ProjectMembershipResource{}

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	resp := fwresource.MoveStateResponse{
		TargetState: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}

	target.MoveState(ctx)[0].StateMover(ctx, fwresource.MoveStateRequest{
		SourceProviderAddress: "registry.terraform.io/syseleven/sys11iam",
		SourceRawState: &tfprotov6.RawState{JSON: []byte(`{
			"id": "service-account-1",
			"organization_id": "org-1",
			"permissions": ["can_read"],
			"project_id": "project-1"
		}`)},
		SourceSchemaVersion: 0,
		SourceTypeName:      "sys11iam_project_membership",
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestMoveStateSkipsUnknownSourceType(t *testing.T) {
	ctx := context.Background()
	target := &ProjectResource{}

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	nullState := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)
	resp := fwresource.MoveStateResponse{
		TargetState: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    nullState,
		},
	}

	target.MoveState(ctx)[0].StateMover(ctx, fwresource.MoveStateRequest{
		SourceProviderAddress: "registry.terraform.io/syseleven/sys11iam",
		SourceRawState:        &tfprotov6.RawState{JSON: []byte(`{"id":"project-1"}`)},
		SourceSchemaVersion:   0,
		SourceTypeName:        "sys11iam_unknown",
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.True(t, resp.TargetState.Raw.Equal(nullState))
}

func TestMoveStateReportsMalformedMatchedState(t *testing.T) {
	ctx := context.Background()
	target := &ProjectResource{}

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	resp := fwresource.MoveStateResponse{
		TargetState: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}

	target.MoveState(ctx)[0].StateMover(ctx, fwresource.MoveStateRequest{
		SourceProviderAddress: "registry.terraform.io/syseleven/sys11iam",
		SourceRawState:        &tfprotov6.RawState{JSON: []byte(`{"id":`)},
		SourceSchemaVersion:   0,
		SourceTypeName:        "sys11iam_project",
	}, &resp)

	require.True(t, resp.Diagnostics.HasError())
}

func TestMoveStateSkipsUnknownProviderAddress(t *testing.T) {
	ctx := context.Background()
	target := &ProjectResource{}

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	nullState := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)
	resp := fwresource.MoveStateResponse{
		TargetState: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    nullState,
		},
	}

	target.MoveState(ctx)[0].StateMover(ctx, fwresource.MoveStateRequest{
		SourceProviderAddress: "registry.terraform.io/example/sys11iam",
		SourceRawState:        &tfprotov6.RawState{JSON: []byte(`{"id":"project-1"}`)},
		SourceSchemaVersion:   0,
		SourceTypeName:        "sys11iam_project",
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.True(t, resp.TargetState.Raw.Equal(nullState))
}

func TestMoveStateAcceptsNamespacedProviderAddress(t *testing.T) {
	state := moveStateWithProviderAddressForTest(t, &ProjectResource{}, "syseleven/sys11iam")

	var data resource_organization_project.OrganizationProjectModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, "project-1", data.Id.ValueString())
}

func TestMoveStateAcceptsMirroredProviderAddress(t *testing.T) {
	state := moveStateWithProviderAddressForTest(t, &ProjectResource{}, "terraform-mirror.example/syseleven/sys11iam")

	var data resource_organization_project.OrganizationProjectModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, "project-1", data.Id.ValueString())
}

func TestMoveStateKeepsExistingOrgID(t *testing.T) {
	state := moveStateForTest(t, &ProjectResource{}, "sys11iam_project", `{
		"description": "old project",
		"id": "project-1",
		"name": "project-name",
		"org_id": "org-new",
		"organization_id": "org-old"
	}`)

	var data resource_organization_project.OrganizationProjectModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, "org-new", data.OrgId.ValueString())
	require.Equal(t, "org-new", data.OrganizationId.ValueString())
}

func moveStateWithProviderAddressForTest(t *testing.T, target resourceWithMoveAndSchema, providerAddress string) tfsdk.State {
	t.Helper()

	ctx := context.Background()

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	resp := fwresource.MoveStateResponse{
		TargetState: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}

	target.MoveState(ctx)[0].StateMover(ctx, fwresource.MoveStateRequest{
		SourceProviderAddress: providerAddress,
		SourceRawState: &tfprotov6.RawState{JSON: []byte(`{
			"description": "old project",
			"id": "project-1",
			"name": "project-name",
			"organization_id": "org-1",
			"tags": []
		}`)},
		SourceSchemaVersion: 0,
		SourceTypeName:      "sys11iam_project",
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	return resp.TargetState
}

// moveAllStateForTest runs every state mover of the target resource, like
// Terraform does, and is used for resources that register multiple movers.
func moveAllStateForTest(t *testing.T, target resourceWithMoveAndSchema, sourceTypeName string, rawJSON string) tfsdk.State {
	t.Helper()

	ctx := context.Background()

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	state := tfsdk.State{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
	}

	resp := fwresource.MoveStateResponse{
		TargetState: state,
	}

	movers := target.MoveState(ctx)
	require.NotEmpty(t, movers)

	for _, mover := range movers {
		mover.StateMover(ctx, fwresource.MoveStateRequest{
			SourceProviderAddress: "sys11iam",
			SourceRawState:        &tfprotov6.RawState{JSON: []byte(rawJSON)},
			SourceSchemaVersion:   0,
			SourceTypeName:        sourceTypeName,
		}, &resp)
	}

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	return resp.TargetState
}

func TestOrganizationTeamResourceMoveStateFromOrganizationTeam(t *testing.T) {
	state := moveAllStateForTest(t, &OrganizationTeamResource{}, "sys11iam_organization_team", `{
		"description": "test team",
		"editable_permissions": ["can_become_project_administrator_in_org"],
		"id": "team-1",
		"name": "testteam",
		"organization_id": "org-1",
		"tags": ["testing2"]
	}`)

	var data resource_organization_team.OrganizationTeamModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)

	require.Equal(t, "team-1", data.Id.ValueString())
	require.Equal(t, "org-1", data.OrgId.ValueString())
	require.Equal(t, "org-1", data.OrganizationId.ValueString())
	require.Equal(t, "testteam", data.Name.ValueString())
	require.Equal(t, "test team", data.Description.ValueString())

	var tags []string
	diags = data.Tags.ElementsAs(context.Background(), &tags, false)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, []string{"testing2"}, tags)

	var permissions []string
	diags = data.OrganizationPermissions.ElementsAs(context.Background(), &permissions, false)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, []string{"can_become_project_administrator_in_org"}, permissions)
	require.True(t, data.Projects.IsNull())
}

func TestOrganizationTeamResourceMoveStateFromProjectTeam(t *testing.T) {
	state := moveAllStateForTest(t, &OrganizationTeamResource{}, "sys11iam_project_team", `{
		"editable_permissions": ["can_become_administrator_in_project"],
		"organization_id": "org-1",
		"project_id": "project-1",
		"team_id": "team-1"
	}`)

	var data resource_organization_team.OrganizationTeamModelFull
	diags := state.Get(context.Background(), &data)
	require.False(t, diags.HasError(), diags)

	require.Equal(t, "team-1", data.Id.ValueString())
	require.Equal(t, "org-1", data.OrgId.ValueString())
	require.Equal(t, "org-1", data.OrganizationId.ValueString())
	require.True(t, data.Name.IsNull())
	require.True(t, data.Description.IsNull())
	require.True(t, data.Tags.IsNull())
	require.True(t, data.OrganizationPermissions.IsNull())

	var projects []*resource_organization_team.ProjectValue
	diags = data.Projects.ElementsAs(context.Background(), &projects, false)
	require.False(t, diags.HasError(), diags)
	require.Len(t, projects, 1)
	require.Equal(t, "project-1", projects[0].Id.ValueString())

	var permissions []string
	diags = projects[0].ProjectPermissions.ElementsAs(context.Background(), &permissions, false)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, []string{"can_become_administrator_in_project"}, permissions)
}

func TestOrganizationTeamMoveStateRejectsProjectTeamWithoutTeamID(t *testing.T) {
	ctx := context.Background()
	target := &OrganizationTeamResource{}

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	resp := fwresource.MoveStateResponse{
		TargetState: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}

	for _, mover := range target.MoveState(ctx) {
		mover.StateMover(ctx, fwresource.MoveStateRequest{
			SourceProviderAddress: "registry.terraform.io/syseleven/sys11iam",
			SourceRawState: &tfprotov6.RawState{JSON: []byte(`{
				"editable_permissions": ["can_become_administrator_in_project"],
				"organization_id": "org-1",
				"project_id": "project-1"
			}`)},
			SourceSchemaVersion: 0,
			SourceTypeName:      "sys11iam_project_team",
		}, &resp)
	}

	require.True(t, resp.Diagnostics.HasError())
}

func TestOrganizationTeamMoveStateSkipsUnknownSourceType(t *testing.T) {
	ctx := context.Background()
	target := &OrganizationTeamResource{}

	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	nullState := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)
	resp := fwresource.MoveStateResponse{
		TargetState: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    nullState,
		},
	}

	for _, mover := range target.MoveState(ctx) {
		mover.StateMover(ctx, fwresource.MoveStateRequest{
			SourceProviderAddress: "registry.terraform.io/syseleven/sys11iam",
			SourceRawState:        &tfprotov6.RawState{JSON: []byte(`{"id":"team-1"}`)},
			SourceSchemaVersion:   0,
			SourceTypeName:        "sys11iam_unknown",
		}, &resp)
	}

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.True(t, resp.TargetState.Raw.Equal(nullState))
}

func TestOrganizationTeamResourceUpgradeStateFromLegacyTeamState(t *testing.T) {
	ctx := context.Background()
	target := &OrganizationTeamResource{}

	upgraders := target.UpgradeState(ctx)
	require.Contains(t, upgraders, int64(0))

	resp := fwresource.UpgradeStateResponse{}
	upgraders[0].StateUpgrader(ctx, fwresource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(`{
			"description": "test team",
			"editable_permissions": ["can_become_project_administrator_in_org"],
			"id": "team-1",
			"name": "testteam",
			"organization_id": "org-1",
			"tags": ["testing2"]
		}`)},
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue)

	var upgraded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &upgraded))

	require.Equal(t, `"org-1"`, string(upgraded["org_id"]))
	require.Equal(t, `"org-1"`, string(upgraded["organization_id"]))
	require.Equal(t, `["can_become_project_administrator_in_org"]`, string(upgraded["organization_permissions"]))
	_, hasLegacyPermissions := upgraded["editable_permissions"]
	require.False(t, hasLegacyPermissions)
}

func TestOrganizationTeamResourceUpgradeStateKeepsExistingOrgIDAndPermissions(t *testing.T) {
	ctx := context.Background()
	target := &OrganizationTeamResource{}

	upgraders := target.UpgradeState(ctx)
	require.Contains(t, upgraders, int64(0))

	resp := fwresource.UpgradeStateResponse{}
	upgraders[0].StateUpgrader(ctx, fwresource.UpgradeStateRequest{
		RawState: &tfprotov6.RawState{JSON: []byte(`{
			"id": "team-1",
			"name": "testteam",
			"org_id": "org-new",
			"organization_id": "org-old",
			"organization_permissions": ["can_invite_members_in_org"],
			"editable_permissions": ["can_become_project_administrator_in_org"]
		}`)},
	}, &resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
	require.NotNil(t, resp.DynamicValue)

	var upgraded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(resp.DynamicValue.JSON, &upgraded))

	require.Equal(t, `"org-new"`, string(upgraded["org_id"]))
	require.Equal(t, `["can_invite_members_in_org"]`, string(upgraded["organization_permissions"]))
	_, hasLegacyPermissions := upgraded["editable_permissions"]
	require.False(t, hasLegacyPermissions)
}
