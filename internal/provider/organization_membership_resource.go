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
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_membership"
)

var _ resource.Resource = (*OrganizationMembershipResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationMembershipResource)(nil)
var _ resource.ResourceWithUpgradeState = (*OrganizationMembershipResource)(nil)

func NewOrganizationMembershipResource() resource.Resource {
	return &OrganizationMembershipResource{}
}

type OrganizationMembershipResource struct {
	client *iam.Client
}

func (r *OrganizationMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_membership"
}

func (r *OrganizationMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_membership.OrganizationMembershipResourceSchema(ctx)
}

func (r *OrganizationMembershipResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: compat.OrgIdStateUpgrader(),
	}
}

func (r *OrganizationMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func getServiceAccountMembershipAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.StringType,
		"name":            types.StringType,
		"affiliation":     types.StringType,
		"membership_type": types.StringType,
		"permissions":     types.ListType{ElemType: types.StringType},
	}
}

func getUserMembershipAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.StringType,
		"email":           types.StringType,
		"affiliation":     types.StringType,
		"membership_type": types.StringType,
		"permissions":     types.ListType{ElemType: types.StringType},
	}
}

func getMembershipAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"service_account_membership": basetypes.ObjectType{AttrTypes: getServiceAccountMembershipAttrTypes()},
		"user_membership":            basetypes.ObjectType{AttrTypes: getUserMembershipAttrTypes()},
	}
}

// buildDataFromV3 populates the Terraform model from the v3 membership list entry.
func (r *OrganizationMembershipResource) buildDataFromV3(ctx context.Context, data *resource_organization_membership.OrganizationMembershipModel, member *iam.IAMOrgMembershipV3) diag.Diagnostics {
	data.Id = types.StringValue(member.ID)

	permissions := iam.FilterActiveDirectPermissions(member.Permissions)
	permissionsAttrs := convertSliceToAttrValues(permissions, func(s string) attr.Value {
		return types.StringValue(s)
	})

	serviceAccountMembershipAttrTypes := getServiceAccountMembershipAttrTypes()
	userMembershipAttrTypes := getUserMembershipAttrTypes()
	membershipAttrTypes := getMembershipAttrTypes()

	var membership basetypes.ObjectValue
	diags := data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return diags
	}

	affiliation := ""
	if member.OrgAffiliation != nil {
		affiliation = *member.OrgAffiliation
	}

	if member.Type == "user" {
		var userMembership resource_organization_membership.UserMembershipValue
		diags = membership.Attributes()["user_membership"].(basetypes.ObjectValue).As(ctx, &userMembership, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		userMembershipPermissionsAttrs := MergeSlices(userMembership.Permissions.Elements(), permissionsAttrs, func(p attr.Value) attr.Value {
			return p
		})

		userMembershipValue := basetypes.NewObjectValueMust(userMembershipAttrTypes, map[string]attr.Value{
			"id":              types.StringValue(member.ID),
			"email":           types.StringValue(member.DisplayName),
			"affiliation":     types.StringValue(affiliation),
			"membership_type": types.StringValue(member.Type),
			"permissions":     types.ListValueMust(types.StringType, userMembershipPermissionsAttrs),
		})

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_membership": basetypes.NewObjectNull(serviceAccountMembershipAttrTypes),
			"user_membership":            userMembershipValue,
		})
	} else if member.Type == "service_account" {
		var serviceAccountMembership resource_organization_membership.ServiceAccountMembershipValue
		diags = membership.Attributes()["service_account_membership"].(basetypes.ObjectValue).As(ctx, &serviceAccountMembership, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		serviceAccountMembershipPermissionsAttrs := MergeSlices(serviceAccountMembership.Permissions.Elements(), permissionsAttrs, func(p attr.Value) attr.Value {
			return p
		})

		serviceAccountMembershipValue := basetypes.NewObjectValueMust(serviceAccountMembershipAttrTypes, map[string]attr.Value{
			"id":              types.StringValue(member.ID),
			"name":            types.StringValue(member.DisplayName),
			"affiliation":     types.StringValue(affiliation),
			"membership_type": types.StringValue(member.Type),
			"permissions":     types.ListValueMust(types.StringType, serviceAccountMembershipPermissionsAttrs),
		})

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_membership": serviceAccountMembershipValue,
			"user_membership":            basetypes.NewObjectNull(userMembershipAttrTypes),
		})
	}

	return diags
}

// writeOrgPermissionsAndAffiliation writes permissions and affiliation via the v3 permission endpoints.
func (r *OrganizationMembershipResource) writeOrgPermissionsAndAffiliation(org_id string, member_id string, membershipType string, permissions []string, affiliation string) error {
	if membershipType == "user" {
		_, err := r.client.PutUserOrgPermissions(org_id, member_id, permissions)
		if err != nil {
			return fmt.Errorf("failed to put user org permissions: %w", err)
		}
		_, err = r.client.PutUserOrgAffiliation(org_id, member_id, affiliation)
		if err != nil {
			return fmt.Errorf("failed to put user org affiliation: %w", err)
		}
	} else if membershipType == "service_account" {
		_, err := r.client.PutServiceAccountOrgPermissions(org_id, member_id, permissions)
		if err != nil {
			return fmt.Errorf("failed to put service account org permissions: %w", err)
		}
		_, err = r.client.PutServiceAccountOrgAffiliation(org_id, member_id, affiliation)
		if err != nil {
			return fmt.Errorf("failed to put service account org affiliation: %w", err)
		}
	}
	return nil
}

func (r *OrganizationMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_membership.OrganizationMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationMembership resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", orgId.ValueString()))
	// Is the organization active?
	org_response, err := r.client.GetOrganization(orgId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create OrganizationMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact the SysEleven GmbH Sales Team <sales@syseleven.de>.\n This can also be done via https://dashboard.syseleven.de/dashboard",
				orgId.ValueString()))
		return
	}

	var permissions []string
	var membership basetypes.ObjectValue
	diag := data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{})
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var affiliation string
	var membershipType string

	if userMembershipAttr, ok := membership.Attributes()["user_membership"]; ok && !userMembershipAttr.IsNull() && !userMembershipAttr.IsUnknown() {
		userMembershipObj := userMembershipAttr.(basetypes.ObjectValue)

		var userMembership resource_organization_membership.UserMembershipValue
		diag = userMembershipObj.As(ctx, &userMembership, basetypes.ObjectAsOptions{})
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		diags := userMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		affiliation = userMembership.Affiliation.ValueString()
		membershipType = userMembership.MembershipType.ValueString()

		// Is the user already a member?
		email := userMembership.Email.ValueString()

		org_membership_response, err := r.client.GetOrgMembershipV3ByEmail(orgId.ValueString(), email)
		if userMembership.Id.IsNull() || userMembership.Id.IsUnknown() && err != nil {
			// Is the e-mail at least invited?
			org_invitation_response, err := r.client.GetOrganizationInvitationByEmail(orgId.ValueString(), email)
			if org_invitation_response.ID == "" || err != nil {
				// Invite the e-mail
				invitationResponse, err := r.client.CreateOrganizationInvitation(orgId.ValueString(), email, permissions)
				if err != nil {
					resp.Diagnostics.AddError("", err.Error())
					return
				}
				data.Id = types.StringValue(invitationResponse.ID)
			}
			// The email is invited, but has to be activated manually
			resp.Diagnostics.AddError("InvitationNotAccepted",
				fmt.Sprintf("Can not create OrganizationMembership in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
					orgId.ValueString(), email))

			// data value setting
			if data.Id.IsNull() || data.Id.IsUnknown() {
				data.Id = types.StringValue(org_invitation_response.ID)
			}

			serviceAccountMembershipAttrTypes := getServiceAccountMembershipAttrTypes()
			userMembershipAttrTypes := getUserMembershipAttrTypes()
			membershipAttrTypes := getMembershipAttrTypes()

			userMembershipValue := basetypes.NewObjectValueMust(userMembershipAttrTypes, map[string]attr.Value{
				"id":              data.Id,
				"email":           userMembership.Email,
				"affiliation":     userMembership.Affiliation,
				"membership_type": userMembership.MembershipType,
				"permissions":     userMembership.Permissions,
			})

			data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
				"service_account_membership": basetypes.NewObjectNull(serviceAccountMembershipAttrTypes),
				"user_membership":            userMembershipValue,
			})

			data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
		data.Id = types.StringValue(org_membership_response.ID)
	} else if serviceAccountMembershipAttr, ok := membership.Attributes()["service_account_membership"]; ok && !serviceAccountMembershipAttr.IsNull() && !serviceAccountMembershipAttr.IsUnknown() {
		serviceAccountMembershipObj := serviceAccountMembershipAttr.(basetypes.ObjectValue)

		var serviceAccountMembership resource_organization_membership.ServiceAccountMembershipValue
		diag = serviceAccountMembershipObj.As(ctx, &serviceAccountMembership, basetypes.ObjectAsOptions{})
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		diags := serviceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		affiliation = serviceAccountMembership.Affiliation.ValueString()
		membershipType = serviceAccountMembership.MembershipType.ValueString()

		if data.Id.IsNull() || data.Id.IsUnknown() {
			if serviceAccountMembership.Id.IsNull() || serviceAccountMembership.Id.IsUnknown() {
				resp.Diagnostics.AddError("", "Service account membership id is required")
				return
			}
			data.Id = serviceAccountMembership.Id
		}
	}

	// Write permissions and affiliation via v3 endpoints
	err = r.writeOrgPermissionsAndAffiliation(orgId.ValueString(), data.Id.ValueString(), membershipType, permissions, affiliation)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Read back the membership to populate state
	member, err := r.client.GetOrgMembershipV3(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	diag = r.buildDataFromV3(ctx, &data, &member)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_membership.OrganizationMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationMembership resource.")

	if data.Id.IsNull() || data.Id.IsUnknown() {
		resp.Diagnostics.AddError("", "Organization membership id is required")
		return
	}

	member, err := r.client.GetOrgMembershipV3(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	diag := r.buildDataFromV3(ctx, &data, &member)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_membership.OrganizationMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationMembership resource.")

	var permissions []string
	var membership basetypes.ObjectValue
	diag := data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	var affiliation string
	var membershipType string

	if userMembershipAttr, ok := membership.Attributes()["user_membership"]; ok && !userMembershipAttr.IsNull() && !userMembershipAttr.IsUnknown() {
		userMembershipObj := userMembershipAttr.(basetypes.ObjectValue)

		var userMembership resource_organization_membership.UserMembershipValue
		diag = userMembershipObj.As(ctx, &userMembership, basetypes.ObjectAsOptions{})
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		diags := userMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		affiliation = userMembership.Affiliation.ValueString()
		membershipType = userMembership.MembershipType.ValueString()

		if userMembership.Id.ValueString() != "" && data.Id.ValueString() == "0" {
			email := userMembership.Email.ValueString()

			_, err := r.client.CreateOrganizationInvitation(orgId.ValueString(), email, permissions)
			if err != nil {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
			return
		}
	} else if serviceAccountMembershipAttr, ok := membership.Attributes()["service_account_membership"]; ok && !serviceAccountMembershipAttr.IsNull() && !serviceAccountMembershipAttr.IsUnknown() {
		serviceAccountMembershipObj := serviceAccountMembershipAttr.(basetypes.ObjectValue)

		var serviceAccountMembership resource_organization_membership.ServiceAccountMembershipValue
		diag = serviceAccountMembershipObj.As(ctx, &serviceAccountMembership, basetypes.ObjectAsOptions{})
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		diags := serviceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		affiliation = serviceAccountMembership.Affiliation.ValueString()
		membershipType = serviceAccountMembership.MembershipType.ValueString()
	}

	// Write permissions and affiliation via v3 endpoints
	err := r.writeOrgPermissionsAndAffiliation(orgId.ValueString(), data.Id.ValueString(), membershipType, permissions, affiliation)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Read back the membership to populate state
	member, err := r.client.GetOrgMembershipV3(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	diag = r.buildDataFromV3(ctx, &data, &member)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_membership.OrganizationMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Delete API call logic
	tflog.Info(ctx, "Deleting OrganizationMembership resource.")
	if data.Id.ValueString() == "0" {
		var membership basetypes.ObjectValue
		diag := data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		if userMembershipAttr, ok := membership.Attributes()["user_membership"]; ok && !userMembershipAttr.IsNull() && !userMembershipAttr.IsUnknown() {
			userMembershipObj := userMembershipAttr.(basetypes.ObjectValue)

			var userMembership resource_organization_membership.UserMembershipValue
			diag = userMembershipObj.As(ctx, &userMembership, basetypes.ObjectAsOptions{})
			if diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}

			email := userMembership.Email.ValueString()

			_, err := r.client.GetOrganizationInvitationByEmail(orgId.ValueString(), email)
			if err != nil {
				return
			}
			err = r.client.DeleteOrganizationInvitation(orgId.ValueString(), email)
			if err != nil {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
		}
		return
	}

	// Determine membership type from state to call the correct endpoint
	var membership basetypes.ObjectValue
	diag := data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	orgIdStr := orgId.ValueString()
	memberId := data.Id.ValueString()

	if userMembershipAttr, ok := membership.Attributes()["user_membership"]; ok && !userMembershipAttr.IsNull() && !userMembershipAttr.IsUnknown() {
		_, err := r.client.PutUserOrgPermissions(orgIdStr, memberId, []string{})
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	} else if serviceAccountMembershipAttr, ok := membership.Attributes()["service_account_membership"]; ok && !serviceAccountMembershipAttr.IsNull() && !serviceAccountMembershipAttr.IsUnknown() {
		_, err := r.client.PutServiceAccountOrgPermissions(orgIdStr, memberId, []string{})
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}
}

func (r *OrganizationMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	tflog.Info(ctx, "Importing OrganizationMembership resource.")
	if len(idParts) != 2 || (len(idParts) == 2 && (idParts[0] == "" || idParts[1] == "")) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,member_id. Got: %q", req.ID),
		)
		return
	}

	// Read membership via v3 list endpoint
	member, err := r.client.GetOrgMembershipV3(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_membership.OrganizationMembershipModel
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(idParts[0])

	// Initialize the membership structure for buildDataFromV3
	serviceAccountMembershipAttrTypes := getServiceAccountMembershipAttrTypes()
	userMembershipAttrTypes := getUserMembershipAttrTypes()
	membershipAttrTypes := getMembershipAttrTypes()

	if member.Type == "user" {
		userMembershipValue := basetypes.NewObjectValueMust(userMembershipAttrTypes, map[string]attr.Value{
			"id":              types.StringValue(member.ID),
			"email":           types.StringValue(member.DisplayName),
			"affiliation":     types.StringValue(""),
			"membership_type": types.StringValue(member.Type),
			"permissions":     types.ListValueMust(types.StringType, []attr.Value{}),
		})
		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_membership": basetypes.NewObjectNull(serviceAccountMembershipAttrTypes),
			"user_membership":            userMembershipValue,
		})
	} else {
		serviceAccountMembershipValue := basetypes.NewObjectValueMust(serviceAccountMembershipAttrTypes, map[string]attr.Value{
			"id":              types.StringValue(member.ID),
			"name":            types.StringValue(member.DisplayName),
			"affiliation":     types.StringValue(""),
			"membership_type": types.StringValue(member.Type),
			"permissions":     types.ListValueMust(types.StringType, []attr.Value{}),
		})
		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_membership": serviceAccountMembershipValue,
			"user_membership":            basetypes.NewObjectNull(userMembershipAttrTypes),
		})
	}

	// Data value setting
	diag := r.buildDataFromV3(ctx, &data, &member)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
