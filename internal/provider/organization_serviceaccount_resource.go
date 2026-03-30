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
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_serviceaccount"
)

var _ resource.Resource = (*OrganizationServiceaccountResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationServiceaccountResource)(nil)
var _ resource.ResourceWithUpgradeState = (*OrganizationServiceaccountResource)(nil)

func NewOrganizationServiceaccountResource() resource.Resource {
	return &OrganizationServiceaccountResource{}
}

type OrganizationServiceaccountResource struct {
	client *iam.Client
}

func (r *OrganizationServiceaccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_serviceaccount"
}

func (r *OrganizationServiceaccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_serviceaccount.OrganizationServiceaccountResourceSchemaFull(ctx)
}

func (r *OrganizationServiceaccountResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: compat.OrgIdStateUpgrader(),
	}
}

func (r *OrganizationServiceaccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationServiceaccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_serviceaccount.OrganizationServiceaccountModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationServiceaccount resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", orgId.ValueString()))
	// Is the organization active?
	org_response, err := r.client.GetOrganization(orgId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create OrganizationServiceaccount in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				orgId.ValueString()))
		return
	}

	response, err := r.client.CreateOrganizationServiceaccount(orgId.ValueString(), data.Name.ValueString(), data.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.ServiceAccountId = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationServiceaccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_serviceaccount.OrganizationServiceaccountModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationServiceaccount resource.")
	response, err := r.client.GetOrganizationServiceaccount(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.ServiceAccountId = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationServiceaccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_serviceaccount.OrganizationServiceaccountModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationServiceaccount resource.")
	response, err := r.client.UpdateOrganizationServiceaccount(orgId.ValueString(), data.Id.ValueString(), data.Name.ValueString(), data.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.ServiceAccountId = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationServiceaccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_serviceaccount.OrganizationServiceaccountModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Delete API call logic
	tflog.Info(ctx, "Deleting OrganizationServiceaccount resource.")
	err := r.client.DeleteOrganizationServiceaccount(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *OrganizationServiceaccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,service_account_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationServiceaccount resource.")
	response, err := r.client.GetOrganizationServiceaccount(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_serviceaccount.OrganizationServiceaccountModelFull

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.ServiceAccountId = types.StringValue(response.ID)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(idParts[0])
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
