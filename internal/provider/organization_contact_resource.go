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
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_contact"
)

var _ resource.Resource = (*OrganizationContactResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationContactResource)(nil)
var _ resource.ResourceWithUpgradeState = (*OrganizationContactResource)(nil)

func NewOrganizationContactResource() resource.Resource {
	return &OrganizationContactResource{}
}

type OrganizationContactResource struct {
	client *iam.Client
}

func (r *OrganizationContactResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_contact"
}

func (r *OrganizationContactResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_contact.OrganizationContactResourceSchemaFull(ctx)
}

func (r *OrganizationContactResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: compat.OrgIdStateUpgrader(),
	}
}

func (r *OrganizationContactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_contact.OrganizationContactModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationContact resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", orgId.ValueString()))

	// Is the organization active?
	org_response, err := r.client.GetOrganization(orgId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create OrganizationContact in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				orgId.ValueString()))
		return
	}

	elements := make([]string, 0, len(data.Roles.Elements()))
	diags := data.Roles.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.CreateOrganizationContact(orgId.ValueString(), data.FirstName.ValueString(), data.LastName.ValueString(), data.Notes.ValueString(), data.Email.ValueString(), data.Phone.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	data.Id = types.StringValue(response.ID)
	data.FirstName = types.StringValue(response.FirstName)
	data.LastName = types.StringValue(response.LastName)
	data.Phone = types.StringValue(response.Phone)
	data.Email = types.StringValue(response.Email)
	data.Notes = types.StringValue(response.Notes)
	data.Roles, _ = types.ListValueFrom(ctx, types.StringType, response.Roles)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_contact.OrganizationContactModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationContact resource.")
	response, err := r.client.GetOrganizationContact(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.FirstName = types.StringValue(response.FirstName)
	data.LastName = types.StringValue(response.LastName)
	data.Phone = types.StringValue(response.Phone)
	data.Email = types.StringValue(response.Email)
	data.Notes = types.StringValue(response.Notes)
	data.Roles, _ = types.ListValueFrom(ctx, types.StringType, response.Roles)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_contact.OrganizationContactModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationContact resource.")
	elements := make([]string, 0, len(data.Roles.Elements()))
	diags := data.Roles.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateOrganizationContact(orgId.ValueString(), data.Id.ValueString(), data.FirstName.ValueString(), data.LastName.ValueString(), data.Notes.ValueString(), data.Email.ValueString(), data.Phone.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.FirstName = types.StringValue(response.FirstName)
	data.LastName = types.StringValue(response.LastName)
	data.Phone = types.StringValue(response.Phone)
	data.Email = types.StringValue(response.Email)
	data.Notes = types.StringValue(response.Notes)
	data.Roles, _ = types.ListValueFrom(ctx, types.StringType, response.Roles)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_contact.OrganizationContactModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Delete API call logic
	tflog.Info(ctx, "Deleting OrganizationContact resource.")
	err := r.client.DeleteOrganizationContact(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *OrganizationContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,contact_id. Got: %q", req.ID),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationContact resource.")
	response, err := r.client.GetOrganizationContact(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_contact.OrganizationContactModelFull
	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.FirstName = types.StringValue(response.FirstName)
	data.LastName = types.StringValue(response.LastName)
	data.Phone = types.StringValue(response.Phone)
	data.Email = types.StringValue(response.Email)
	data.Notes = types.StringValue(response.Notes)
	data.Roles, _ = types.ListValueFrom(ctx, types.StringType, response.Roles)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(idParts[0])
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
