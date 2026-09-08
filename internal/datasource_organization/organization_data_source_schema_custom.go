package datasource_organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OrganizationDataSourceSchemaFull turns the generated path lookup into the
// public id-or-name lookup and preserves the flattened company info fields.
func OrganizationDataSourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationDataSourceSchema(ctx)
	delete(s.Attributes, "org_id")

	id := s.Attributes["id"].(schema.StringAttribute)
	id.Optional = true
	id.Computed = true
	id.Description = "The UUID of the organization. Configure id, name, or both."
	id.MarkdownDescription = "The UUID of the organization. Configure `id`, `name`, or both."
	s.Attributes["id"] = id

	name := s.Attributes["name"].(schema.StringAttribute)
	name.Optional = true
	name.Computed = true
	name.Description = "A unique name for the organization. Configure id, name, or both."
	name.MarkdownDescription = "A unique name for the organization. Configure `id`, `name`, or both."
	s.Attributes["name"] = name

	for _, attribute := range []string{
		"company_info_street",
		"company_info_street_number",
		"company_info_zip_code",
		"company_info_city",
		"company_info_country",
		"company_info_vat_id",
		"company_info_preferred_billing_method",
		"company_info_phone",
		"company_info_company_name",
	} {
		s.Attributes[attribute] = schema.StringAttribute{Computed: true}
	}
	s.Attributes["company_info_accepted_tos"] = schema.BoolAttribute{Computed: true}

	return s
}

type OrganizationModelFull struct {
	CreatedAt                         types.String `tfsdk:"created_at"`
	Description                       types.String `tfsdk:"description"`
	Id                                types.String `tfsdk:"id"`
	IsActive                          types.Bool   `tfsdk:"is_active"`
	Name                              types.String `tfsdk:"name"`
	Tags                              types.List   `tfsdk:"tags"`
	UpdatedAt                         types.String `tfsdk:"updated_at"`
	CompanyInfoStreet                 types.String `tfsdk:"company_info_street"`
	CompanyInfoStreetNumber           types.String `tfsdk:"company_info_street_number"`
	CompanyInfoZipCode                types.String `tfsdk:"company_info_zip_code"`
	CompanyInfoCity                   types.String `tfsdk:"company_info_city"`
	CompanyInfoCountry                types.String `tfsdk:"company_info_country"`
	CompanyInfoVatID                  types.String `tfsdk:"company_info_vat_id"`
	CompanyInfoPreferredBillingMethod types.String `tfsdk:"company_info_preferred_billing_method"`
	CompanyInfoPhone                  types.String `tfsdk:"company_info_phone"`
	CompanyInfoAcceptedTos            types.Bool   `tfsdk:"company_info_accepted_tos"`
	CompanyInfoCompanyName            types.String `tfsdk:"company_info_company_name"`
}
