package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization"
)

var _ resource.Resource = (*organizationResource)(nil)
var _ resource.ResourceWithConfigure = (*organizationResource)(nil)
var _ resource.ResourceWithUpgradeState = (*organizationResource)(nil)

func NewOrganizationResource() resource.Resource {
	return &organizationResource{}
}

type organizationResource struct {
	client *iam.Client
}

func (r *organizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization.OrganizationResourceSchema(ctx)
}

func (r *organizationResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: upgradeOrganizationStateV0,
		},
	}
}

// legacyOrganizationState is the state shape written by provider versions
// <= v1.5.x (schema version 0), where company info was stored as flat
// company_info_* attributes.
type legacyOrganizationState struct {
	Id                                *string  `json:"id"`
	Name                              *string  `json:"name"`
	Description                       *string  `json:"description"`
	IsActive                          *bool    `json:"is_active"`
	Tags                              []string `json:"tags"`
	CreatedAt                         *string  `json:"created_at"`
	UpdatedAt                         *string  `json:"updated_at"`
	CompanyInfoStreet                 *string  `json:"company_info_street"`
	CompanyInfoStreetNumber           *string  `json:"company_info_street_number"`
	CompanyInfoZipCode                *string  `json:"company_info_zip_code"`
	CompanyInfoCity                   *string  `json:"company_info_city"`
	CompanyInfoCountry                *string  `json:"company_info_country"`
	CompanyInfoVatId                  *string  `json:"company_info_vat_id"`
	CompanyInfoPreferredBillingMethod *string  `json:"company_info_preferred_billing_method"`
	CompanyInfoPhone                  *string  `json:"company_info_phone"`
	CompanyInfoAcceptedTos            *bool    `json:"company_info_accepted_tos"`
	CompanyInfoCompanyName            *string  `json:"company_info_company_name"`
}

// upgradeOrganizationStateV0 migrates the flat company_info_* attributes into
// the current nested company_info object.
func upgradeOrganizationStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	if req.RawState == nil || len(req.RawState.JSON) == 0 {
		resp.Diagnostics.AddError(
			"Error reading prior state",
			"No prior state data available for migration.",
		)
		return
	}

	var legacy legacyOrganizationState
	if err := json.Unmarshal(req.RawState.JSON, &legacy); err != nil {
		resp.Diagnostics.AddError(
			"Error reading prior state",
			"Could not unmarshal prior state data: "+err.Error(),
		)
		return
	}

	attrTypes := resource_organization.CompanyInfoValue{}.AttributeTypes(ctx)
	companyInfoAttrs := map[string]attr.Value{
		"accepted_tos":             basetypes.NewBoolNull(),
		"city":                     basetypes.NewStringNull(),
		"company_name":             basetypes.NewStringNull(),
		"country":                  basetypes.NewStringNull(),
		"phone_number":             basetypes.NewStringNull(),
		"preferred_billing_method": basetypes.NewStringNull(),
		"street":                   basetypes.NewStringNull(),
		"street_number":            basetypes.NewStringNull(),
		"vat_id":                   basetypes.NewStringNull(),
		"zip_code":                 basetypes.NewStringNull(),
	}

	hasCompanyInfo := false
	setString := func(name string, value *string) {
		if value != nil && *value != "" {
			companyInfoAttrs[name] = basetypes.NewStringValue(*value)
			hasCompanyInfo = true
		}
	}
	setBool := func(name string, value *bool) {
		if value != nil {
			companyInfoAttrs[name] = basetypes.NewBoolValue(*value)
			hasCompanyInfo = true
		}
	}

	setString("street", legacy.CompanyInfoStreet)
	setString("street_number", legacy.CompanyInfoStreetNumber)
	setString("zip_code", legacy.CompanyInfoZipCode)
	setString("city", legacy.CompanyInfoCity)
	setString("country", legacy.CompanyInfoCountry)
	setString("vat_id", legacy.CompanyInfoVatId)
	setString("preferred_billing_method", legacy.CompanyInfoPreferredBillingMethod)
	setString("phone_number", legacy.CompanyInfoPhone)
	setBool("accepted_tos", legacy.CompanyInfoAcceptedTos)
	setString("company_name", legacy.CompanyInfoCompanyName)

	var companyInfo resource_organization.CompanyInfoValue
	if hasCompanyInfo {
		var diags diag.Diagnostics
		companyInfo, diags = resource_organization.NewCompanyInfoValue(attrTypes, companyInfoAttrs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		companyInfo = resource_organization.NewCompanyInfoValueNull()
	}

	tags := []attr.Value{}
	for _, tag := range legacy.Tags {
		tags = append(tags, types.StringValue(tag))
	}

	data := resource_organization.OrganizationModel{
		CompanyInfo: companyInfo,
		Id:          stringOrNull(legacy.Id),
		Name:        stringOrNull(legacy.Name),
		Description: stringOrNull(legacy.Description),
		IsActive:    basetypes.NewBoolNull(),
		Tags:        types.ListValueMust(types.StringType, tags),
		CreatedAt:   stringOrNull(legacy.CreatedAt),
		UpdatedAt:   stringOrNull(legacy.UpdatedAt),
		OrgId:       types.StringNull(),
	}
	if legacy.IsActive != nil {
		data.IsActive = basetypes.NewBoolValue(*legacy.IsActive)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// stringOrNull converts a legacy *string state field into a types.String,
// mapping absent/empty values to null.
func stringOrNull(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization.OrganizationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Info(ctx, "Creating organization resource.")
	elements := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.ValueString() != "" {
		response, err := r.client.GetOrganization(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
		// Data value setting
		data.Id = types.StringValue(response.ID)
		data.Name = types.StringValue(response.Name)
		data.Description = types.StringValue(response.Description)
		data.CreatedAt = types.StringValue(response.CreatedAt)
		data.UpdatedAt = types.StringValue(response.UpdatedAt)
		data.IsActive = types.BoolValue(response.IsActive)
		data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)
	} else {
		iAMOrganization := iam.IAMOrganization{
			Name:        data.Name.ValueString(),
			Description: data.Description.ValueString(),
			Tags:        elements,
			CompanyInfo: iam.IAMOrganizationCompanyInfo{
				Street:                 data.CompanyInfo.Street.ValueString(),
				StreetNumber:           data.CompanyInfo.StreetNumber.ValueString(),
				ZipCode:                data.CompanyInfo.ZipCode.ValueString(),
				City:                   data.CompanyInfo.City.ValueString(),
				VatID:                  data.CompanyInfo.VatId.ValueString(),
				PreferredBillingMethod: data.CompanyInfo.PreferredBillingMethod.ValueString(),
				AcceptedTos:            data.CompanyInfo.AcceptedTos.ValueBool(),
				CompanyName:            data.CompanyInfo.CompanyName.ValueString(),
			},
		}
		response, err := r.client.CreateOrganization(iAMOrganization)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
		// Data value setting
		data.Id = types.StringValue(response.ID)
		data.Name = types.StringValue(response.Name)
		data.Description = types.StringValue(response.Description)
		data.CreatedAt = types.StringValue(response.CreatedAt)
		data.UpdatedAt = types.StringValue(response.UpdatedAt)
		data.IsActive = types.BoolValue(response.IsActive)
		data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)
	}

	// Emit manual steps as warnings
	if !data.IsActive.ValueBool() {
		resp.Diagnostics.AddWarning("OrganizationNotActiveWarning",
			fmt.Sprintf("Organization with id %s is not active. Organization activation is a manual step, please contact the SysEleven GmbH Sales Team <sales@syseleven.de>.\n This can also be done via https://dashboard.syseleven.de/dashboard.",
				data.Id.ValueString()))
	} else {
		resp.Diagnostics.AddWarning("OrganizationAlreadyActiveWarning",
			fmt.Sprintf("Organization with id %s did already exist and is active. Please rerun terraform to create the resources depending on this organization.",
				data.Id.ValueString()))
	}
	data.IsActive = types.BoolValue(false)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization.OrganizationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading organization resource.")
	response, err := r.client.GetOrganizationByName(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if response.Name != data.Name.ValueString() {
		response, err = r.client.GetOrganization(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	// Data value setting
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.IsActive = types.BoolValue(response.IsActive)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Emit manual steps as warnings
	if !data.IsActive.ValueBool() {
		resp.Diagnostics.AddWarning("OrganizationNotActiveWarning",
			fmt.Sprintf("Organization with id %s is not active. Organization activation is a manual step, please contact the SysEleven GmbH Sales Team <sales@syseleven.de>.\n This can also be done via https://dashboard.syseleven.de/dashboard.",
				data.Id.ValueString()))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization.OrganizationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("is_active"), &data.IsActive)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating organization resource.")
	elements := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	iAMOrganization := iam.IAMOrganization{
		Description: data.Description.ValueString(),
		Tags:        elements,
	}

	response, err := r.client.UpdateOrganization(data.Id.ValueString(), iAMOrganization)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.IsActive = types.BoolValue(response.IsActive)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization.OrganizationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Info(ctx, "Reading organization resource.")
	err := r.client.DeleteOrganization(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 1 || idParts[0] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic

	response, err := r.client.GetOrganization(idParts[0])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization.OrganizationModel
	// Data value setting
	data.Id = types.StringValue(idParts[0])
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.IsActive = types.BoolValue(response.IsActive)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Emit manual steps as warnings
	if !data.IsActive.ValueBool() {
		resp.Diagnostics.AddWarning("OrganizationNotActiveWarning",
			fmt.Sprintf("Organization with id %s is not active. Organization activation is a manual step, please contact the SysEleven GmbH Sales Team <sales@syseleven.de>.\n This can also be done via https://dashboard.syseleven.de/dashboard.",
				data.Id.ValueString()))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
