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

func (r *OrganizationTeamMembershipResource) buildData(data resource_organization_team_membership.OrganizationTeamMembershipModel, response *iam.IAMOrganizationTeamMembership, teamPermissions []string) resource_organization_team_membership.OrganizationTeamMembershipModel {
	if len(teamPermissions) <= 0 {
		teamPermissions = response.TeamPermissions
	}

	data.TeamID = types.StringValue(response.Team.ID)
	data.TeamName = types.StringValue(response.Team.Name)
	data.OrganizationId = types.StringValue(response.Organisation.ID)
	data.MembershipType = types.StringValue(response.MembershipType)

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

	if response.User.ID != "" {
		data.Id = types.StringValue(response.User.ID)

		userMembershipValue := basetypes.NewObjectValueMust(userMembershipAttrTypes, map[string]attr.Value{
			"user": basetypes.NewObjectValueMust(map[string]attr.Type{
				"email": types.StringType,
			}, map[string]attr.Value{
				"email": types.StringValue(response.User.Email),
			}),
			"team_permissions": types.ListValueMust(types.StringType, teamPermissionAttrValues),
		})

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_team_membership": basetypes.NewObjectNull(serviceAccountMembershipAttrTypes),
			"user_team_membership":            userMembershipValue,
		})
	} else if response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(response.ServiceAccount.ID)

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

	// Access email through the Attributes map
	if userTeamMembershipAttr, ok := membership.Attributes()["user_team_membership"]; ok {
		if !userTeamMembershipAttr.IsNull() && !userTeamMembershipAttr.IsUnknown() {
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
						org_membership_response, err := r.client.GetOrganizationMembershipByEmail(ctx, data.OrganizationId.ValueString(), email)
						if err != nil {
							resp.Diagnostics.AddError("", err.Error())
							return
						}

						if data.Id.IsNull() || data.Id.IsUnknown() {
							data.Id = types.StringValue(org_membership_response.User.ID)
						}
					}
				}
			}

		}
	}

	response, err := r.client.CreateOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	teamPermissions, diag := r.getTeamPermissions(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	if len(teamPermissions) > 0 {
		_, err := r.client.GrantOrganizationTeamMemberPermissions(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString(), teamPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	// Data value setting
	data = r.buildData(data, &response, teamPermissions)

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
	response, err := r.client.GetOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	teamPermissions, diag := r.getTeamPermissions(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	teamPermissions = MergeSlices(teamPermissions, response.TeamPermissions, func(p string) string {
		return p
	})

	// Data value setting
	data = r.buildData(data, &response, teamPermissions)

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

	response, err := r.client.UpdateOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "409") {
			response, err = r.client.GetOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
		} else {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	if len(teamPermissions) > 0 {
		teamPermissionsResponse, err := r.client.UpdateOrganizationTeamMemberPermissions(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString(), teamPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}

		teamPermissions = MergeSlices(teamPermissions, teamPermissionsResponse.UpdatedPermissions, func(p string) string {
			return p
		})
	}

	// Data value setting
	data = r.buildData(data, &response, teamPermissions)

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

	// Delete API call logic
	// Remove a member (user, service account) from a team in an organization, or only revoke the passed-in permissions if any
	tflog.Info(ctx, "Deleting OrganizationTeamMembership resource.")

	teamPermissions, diag := r.getTeamPermissions(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	if len(teamPermissions) > 0 {
		_, err := r.client.RevokeOrganizationTeamMemberPermissions(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString(), teamPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	err := r.client.DeleteOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamID.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
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

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationTeamMembership resource.")
	response, err := r.client.GetOrganizationTeamMembership(idParts[0], idParts[1], idParts[2])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	var data resource_organization_team_membership.OrganizationTeamMembershipModel
	data = r.buildData(data, &response, nil)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
