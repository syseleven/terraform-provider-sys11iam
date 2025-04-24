package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_project_membership"
)

var _ resource.Resource = (*ProjectMembershipResource)(nil)
var _ resource.ResourceWithConfigure = (*ProjectMembershipResource)(nil)

func NewProjectMembershipResource() resource.Resource {
	return &ProjectMembershipResource{}
}

type ProjectMembershipResource struct {
	client *iam.Client
}

func (r *ProjectMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_membership"
}

func (r *ProjectMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_project_membership.ProjectMembershipResourceSchema(ctx)
}

func (r *ProjectMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_project_membership.ProjectMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Info(ctx, "Creating ProjectMembership resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))
	// Is the organization active?
	org_response, err := r.client.GetOrganization(data.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create ProjectMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.OrganizationId.ValueString()))
		return
	}

	elements := make([]string, 0, len(data.Permissions.Elements()))
	diags := data.Permissions.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Is the e-mail already a member?
	org_membership_response, err := r.client.GetOrganizationMembershipByEmail(data.OrganizationId.ValueString(), data.Email.ValueString())
	if err != nil {
		// Is the e-mail at least invited?
		_, err := r.client.GetOrganizationInvitationByEmail(data.OrganizationId.ValueString(), data.Email.ValueString())
		if err != nil {
			// Invite the e-mail
			_, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), data.Email.ValueString(), elements)
			if err != nil {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
		}
		// The email is invited, but has to be activated manually
		resp.Diagnostics.AddError("InvitationNotAcceptedError",
			fmt.Sprintf("Can not create ProjectMembership in project with id %s in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
				data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Email.ValueString()))
		return
	}
	if org_membership_response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(org_membership_response.ServiceAccount.ID)
	}
	if org_membership_response.User.ID != "" {
		data.Id = types.StringValue(org_membership_response.User.ID)
	}

	response, err := r.client.CreateProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString(), elements)
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
	data.ProjectId = types.StringValue(response.ProjectId)
	data.Email = types.StringValue(response.User.Email)
	sort.Sort(sort.StringSlice(response.Permissions))
	data.Permissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_project_membership.ProjectMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectMembership resource.")
	response, err := r.client.GetProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString())
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
	data.ProjectId = types.StringValue(response.Project.ID)
	data.Email = types.StringValue(response.User.Email)
	sort.Sort(sort.StringSlice(response.Permissions))
	data.Permissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_project_membership.ProjectMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	planned_email := data.Email.ValueString()
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("email"), &data.Email)...)
	if planned_email != data.Email.ValueString() {
		resp.Diagnostics.AddError("", "Updating the 'email' field of a project membership is currently not implemented.")
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Creating ProjectMembership resource.")
	elements := make([]string, 0, len(data.Permissions.Elements()))
	diags := data.Permissions.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString(), elements)
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
	data.ProjectId = types.StringValue(response.ProjectId)
	data.Email = types.StringValue(response.User.Email)
	sort.Sort(sort.StringSlice(response.Permissions))
	data.Permissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_project_membership.ProjectMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Info(ctx, "Reading ProjectMembership resource.")
	err := r.client.DeleteProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *ProjectMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,project_id,member_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectMembership resource.")
	response, err := r.client.GetProjectMembership(idParts[0], idParts[1], idParts[2])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_project_membership.ProjectMembershipModel

	// Data value setting
	if response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(response.ServiceAccount.ID)
	}
	if response.User.ID != "" {
		data.Id = types.StringValue(response.User.ID)
	}
	data.ProjectId = types.StringValue(response.Project.ID)
	data.OrganizationId = types.StringValue(idParts[0])
	data.Email = types.StringValue(response.User.Email)
	sort.Sort(sort.StringSlice(response.Permissions))
	data.Permissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
