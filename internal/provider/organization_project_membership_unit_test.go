package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_membership"
)

// TestProjectMembershipPlanDecodesPartiallyUnknownValues guards against the
// "Value Conversion Error" that occurred when the framework marked computed
// attributes (for example user.id) as unknown in the plan and the resource
// tried to decode it into a struct that could not represent unknown values.
func TestProjectMembershipPlanDecodesPartiallyUnknownValues(t *testing.T) {
	ctx := context.Background()

	target := &ProjectMembershipResource{}
	var schemaResp fwresource.SchemaResponse
	target.Schema(ctx, fwresource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	user, diags := basetypes.NewObjectValue(resource_organization_project_membership.UserAttrTypes(), map[string]attr.Value{
		"email": types.StringValue("user@example.com"),
		"id":    types.StringUnknown(),
	})
	require.False(t, diags.HasError(), diags)

	userMembership, diags := basetypes.NewObjectValue(resource_organization_project_membership.UserMembershipAttrTypes(), map[string]attr.Value{
		"membership_type": types.StringValue("user"),
		"permissions":     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("can_read")}),
		"user":            user,
	})
	require.False(t, diags.HasError(), diags)

	membership, diags := basetypes.NewObjectValue(resource_organization_project_membership.MembershipAttrTypes(), map[string]attr.Value{
		"user_membership":            userMembership,
		"service_account_membership": basetypes.NewObjectNull(resource_organization_project_membership.ServiceAccountMembershipAttrTypes()),
	})
	require.False(t, diags.HasError(), diags)

	model := resource_organization_project_membership.OrganizationProjectMembershipModel{
		Id:             types.StringUnknown(),
		OrgId:          types.StringValue("org-1"),
		OrganizationId: types.StringValue("org-1"),
		ProjectId:      types.StringValue("project-1"),
		ProjectName:    types.StringUnknown(),
		Membership:     membership,
	}

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	diags = plan.Set(ctx, &model)
	require.False(t, diags.HasError(), diags)

	var decoded resource_organization_project_membership.OrganizationProjectMembershipModel
	diags = plan.Get(ctx, &decoded)
	require.False(t, diags.HasError(), diags)

	require.True(t, decoded.Id.IsUnknown())
	require.True(t, decoded.ProjectName.IsUnknown())

	var membershipValue resource_organization_project_membership.MembershipValue
	diags = decoded.Membership.As(ctx, &membershipValue, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), diags)
	require.True(t, membershipValue.ServiceAccountMembership.IsNull())

	var userMembershipValue resource_organization_project_membership.UserMembershipValue
	diags = membershipValue.UserMembership.As(ctx, &userMembershipValue, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), diags)

	var userValue resource_organization_project_membership.UserValue
	diags = userMembershipValue.User.As(ctx, &userValue, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), diags)
	require.Equal(t, "user@example.com", userValue.Email.ValueString())
	require.True(t, userValue.Id.IsUnknown())
}
