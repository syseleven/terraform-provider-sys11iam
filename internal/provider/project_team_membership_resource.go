package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_project_team_membership"
)

var _ resource.Resource = (*ProjectTeamMembershipResource)(nil)
var _ resource.ResourceWithConfigure = (*ProjectTeamMembershipResource)(nil)

func NewProjectTeamMembershipResource() resource.Resource {
	return &ProjectTeamMembershipResource{}
}

type ProjectTeamMembershipResource struct {
	client *iam.Client
}

func (r *ProjectTeamMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_team_membership"
}

func (r *ProjectTeamMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_project_team_membership.ProjectTeamMembershipResourceSchema(ctx)
}

func (r *ProjectTeamMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectTeamMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_project_team_membership.ProjectTeamMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Info(ctx, "Creating ProjectTeamMembership resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))

	// Is the organization active?
	org_response, err := r.client.GetOrganization(data.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create ProjectTeamMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.OrganizationId.ValueString()))
		return
	}

	elements := make([]string, 0, len(data.Permissions.Elements()))
	diags := data.Permissions.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.CreateProjectTeamMembership(data.OrganizationId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString(), data.Id.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	data.Permissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectTeamMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_project_team_membership.ProjectTeamMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectTeamMembership resource.")
	response, err := r.client.GetProjectTeamMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Permissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectTeamMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_project_team_membership.ProjectTeamMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating ProjectTeamMembership resource.")
	elements := make([]string, 0, len(data.Permissions.Elements()))
	diags := data.Permissions.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateProjectTeamMembership(data.OrganizationId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString(), data.Id.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Permissions, _ = types.ListValueFrom(ctx, types.StringType, response.Permissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectTeamMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_project_team_membership.ProjectTeamMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Info(ctx, "Deleting ProjectTeamMembership resource.")
	err := r.client.DeleteProjectTeamMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}
