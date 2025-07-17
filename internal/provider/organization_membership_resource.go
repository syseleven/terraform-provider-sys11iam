package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func (r *OrganizationMembershipResource) createMembershipAttrTypes(membershipType string) map[string]attr.Type {
	switch membershipType {
	case "user":
		return map[string]attr.Type{
			"email":       types.StringType,
			"permissions": types.ListType{ElemType: types.StringType},
		}
	case "service_account":
		return map[string]attr.Type{
			"id":          types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
			"permissions": types.ListType{ElemType: types.StringType},
		}
	default:
		return nil
	}
}

func (r *OrganizationMembershipResource) createMembershipAttrValues(membershipType string, membership *iam.IAMOrganizationMembership) map[string]attr.Value {
	memberAttrTypes := map[string]map[string]attr.Type{
		user: {
			"id":    types.StringType,
			"email": types.StringType,
		},
		serviceAccount: {
			"id":          types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		},
	}

	switch membershipType {
	case user:
		return map[string]attr.Value{
			"affiliation": types.StringValue(membership.Affiliation),
			"permissions": types.ListValueMust(types.StringType, convertSliceToAttrValues(membership.Permissions, func(s string) attr.Value {
				return types.StringValue(s)
			})),
			"id":              types.StringValue(membership.User.ID),
			"membership_type": types.StringValue(membership.MembershipType),
			"organization_id": types.StringValue(membership.Organisation.ID),
			"user": resource_organization_membership.NewUserValueMust(memberAttrTypes[membershipType], map[string]attr.Value{
				"id":          types.StringValue(membership.User.ID),
				"name":        types.StringValue(membership.User.Name),
				"description": types.StringValue(membership.User.Description),
				"created_at":  types.StringValue(membership.User.CreatedAt),
				"updated_at":  types.StringValue(membership.User.UpdatedAt),
			}),
		}
	case serviceAccount:
		return map[string]attr.Value{
			"affiliation": types.StringValue(membership.Affiliation),
			"permissions": types.ListValueMust(types.StringType, convertSliceToAttrValues(membership.Permissions, func(s string) attr.Value {
				return types.StringValue(s)
			})),
			"id":              types.StringValue(membership.ServiceAccount.ID),
			"membership_type": types.StringValue(membership.MembershipType),
			"organization_id": types.StringValue(membership.Organisation.ID),
			"service_account": resource_organization_membership.NewServiceAccountValueMust(memberAttrTypes[membershipType], map[string]attr.Value{
				"id":          types.StringValue(membership.ServiceAccount.ID),
				"name":        types.StringValue(membership.ServiceAccount.Name),
				"description": types.StringValue(membership.ServiceAccount.Description),
				"created_at":  types.StringValue(membership.ServiceAccount.CreatedAt),
				"updated_at":  types.StringValue(membership.ServiceAccount.UpdatedAt),
			}),
		}
	default:
		return nil
	}
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

	var elements []string
	if !data.UserMembership.EditablePermissions.IsNull() || !data.UserMembership.EditablePermissions.IsUnknown() {
		diags := data.UserMembership.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !data.ServiceAccountMembership.EditablePermissions.IsNull() || !data.ServiceAccountMembership.EditablePermissions.IsUnknown() {
		diags := data.ServiceAccountMembership.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Is the user already a member?
	if data.UserMembership.Id.ValueString() != "" {
		var user map[string]string
		diags := data.UserMembership.User.As(ctx, &user, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		email, ok := user["email"]
		if !ok || email == "" {
			resp.Diagnostics.AddError("InvalidUserEmailError",
				"User email is not set or invalid. Please provide a valid user email.")
			return
		}

		org_membership_response, err := r.client.GetOrganizationMembershipByEmail(data.OrganizationId.ValueString(), email)
		if data.UserMembership.Id.ValueString() == "" && err != nil {
			// Is the e-mail at least invited?
			_, err := r.client.GetOrganizationInvitationByEmail(data.OrganizationId.ValueString(), email)
			if err != nil {
				// Invite the e-mail
				_, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), email, elements)
				if err != nil {
					resp.Diagnostics.AddError("", err.Error())
					return
				}
			}
			// The email is invited, but has to be activated manually
			resp.Diagnostics.AddWarning("InvitationNotAcceptedWarning",
				fmt.Sprintf("Can not create OrganizationMembership in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
					data.OrganizationId.ValueString(), email))
			// Save data into Terraform state
			data.UserMembership.Id = types.StringValue("0")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}

		if org_membership_response.User.ID != "" {
			data.MemberId = types.StringValue(org_membership_response.User.ID)
		}
	}

	var affiliation string
	if !data.UserMembership.Affiliation.IsNull() && !data.UserMembership.Affiliation.IsUnknown() {
		affiliation = data.UserMembership.Affiliation.ValueString()
	} else if !data.ServiceAccountMembership.Affiliation.IsNull() && !data.ServiceAccountMembership.Affiliation.IsUnknown() {
		affiliation = data.ServiceAccountMembership.Affiliation.ValueString()
	}

	response, err := r.client.CreateOrUpdateOrganizationMembership(data.OrganizationId.ValueString(), data.MemberId.ValueString(), affiliation, elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	if response.User.ID != "" {
		data.UserMembership = resource_organization_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(user, &response))
		data.MemberId = types.StringValue(response.User.ID)
	} else if response.ServiceAccount.ID != "" {
		data.ServiceAccountMembership = resource_organization_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(serviceAccount, &response))
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
	} else {
		resp.Diagnostics.AddError(
			"Invalid Membership Type",
			"Membership type must be either 'user' or 'service_account'.",
		)
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

	response, err := r.client.GetOrganizationMembership(data.OrganizationId.ValueString(), data.MemberId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	if response.User.ID != "" {
		data.UserMembership = resource_organization_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(user, &response))
		data.MemberId = types.StringValue(response.User.ID)
	} else if response.ServiceAccount.ID != "" {
		data.ServiceAccountMembership = resource_organization_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(serviceAccount, &response))
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
	} else {
		resp.Diagnostics.AddError(
			"Invalid Membership Type",
			"Membership type must be either 'user' or 'service_account'.",
		)
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
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.MemberId)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationMembership resource.")
	var elements []string
	if !data.UserMembership.EditablePermissions.IsNull() || !data.UserMembership.EditablePermissions.IsUnknown() {
		diags := data.UserMembership.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !data.ServiceAccountMembership.EditablePermissions.IsNull() || !data.ServiceAccountMembership.EditablePermissions.IsUnknown() {
		diags := data.ServiceAccountMembership.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if data.UserMembership.Id.ValueString() != "" && data.MemberId.ValueString() == "0" {
		var user map[string]string
		diags := data.UserMembership.User.As(ctx, &user, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		email, ok := user["email"]
		if !ok || email == "" {
			resp.Diagnostics.AddError("InvalidUserEmailError",
				"User email is not set or invalid. Please provide a valid user email.")
			return
		}

		_, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), email, elements)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
		return
	}

	var affiliation string
	if !data.UserMembership.Affiliation.IsNull() && !data.UserMembership.Affiliation.IsUnknown() {
		affiliation = data.UserMembership.Affiliation.ValueString()
	} else if !data.ServiceAccountMembership.Affiliation.IsNull() && !data.ServiceAccountMembership.Affiliation.IsUnknown() {
		affiliation = data.ServiceAccountMembership.Affiliation.ValueString()
	}

	response, err := r.client.UpdateOrganizationMembership(data.OrganizationId.ValueString(), data.MemberId.ValueString(), affiliation, elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	if response.User.ID != "" {
		data.UserMembership = resource_organization_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(user, &response))
		data.MemberId = types.StringValue(response.User.ID)
	} else if response.ServiceAccount.ID != "" {
		data.ServiceAccountMembership = resource_organization_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(serviceAccount, &response))
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
	} else {
		resp.Diagnostics.AddError(
			"Invalid Membership Type",
			"Membership type must be either 'user' or 'service_account'.",
		)
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
	if data.MemberId.ValueString() == "0" {
		var user map[string]string
		diags := data.UserMembership.User.As(ctx, &user, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		email, ok := user["email"]
		if !ok || email == "" {
			resp.Diagnostics.AddError("InvalidUserEmailError",
				"User email is not set or invalid. Please provide a valid user email.")
			return
		}

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
	err := r.client.DeleteOrganizationMembership(data.OrganizationId.ValueString(), data.MemberId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *OrganizationMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || (len(idParts) == 2 && (idParts[0] == "" || idParts[1] == "")) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,member_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationMembership resource.")
	response, err := r.client.GetOrganizationMembership(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_membership.OrganizationMembershipModel
	// Data value setting
	if response.User.ID != "" {
		data.UserMembership = resource_organization_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(user, &response))
		data.MemberId = types.StringValue(response.User.ID)
	} else if response.ServiceAccount.ID != "" {
		data.ServiceAccountMembership = resource_organization_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(serviceAccount, &response))
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
	} else {
		resp.Diagnostics.AddError(
			"Invalid Membership Type",
			"Membership type must be either 'user' or 'service_account'.",
		)
		return
	}
	data.OrganizationId = types.StringValue(response.Organisation.ID)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
