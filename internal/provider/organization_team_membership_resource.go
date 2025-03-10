package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/clients/iam"
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/resource_organization_team_membership"
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

	// Is the organization active?
	org_response, err := r.client.GetOrganization(data.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create OrganizationTeamMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.OrganizationId.ValueString()))
		return
	}

	_, err = r.client.CreateOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	//sort.Sort(sort.StringSlice(response.TeamPermissions))
	//data.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response.TeamPermissions)

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
	response, err := r.client.GetOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	sort.Sort(sort.StringSlice(response.TeamPermissions))

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

	response, err := r.client.UpdateOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	sort.Sort(sort.StringSlice(response.TeamPermissions))

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
	tflog.Info(ctx, "Deleting OrganizationTeamMembership resource.")
	err := r.client.DeleteOrganizationTeamMembership(data.OrganizationId.ValueString(), data.TeamId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}
