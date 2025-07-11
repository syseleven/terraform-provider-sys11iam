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
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_membership"
)

var _ resource.Resource = (*OrganizationMembershipResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationMembershipResource)(nil)

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

func (r *OrganizationMembershipResource) buildData(ctx context.Context, data *resource_organization_membership.OrganizationMembershipModel, response *iam.IAMOrganizationMembership) diag.Diagnostics {
	data.Id = types.StringValue(response.ID)

	permissionsAttrs := convertSliceToAttrValues(response.Permissions, func(s string) attr.Value {
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

	if response.User.ID != "" {
		var userMembership resource_organization_membership.UserMembershipValue
		diags = membership.Attributes()["user_membership"].(basetypes.ObjectValue).As(ctx, &userMembership, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		userMembershipPermissionsAttrs := MergeSlices(userMembership.Permissions.Elements(), permissionsAttrs, func(p attr.Value) attr.Value {
			return p
		})

		userMembershipValue := basetypes.NewObjectValueMust(userMembershipAttrTypes, map[string]attr.Value{
			"id":              types.StringValue(response.User.ID),
			"email":           types.StringValue(response.User.Email),
			"affiliation":     types.StringValue(response.Affiliation),
			"membership_type": types.StringValue(response.MembershipType),
			"permissions":     types.ListValueMust(types.StringType, userMembershipPermissionsAttrs),
		})

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_membership": basetypes.NewObjectNull(serviceAccountMembershipAttrTypes),
			"user_membership":            userMembershipValue,
		})
	} else if response.ServiceAccount.ID != "" {
		var serviceAccountMembership resource_organization_membership.ServiceAccountMembershipValue
		diags = membership.Attributes()["service_account_membership"].(basetypes.ObjectValue).As(ctx, &serviceAccountMembership, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return diags
		}

		serviceAccountMembershipPermissionsAttrs := MergeSlices(serviceAccountMembership.Permissions.Elements(), permissionsAttrs, func(p attr.Value) attr.Value {
			return p
		})

		serviceAccountMembershipValue := basetypes.NewObjectValueMust(serviceAccountMembershipAttrTypes, map[string]attr.Value{
			"id":              types.StringValue(response.ServiceAccount.ID),
			"name":            types.StringValue(response.ServiceAccount.Name),
			"affiliation":     types.StringValue(response.Affiliation),
			"membership_type": types.StringValue(response.MembershipType),
			"permissions":     types.ListValueMust(types.StringType, serviceAccountMembershipPermissionsAttrs),
		})

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"service_account_membership": serviceAccountMembershipValue,
			"user_membership":            basetypes.NewObjectNull(userMembershipAttrTypes),
		})
	}

	return diags
}

func (r *OrganizationMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_membership.OrganizationMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationMembership resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))
	// Is the organization active?
	org_response, err := r.client.GetOrganization(data.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create OrganizationMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact the SysEleven GmbH Sales Team <sales@syseleven.de>.\n This can also be done via https://dashboard.syseleven.de/dashboard",
				data.OrganizationId.ValueString()))
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

		org_membership_response, err := r.client.GetOrganizationMembershipByEmail(ctx, data.OrganizationId.ValueString(), email)
		if userMembership.Id.IsNull() || userMembership.Id.IsUnknown() && err != nil {
			// Is the e-mail at least invited?
			org_invitation_response, err := r.client.GetOrganizationInvitationByEmail(data.OrganizationId.ValueString(), email)
			if org_invitation_response.ID == "" || err != nil {
				// Invite the e-mail
				invitationResponse, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), email, permissions)
				if err != nil {
					resp.Diagnostics.AddError("", err.Error())
					return
				}
				data.Id = types.StringValue(invitationResponse.ID)
			}
			// The email is invited, but has to be activated manually
			resp.Diagnostics.AddError("InvitationNotAccepted",
				fmt.Sprintf("Can not create OrganizationMembership in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
					data.OrganizationId.ValueString(), email))

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

	response, err := r.client.CreateOrUpdateOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString(), affiliation, membershipType, permissions)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	diag = r.buildData(ctx, &data, &response)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.OrganizationId = types.StringValue(response.Organisation.ID)

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

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationMembership resource.")

	if data.Id.IsNull() || data.Id.IsUnknown() {
		resp.Diagnostics.AddError("", "Organization membership id is required")
		return
	}

	response, err := r.client.GetOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	diag := r.buildData(ctx, &data, &response)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	data.OrganizationId = types.StringValue(response.Organisation.ID)

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

			_, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), email, permissions)
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

	response, err := r.client.UpdateOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString(), affiliation, membershipType, permissions)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	diag = r.buildData(ctx, &data, &response)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	data.OrganizationId = types.StringValue(response.Organisation.ID)

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

			_, err := r.client.GetOrganizationInvitationByEmail(data.OrganizationId.ValueString(), email)
			if err != nil {
				return
			}
			err = r.client.DeleteOrganizationInvitation(data.OrganizationId.ValueString(), email)
			if err != nil {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
		}
	}
	err := r.client.DeleteOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
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

	// Read API call logic
	response, err := r.client.GetOrganizationMembership(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_membership.OrganizationMembershipModel
	// Data value setting
	diag := r.buildData(ctx, &data, &response)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	data.OrganizationId = types.StringValue(response.Organisation.ID)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
