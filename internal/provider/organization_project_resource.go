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
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project"
)

var _ resource.Resource = (*ProjectResource)(nil)
var _ resource.ResourceWithConfigure = (*ProjectResource)(nil)
var _ resource.ResourceWithMoveState = (*ProjectResource)(nil)
var _ resource.ResourceWithUpgradeState = (*ProjectResource)(nil)

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	client *iam.Client
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_project.OrganizationProjectResourceSchemaFull(ctx)
}

func (r *ProjectResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: compat.OrgIdStateUpgrader(),
	}
}

func (r *ProjectResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		compat.RawStateMover("sys11iam_project", func(ctx context.Context, rawState compat.RawState, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			orgId, organizationId, err := compat.RawOrgIDs(rawState)
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read organization identifier: "+err.Error())
				return
			}

			id, err := compat.RawString(rawState, "id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project id: "+err.Error())
				return
			}

			description, err := compat.RawString(rawState, "description")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project description: "+err.Error())
				return
			}

			name, err := compat.RawString(rawState, "name")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project name: "+err.Error())
				return
			}

			tags, err := compat.RawStringList(ctx, rawState, "tags")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project tags: "+err.Error())
				return
			}

			data := resource_organization_project.OrganizationProjectModelFull{
				CreatedAt:      types.StringNull(),
				Description:    description,
				Id:             id,
				Name:           name,
				OrgId:          orgId,
				OrganizationId: organizationId,
				ProjectId:      id,
				Status:         types.StringNull(),
				Tags:           tags,
				UpdatedAt:      types.StringNull(),
			}

			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &data)...)
		}),
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) buildData(ctx context.Context, data *resource_organization_project.OrganizationProjectModelFull, response *iam.IAMProject, organizationId string) {
	data.Id = types.StringValue(response.ID)
	data.ProjectId = types.StringValue(response.ID)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(organizationId)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.Status = types.StringValue(response.Status)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_project.OrganizationProjectModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Create API call logic
	tflog.Info(ctx, "Creating Project resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", orgId.ValueString()))
	org_response, err := r.client.GetOrganization(orgId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create project in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				orgId.ValueString()))
		return
	}

	elements := make([]string, 0, len(data.Tags.Elements()))
	if len(data.Tags.Elements()) > 0 {
		diags := data.Tags.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	response, err := r.client.CreateProject(orgId.ValueString(), data.Name.ValueString(), data.Description.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	r.buildData(ctx, &data, &response, orgId.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_project.OrganizationProjectModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Read API call logic
	tflog.Info(ctx, "Reading Project resource.")
	response, err := r.client.GetProject(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	r.buildData(ctx, &data, &response, orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_project.OrganizationProjectModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)
	if data.Name.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &data.Name)...)
	}
	if data.Description.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("description"), &data.Description)...)
	}
	if data.Tags.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("tags"), &data.Tags)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Update API call logic
	tflog.Info(ctx, "Updating Project resource.")
	elements := make([]string, 0, len(data.Tags.Elements()))
	if len(data.Tags.Elements()) > 0 {
		diags := data.Tags.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	response, err := r.client.UpdateProject(orgId.ValueString(), data.Id.ValueString(), data.Name.ValueString(), data.Description.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	r.buildData(ctx, &data, &response, orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_project.OrganizationProjectModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Delete API call logic
	tflog.Info(ctx, "Deleting Project resource.")
	err := r.client.DeleteProject(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,project_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Importing Project resource.")
	response, err := r.client.GetProject(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_project.OrganizationProjectModelFull
	// Data value setting
	r.buildData(ctx, &data, &response, idParts[0])

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
