package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/datasource_organization"
)

var (
	_ datasource.DataSourceWithValidateConfig = &organizationDataSource{}
)

func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

type organizationDataSource struct {
	client *iam.Client
}

func (r *organizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_organization.OrganizationDataSourceSchemaFull(ctx)
}

func (r *organizationDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data datasource_organization.OrganizationModelFull
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idConfigured := !data.Id.IsNull()
	nameConfigured := !data.Name.IsNull()
	if !idConfigured && !nameConfigured {
		resp.Diagnostics.AddError(
			"Invalid organization lookup",
			"At least one of id or name must be configured.",
		)
		return
	}
	if idConfigured && nameConfigured {
		resp.Diagnostics.AddWarning(
			"Both organization lookup attributes are configured",
			"Both id and name are currently accepted for compatibility. Configure only one because providing both may be deprecated in a future version.",
		)
	}
}

func (r *organizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_organization.OrganizationModelFull

	// Read Terraform configuration into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		response iam.IAMOrganization
		err      error
	)

	switch {
	case !data.Id.IsNull() && !data.Id.IsUnknown():
		tflog.Info(ctx, "Reading organization by ID.")
		response, err = r.client.GetOrganization(data.Id.ValueString())
	case !data.Name.IsNull() && !data.Name.IsUnknown():
		tflog.Info(ctx, "Reading organization by name.")
		response, err = r.client.GetOrganizationByName(data.Name.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Invalid organization lookup",
			"At least one of id or name must be configured with a known value.",
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if response.ID == "" {
		resp.Diagnostics.AddError(
			"Organization not found",
			"The organization lookup did not return an organization.",
		)
		return
	}

	// Data value setting.
	data.Id = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)
	data.IsActive = types.BoolValue(response.IsActive)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)
	data.CompanyInfoStreet = types.StringValue(response.CompanyInfo.Street)
	data.CompanyInfoStreetNumber = types.StringValue(response.CompanyInfo.StreetNumber)
	data.CompanyInfoZipCode = types.StringValue(response.CompanyInfo.ZipCode)
	data.CompanyInfoCity = types.StringValue(response.CompanyInfo.City)
	data.CompanyInfoCountry = types.StringValue(response.CompanyInfo.Country)
	data.CompanyInfoVatID = types.StringValue(response.CompanyInfo.VatID)
	data.CompanyInfoPreferredBillingMethod = types.StringValue(response.CompanyInfo.PreferredBillingMethod)
	data.CompanyInfoPhone = types.StringValue(response.CompanyInfo.Phone)
	data.CompanyInfoAcceptedTos = types.BoolValue(response.CompanyInfo.AcceptedTos)
	data.CompanyInfoCompanyName = types.StringValue(response.CompanyInfo.CompanyName)

	// Emit manual steps as warnings.
	if !data.IsActive.ValueBool() {
		resp.Diagnostics.AddWarning("OrganizationNotActiveWarning",
			fmt.Sprintf("Organization with id %s is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.Id.ValueString()))
	}

	// Save updated data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

/**
func (r *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datasource_organization.OrganizationModel

	// Read Terraform prior state data into the model
	//resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Read API call logic
	tflog.Info(ctx, "Reading organization datasource.")
	response, err := r.client.GetOrganizations()
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Map response body to model
	for _, org := range response {
		orgState := datasource_organization.OrganizationModel{
			Id:          types.StringValue(org.ID),
			Name:        types.StringValue(org.Name),
			Description: types.StringValue(org.Description),
			CreatedAt:   types.StringValue(org.CreatedAt),
			UpdatedAt:   types.StringValue(org.UpdatedAt),
			IsActive:    types.BoolValue(org.IsActive),
			Tags:        types.ListValueFrom(ctx, types.StringType, org.Tags),
		}
		state = append(state, orgState)

		// Emit manual steps as warnings
		if !org.IsActive {
			resp.Diagnostics.AddWarning("OrganizationNotActiveWarning",
				fmt.Sprintf("Organization with id %s is not active. Organization activation is a manual step, please contact an IAM administrator.",
					org.ID))
		}
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
**/
