package provider

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_membership"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_s3_user"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_s3_user_key"
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
	require.NotNil(t, data.Membership)
	require.NotNil(t, data.Membership.UserMembership)
	require.Nil(t, data.Membership.ServiceAccountMembership)
	require.Equal(t, "user", data.Membership.UserMembership.MembershipType.ValueString())
	require.Equal(t, "user@example.com", data.Membership.UserMembership.User.Email.ValueString())
	require.Equal(t, "user-1", data.Membership.UserMembership.User.Id.ValueString())

	var permissions []string
	diags = data.Membership.UserMembership.Permissions.ElementsAs(context.Background(), &permissions, false)
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
