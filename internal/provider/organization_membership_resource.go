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
	if !data.User.EditablePermissions.IsNull() || !data.User.EditablePermissions.IsUnknown() {
		diags := data.User.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !data.ServiceAccount.EditablePermissions.IsNull() || !data.ServiceAccount.EditablePermissions.IsUnknown() {
		diags := data.ServiceAccount.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Is the user already a member?
	if data.User.Id.ValueString() != "" {
		var user map[string]string
		diags := data.User.User.As(ctx, &user, basetypes.ObjectAsOptions{})
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
		if data.User.Id.ValueString() == "" && err != nil {
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
			data.User.Id = types.StringValue("0")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}

		if org_membership_response.User.ID != "" {
			data.MemberId = types.StringValue(org_membership_response.User.ID)
		}
	}

	var affiliation string
	if !data.User.Affiliation.IsNull() && !data.User.Affiliation.IsUnknown() {
		affiliation = data.User.Affiliation.ValueString()
	} else if !data.ServiceAccount.Affiliation.IsNull() && !data.ServiceAccount.Affiliation.IsUnknown() {
		affiliation = data.ServiceAccount.Affiliation.ValueString()
	}

	response, err := r.client.CreateOrUpdateOrganizationMembership(data.OrganizationId.ValueString(), data.MemberId.ValueString(), affiliation, elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	if response.ServiceAccount.ID != "" {
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
		data.ServiceAccount.Affiliation = types.StringValue(response.Affiliation)
		data.ServiceAccount.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType
			response.Permissions)
		data.ServiceAccount.ServiceAccount, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"id":          types.StringType,
			"email":       types.StringType,
			"name":        types.StringType,
			"description": types.StringType,	
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		}, response.ServiceAccount)
	}

	if response.User.ID != "" {
		data.MemberId = types.StringValue(response.User.ID)
		data.User.User, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"id":          types.StringType,
			"email":       types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		}, response.User)
		data.User.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)
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

	if response.ServiceAccount.ID != "" {
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
		data.ServiceAccount.Affiliation = types.StringValue(response.Affiliation)
		data.ServiceAccount.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)
		data.ServiceAccount.ServiceAccount = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"id":          types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		}, response.ServiceAccount)
		)
	}
	if response.User.ID != "" {
		data.MemberId = types.StringValue(response.User.ID)
		data.User.Affiliation = types.StringValue(response.Affiliation)
		data.User.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)
		data.User.User = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"id":          types.StringType,
			"email":       types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		}, response.User)
	}

	// Data value setting
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

	// planned_email := data.Email.ValueString()
	// resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("email"), &data.Email)...)
	// if planned_email != data.Email.ValueString() {
	// 	resp.Diagnostics.AddError("", "Updating the 'email' field of an organization membership is currently not implemented.")
	// 	return
	// }

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationMembership resource.")
	var elements []string
	if !data.User.EditablePermissions.IsNull() || !data.User.EditablePermissions.IsUnknown() {
		diags := data.User.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !data.ServiceAccount.EditablePermissions.IsNull() || !data.ServiceAccount.EditablePermissions.IsUnknown() {
		diags := data.ServiceAccount.EditablePermissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}



	if data.User.Id.ValueString() != "" && data.MemberId.ValueString() == "0" {
		var user map[string]string
		diags := data.User.User.As(ctx, &user, basetypes.ObjectAsOptions{})
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
	
	response, err := r.client.UpdateOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString(), data.Affiliation.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	if response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(response.ServiceAccount.ID)
	}
	if response.User.ID != "" {
		data.Id = types.StringValue(response.User.ID)
	}
	data.OrganizationId = types.StringValue(response.Organisation.ID)
	data.Email = types.StringValue(response.User.Email)
	data.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)
	data.IsActive = types.BoolValue(true)

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
		_, err := r.client.GetOrganizationInvitationByEmail(data.OrganizationId.ValueString(), data.Email.ValueString())
		if err != nil {
			return
		}
		err = r.client.DeleteOrganizationInvitation(data.OrganizationId.ValueString(), data.Email.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
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
	if response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(response.ServiceAccount.ID)
	}
	if response.User.ID != "" {
		data.Id = types.StringValue(response.User.ID)
	}
	data.OrganizationId = types.StringValue(idParts[0])
	data.Affiliation = types.StringValue(response.Affiliation)
	data.Email = types.StringValue(response.User.Email)
	data.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)
	data.IsActive = types.BoolValue(true)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
