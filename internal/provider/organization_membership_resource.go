package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/clients/iam"
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/resource_organization_membership"
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
			fmt.Sprintf("Can not create OrganizationMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.OrganizationId.ValueString()))
		return
	}

	elements := make([]string, 0, len(data.EditablePermissions.Elements()))
	diags := data.EditablePermissions.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Is the e-mail already a member?
	org_membership_response, err := r.client.GetOrganizationMembershipByEmail(data.OrganizationId.ValueString(), data.Email.ValueString())
	if err != nil {
		// Is the e-mail at least invited with the specified permissions?
		_, err := r.client.GetOrganizationInvitationByEmailAndPermissions(data.OrganizationId.ValueString(), data.Email.ValueString())
		if err != nil {
			// Invite the e-mail
			_, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), data.Email.ValueString(), elements)
			if err != nil {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
		}
		// The email is invited with the specified permissions, but has to be activated manually
		resp.Diagnostics.AddError("InvitationNotAcceptedError",
			fmt.Sprintf("Can not create OrganizationMembership in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
				data.OrganizationId.ValueString(), data.Email.ValueString()))
		return
	}

	response, err := r.client.CreateOrganizationMembership(data.OrganizationId.ValueString(), org_membership_response.ID, elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.OrganizationId = types.StringValue(response.OrganizationId)
	data.Email = types.StringValue(response.Email)
	data.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)
	//data.NonEditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

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
	response, err := r.client.GetOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.OrganizationId = types.StringValue(response.OrganizationId)
	data.Email = types.StringValue(response.Email)
	data.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_membership.OrganizationMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Creating OrganizationMembership resource.")
	elements := make([]string, 0, len(data.EditablePermissions.Elements()))
	diags := data.EditablePermissions.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.OrganizationId = types.StringValue(response.OrganizationId)
	data.Email = types.StringValue(response.Email)
	data.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

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
	tflog.Info(ctx, "Reading OrganizationMembership resource.")
	err := r.client.DeleteOrganizationMembership(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}
