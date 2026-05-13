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
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_s3_user"
)

var _ resource.Resource = (*ProjectS3UserResource)(nil)
var _ resource.ResourceWithConfigure = (*ProjectS3UserResource)(nil)
var _ resource.ResourceWithMoveState = (*ProjectS3UserResource)(nil)
var _ resource.ResourceWithUpgradeState = (*ProjectS3UserResource)(nil)

func NewProjectS3UserResource() resource.Resource {
	return &ProjectS3UserResource{}
}

type ProjectS3UserResource struct {
	client *iam.Client
}

func (r *ProjectS3UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_project_s3_user"
}

func (r *ProjectS3UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_project_s3_user.OrganizationProjectS3UserResourceSchemaFull(ctx)
}

func (r *ProjectS3UserResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: compat.OrgIdStateUpgrader(),
	}
}

func (r *ProjectS3UserResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		compat.RawStateMover("sys11iam_project_s3user", func(ctx context.Context, rawState compat.RawState, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			orgId, organizationId, err := compat.RawOrgIDs(rawState)
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read organization identifier: "+err.Error())
				return
			}

			id, err := compat.RawString(rawState, "id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read S3 user id: "+err.Error())
				return
			}

			description, err := compat.RawString(rawState, "description")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read S3 user description: "+err.Error())
				return
			}

			name, err := compat.RawString(rawState, "name")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read S3 user name: "+err.Error())
				return
			}

			projectId, err := compat.RawString(rawState, "project_id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project id: "+err.Error())
				return
			}

			data := resource_organization_project_s3_user.OrganizationProjectS3UserModelFull{
				Description: description,
				Id:          id,
				Keys: types.ListNull(resource_organization_project_s3_user.KeysType{
					ObjectType: types.ObjectType{
						AttrTypes: resource_organization_project_s3_user.KeysValue{}.AttributeTypes(ctx),
					},
				}),
				Name:           name,
				OrgId:          orgId,
				OrganizationId: organizationId,
				ProjectId:      projectId,
				S3UserId:       id,
			}

			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &data)...)
		}),
	}
}

func (r *ProjectS3UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectS3UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_project_s3_user.OrganizationProjectS3UserModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Create API call logic
	tflog.Info(ctx, "Creating ProjectS3User resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", orgId.ValueString()))

	// Is the organization active?
	org_response, err := r.client.GetOrganization(orgId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create ProjectS3User in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				orgId.ValueString()))
		return
	}

	response, err := r.client.CreateProjectS3User(orgId.ValueString(), data.ProjectId.ValueString(), data.Name.ValueString(), data.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	data.Id = types.StringValue(response.ID)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectS3UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_project_s3_user.OrganizationProjectS3UserModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectS3User resource.")
	response, err := r.client.GetProjectS3User(orgId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectS3UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_project_s3_user.OrganizationProjectS3UserModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Update API call logic
	tflog.Info(ctx, "Updating ProjectS3User resource.")

	response, err := r.client.UpdateProjectS3User(orgId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString(), data.Name.ValueString(), data.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectS3UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_project_s3_user.OrganizationProjectS3UserModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Delete API call logic
	tflog.Info(ctx, "Deleting ProjectS3User resource.")
	err := r.client.DeleteProjectS3User(orgId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *ProjectS3UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,project_id,s3_user_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectS3User resource.")
	response, err := r.client.GetProjectS3User(idParts[0], idParts[1], idParts[2])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_project_s3_user.OrganizationProjectS3UserModelFull
	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.ProjectId = types.StringValue(idParts[1])
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(idParts[0])
	data.Description = types.StringValue(response.Description)
	data.Name = types.StringValue(response.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
