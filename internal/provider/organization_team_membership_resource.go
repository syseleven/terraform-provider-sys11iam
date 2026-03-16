package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_team_membership"
)

var _ resource.Resource = (*OrganizationTeamMembershipResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationTeamMembershipResource)(nil)

func NewOrganizationTeamMembershipResource() resource.Resource {
	return &OrganizationTeamMembershipResource{}
}

type OrganizationTeamMembershipResource struct {
	client *iam.Client
}

func (r *OrganizationTeamMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_team_membership"
}

func (r *OrganizationTeamMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_team_membership.OrganizationTeamMembershipResourceSchema(ctx)
}

func (r *OrganizationTeamMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*iam.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *iam.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// buildDataFromV3 populates the Terraform model from the v3 team membership list entry.
func (r *OrganizationTeamMembershipResource) buildDataFromV3(data resource_organization_team_membership.OrganizationTeamMembershipModel, member *iam.IAMTeamMembershipV3, teamPermissions []string) resource_organization_team_membership.OrganizationTeamMembershipModel {
	data.MembershipType = types.StringValue(member.Type)

	serviceAccountMembershipAttrTypes := map[string]attr.Type{
		"team_permissions": types.ListType{ElemType: types.StringType},
	}

	userMembershipAttrTypes := map[string]attr.Type{
		"user": basetypes.ObjectType{
			AttrTypes: map[string]attr.Type{
				"email": types.StringType,
			},
		},
		"team_permissions": types.ListType{ElemType: types.StringType},
	}

	membershipAttrTypes := map[string]attr.Type{
		"service_account_team_membership": basetypes.ObjectType{AttrTypes: serviceAccountMembershipAttrTypes},
		"user_team_membership":            basetypes.ObjectType{AttrTypes: userMembershipAttrTypes},
	}

	teamPermissionAttrValues := convertSliceToAttrValues(teamPermissions, func(permission string) attr.Value {
		return types.StringValue(permission)
	})

	if member.Type == "user" {
		data.Id = types.StringValue(member.ID)

		userMembershipValue := basetypes.NewObjectValueMust(userMembershipAttrTypes, map[string]attr.Value{
			"user": basetypes.NewObjectValueMust(map[string]attr.Type{
				"email": types.StringType,
			}, map[string]attr.Value{
				"email": types.StringValue(member.DisplayName),
			}),
			"team_permissions": types.ListValueMust(types.StringType, teamPermissionAttrValues),
		})

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_team_membership": basetypes.NewObjectNull(serviceAccountMembershipAttrTypes),
			"user_team_membership":            userMembershipValue,
		})
	} else if member.Type == "service_account" {
		data.Id = types.StringValue(member.ID)

		serviceAccountMembershipValue := basetypes.NewObjectValueMust(serviceAccountMembershipAttrTypes, map[string]attr.Value{
			"team_permissions": types.ListValueMust(types.StringType, teamPermissionAttrValues),
		})

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_team_membership": serviceAccountMembershipValue,
			"user_team_membership":            basetypes.NewObjectNull(userMembershipAttrTypes),
		})
	}

	return data
}

func (r *OrganizationTeamMembershipResource) getTeamPermissions(ctx context.Context, data resource_organization_team_membership.OrganizationTeamMembershipModel) ([]string, diag.Diagnostics) {
	var teamPermissions []string
	var membership basetypes.ObjectValue

	diag := data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{})
	if diag.HasError() {
		return nil, diag
	}

	if data.MembershipType.ValueString() == "user" {
		if userTeamMembershipAttr, ok := membership.Attributes()["user_team_membership"]; ok {
			if !userTeamMembershipAttr.IsNull() && !userTeamMembershipAttr.IsUnknown() {
				userTeamMembershipObj := userTeamMembershipAttr.(basetypes.ObjectValue)

				var userMembership resource_organization_team_membership.UserTeamMembershipValue
				diag = userTeamMembershipObj.As(ctx, &userMembership, basetypes.ObjectAsOptions{})
				if diag.HasError() {
					return nil, diag
				}

				diag = userMembership.TeamPermissions.ElementsAs(ctx, &teamPermissions, false)
				if diag.HasError() {
					return nil, diag
				}
			}
		}
	} else if data.MembershipType.ValueString() == "service_account" {
		if serviceAccountMembershipAttr, ok := membership.Attributes()["service_account_team_membership"]; ok {
			if !serviceAccountMembershipAttr.IsNull() && !serviceAccountMembershipAttr.IsUnknown() {
				serviceAccountMembershipObj := serviceAccountMembershipAttr.(basetypes.ObjectValue)

				var serviceAccountMembership resource_organization_team_membership.ServiceAccountMembershipValue
				diag = serviceAccountMembershipObj.As(ctx, &serviceAccountMembership, basetypes.ObjectAsOptions{})
				if diag.HasError() {
					return nil, diag
				}

				diag = serviceAccountMembership.TeamPermissions.ElementsAs(ctx, &teamPermissions, false)
				if diag.HasError() {
					return nil, diag
				}
			}
		}
	}
	return teamPermissions, diag
}

func (r *OrganizationTeamMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_team_membership.OrganizationTeamMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// For Optional+Computed attributes on a new resource, the plan may mark id
	// as unknown even when the user supplies a value in the configuration.
	// Fall back to the config value so the member ID is available for API calls.
	if data.Id.IsNull() || data.Id.IsUnknown() || data.Id.ValueString() == "" {
		var configId types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &configId)...)
		if !resp.Diagnostics.HasError() && !configId.IsNull() && !configId.IsUnknown() && configId.ValueString() != "" {
			data.Id = configId
		}
	}

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationTeamMembership resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))

	var membership basetypes.ObjectValue
	var diag diag.Diagnostics
	diag = data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var membershipType string

	// Resolve member ID from email if needed
	if userTeamMembershipAttr, ok := membership.Attributes()["user_team_membership"]; ok {
		if !userTeamMembershipAttr.IsNull() && !userTeamMembershipAttr.IsUnknown() {
			membershipType = "user"
			userTeamMembershipObj := userTeamMembershipAttr.(basetypes.ObjectValue)

			var userMembership resource_organization_team_membership.UserTeamMembershipValue
			diag = userTeamMembershipObj.As(ctx, &userMembership, basetypes.ObjectAsOptions{})
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}

			var user basetypes.ObjectValue
			diag = userMembership.User.As(ctx, &user, basetypes.ObjectAsOptions{})
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}

			if emailAttr, ok := user.Attributes()["email"]; ok {
				if !emailAttr.IsNull() && !emailAttr.IsUnknown() {
					email := emailAttr.(basetypes.StringValue).ValueString()
					if email != "" {
						orgMember, err := r.client.GetOrgMembershipV3ByEmail(data.OrganizationId.ValueString(), email)
						if err != nil {
							resp.Diagnostics.AddError("", err.Error())
							return
						}

						if data.Id.IsNull() || data.Id.IsUnknown() || data.Id.ValueString() == "" {
							data.Id = types.StringValue(orgMember.ID)
						}
					}
				}
			}
		}
	}

	// If the user_team_membership block was not recognized above (e.g. because it
	// was null/unknown) but a membership_type is given in config, use that.
	if membershipType == "" {
		var configMembershipType types.String
		_ = req.Config.GetAttribute(ctx, path.Root("membership_type"), &configMembershipType)
		if !configMembershipType.IsNull() && !configMembershipType.IsUnknown() && configMembershipType.ValueString() != "" {
			membershipType = configMembershipType.ValueString()
		}
	}

	// If the member ID is still unresolved and the membership is for a user,
	// try reading the email directly from the configuration. The plan may mark
	// deeply nested Optional+Computed attributes as unknown during Create,
	// which prevents the plan-based extraction above from finding the email.
	if (data.Id.IsNull() || data.Id.IsUnknown() || data.Id.ValueString() == "") && membershipType == "user" {
		emailPath := path.Root("membership").AtName("user_team_membership").AtName("user").AtName("email")
		var configEmail types.String
		configDiags := req.Config.GetAttribute(ctx, emailPath, &configEmail)
		if !configDiags.HasError() && !configEmail.IsNull() && !configEmail.IsUnknown() && configEmail.ValueString() != "" {
			tflog.Info(ctx, fmt.Sprintf("Resolving member ID from config email: %s", configEmail.ValueString()))
			orgMember, err := r.client.GetOrgMembershipV3ByEmail(data.OrganizationId.ValueString(), configEmail.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
			data.Id = types.StringValue(orgMember.ID)
		}
	}

	if membershipType == "" {
		membershipType = "service_account"
	}

	teamPermissions, diag := r.getTeamPermissions(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Add member to team first (v3 requires membership before setting permissions)
	orgId := data.OrganizationId.ValueString()
	teamId := data.TeamID.ValueString()
	memberId := data.Id.ValueString()

	if memberId == "" {
		resp.Diagnostics.AddError("",
			fmt.Sprintf("member ID is empty: provide a valid 'id' or 'email' so the member can be resolved "+
				"(id null=%v unknown=%v, membershipType=%q)",
				data.Id.IsNull(), data.Id.IsUnknown(), membershipType))
		return
	}

	if membershipType == "user" {
		err := r.client.AddUserToTeam(orgId, teamId, memberId)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	} else {
		err := r.client.AddServiceAccountToTeam(orgId, teamId, memberId)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	// Write team permissions via v3 endpoints
	if membershipType == "user" {
		_, err := r.client.PutUserTeamPermissions(orgId, teamId, memberId, teamPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	} else {
		_, err := r.client.PutServiceAccountTeamPermissions(orgId, teamId, memberId, teamPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	// Read back to populate state
	member, err := r.client.GetTeamMembershipV3(orgId, teamId, memberId)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	permissions := iam.FilterActiveDirectPermissions(member.Permissions)
	data.TeamID = types.StringValue(teamId)
	data.OrganizationId = types.StringValue(orgId)

	// Fetch team name
	team, err := r.client.GetOrganizationTeam(orgId, teamId)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	data.TeamName = types.StringValue(team.Name)

	// Data value setting
	data = r.buildDataFromV3(data, &member, permissions)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_team_membership.OrganizationTeamMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationTeamMembership resource.")
	member, err := r.client.GetTeamMembershipV3(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	permissions := iam.FilterActiveDirectPermissions(member.Permissions)

	teamPermissions, diag := r.getTeamPermissions(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	teamPermissions = MergeSlices(teamPermissions, permissions, func(p string) string {
		return p
	})

	// Data value setting
	data = r.buildDataFromV3(data, &member, teamPermissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_team_membership.OrganizationTeamMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationTeamMembership resource.")

	teamPermissions, diag := r.getTeamPermissions(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	orgId := data.OrganizationId.ValueString()
	teamId := data.TeamID.ValueString()
	memberId := data.Id.ValueString()

	// Write team permissions via v3 endpoints
	membershipType := data.MembershipType.ValueString()
	if membershipType == "user" {
		_, err := r.client.PutUserTeamPermissions(orgId, teamId, memberId, teamPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	} else {
		_, err := r.client.PutServiceAccountTeamPermissions(orgId, teamId, memberId, teamPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	// Read back to populate state
	member, err := r.client.GetTeamMembershipV3(orgId, teamId, memberId)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	permissions := iam.FilterActiveDirectPermissions(member.Permissions)

	// Data value setting
	data = r.buildDataFromV3(data, &member, permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_team_membership.OrganizationTeamMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic — remove member from team
	tflog.Info(ctx, "Deleting OrganizationTeamMembership resource.")

	orgId := data.OrganizationId.ValueString()
	teamId := data.TeamID.ValueString()
	memberId := data.Id.ValueString()

	membershipType := data.MembershipType.ValueString()
	if membershipType == "user" {
		err := r.client.RemoveUserFromTeam(orgId, teamId, memberId)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	} else {
		err := r.client.RemoveServiceAccountFromTeam(orgId, teamId, memberId)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}
}

func (r *OrganizationTeamMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,team_id,id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic via v3
	tflog.Info(ctx, "Reading OrganizationTeamMembership resource.")
	member, err := r.client.GetTeamMembershipV3(idParts[0], idParts[1], idParts[2])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	permissions := iam.FilterActiveDirectPermissions(member.Permissions)

	// Fetch team name
	team, err := r.client.GetOrganizationTeam(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	var data resource_organization_team_membership.OrganizationTeamMembershipModel
	data.TeamID = types.StringValue(idParts[1])
	data.TeamName = types.StringValue(team.Name)
	data.OrganizationId = types.StringValue(idParts[0])
	data = r.buildDataFromV3(data, &member, permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
