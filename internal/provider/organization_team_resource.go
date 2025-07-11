package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_team"
)

var _ resource.Resource = (*OrganizationTeamResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationTeamResource)(nil)

func NewOrganizationTeamResource() resource.Resource {
	return &OrganizationTeamResource{}
}

type OrganizationTeamResource struct {
	client *iam.Client
}

func (r *OrganizationTeamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_team"
}

func (r *OrganizationTeamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_team.OrganizationTeamResourceSchema(ctx)
}

func (r *OrganizationTeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationTeam resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))

	// Is the organization active?
	org_response, err := r.client.GetOrganization(data.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create OrganizationTeam in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.OrganizationId.ValueString()))
		return
	}

	elements := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.CreateOrganizationTeam(data.OrganizationId.ValueString(), data.Name.ValueString(), data.Description.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	data.Id = types.StringValue(response.ID)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationTeam resource.")
	response, err := r.client.GetOrganizationTeam(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationTeam resource.")
	elements := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateOrganizationTeam(data.OrganizationId.ValueString(), data.Id.ValueString(), data.Name.ValueString(), data.Description.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Info(ctx, "Deleting OrganizationTeam resource.")
	err := r.client.DeleteOrganizationTeam(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *OrganizationTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || (len(idParts) == 2 && (idParts[0] == "" || idParts[1] == "")) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,team_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationTeam resource.")
	response, err := r.client.GetOrganizationTeam(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	response_permissions, err := r.client.GetOrganizationTeamPermissions(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_team.OrganizationTeamModel

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.OrganizationId = types.StringValue(idParts[0])
	data.Description = types.StringValue(response.Description)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)
	data.EditablePermissions, _ = types.ListValueFrom(ctx, types.StringType, response_permissions.TeamPermissions)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
