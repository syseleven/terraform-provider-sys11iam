package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_project_s3_user_key"
)

var _ resource.Resource = (*ProjectS3UserKeyResource)(nil)
var _ resource.ResourceWithConfigure = (*ProjectS3UserKeyResource)(nil)

func NewProjectS3UserKeyResource() resource.Resource {
	return &ProjectS3UserKeyResource{}
}

type ProjectS3UserKeyResource struct {
	client *iam.Client
}

func (r *ProjectS3UserKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_s3user_key"
}

func (r *ProjectS3UserKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_project_s3_user_key.ProjectS3UserKeyResourceSchema(ctx)
}

func (r *ProjectS3UserKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectS3UserKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_project_s3_user_key.ProjectS3UserKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the S3User key
	tflog.Info(ctx, "Creating S3User key Resource")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))

	org_response, err := r.client.GetOrganization(data.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create ProjectS3User in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.OrganizationId.ValueString()))
		return
	}

	response, err := r.client.CreateProjectS3UserKey(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.S3UserId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	data.AccessKey = types.StringValue(response.AccessKey)
	data.SecretKey = types.StringValue(response.SecretKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectS3UserKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_project_s3_user_key.ProjectS3UserKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectS3User resource.")
	response, err := r.client.GetProjectS3UserKey(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.S3UserId.ValueString(), data.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.AccessKey = types.StringValue(response.AccessKey)
	data.SecretKey = types.StringValue(response.SecretKey)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectS3UserKeyResource) ImportState(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_project_s3_user_key.ProjectS3UserKeyModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Import the S3User key
	tflog.Info(ctx, "Importing S3User key Resource")

	response, err := r.client.GetProjectS3UserKey(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.S3UserId.ValueString(), data.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	data.AccessKey = types.StringValue(response.AccessKey)
	data.SecretKey = types.StringValue(response.SecretKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectS3UserKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_project_s3_user_key.ProjectS3UserKeyModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the S3User key
	tflog.Info(ctx, "Deleting S3User key Resource")

	err := r.client.DeleteProjectS3UserKey(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.S3UserId.ValueString(), data.AccessKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *ProjectS3UserKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_project_s3_user_key.ProjectS3UserKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "ProjectS3UserKey can't be updated. Passing in unchanged state.")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
