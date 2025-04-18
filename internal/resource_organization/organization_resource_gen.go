// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_organization

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func OrganizationResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The time the resource was created.",
				MarkdownDescription: "The time the resource was created.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A description for the organization.",
				MarkdownDescription: "A description for the organization.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]*$"), ""),
				},
				Default: stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The UUID of the organization",
				MarkdownDescription: "The UUID of the organization",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_active": schema.BoolAttribute{
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether the organization is active or not.",
				MarkdownDescription: "Whether the organization is active or not.",
				PlanModifiers:       []planmodifier.Bool{UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "A unique name for the organization.",
				MarkdownDescription: "A unique name for the organization.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The tags of the organization.",
				MarkdownDescription: "The tags of the organization.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The time the resource was last updated.",
				MarkdownDescription: "The time the resource was last updated.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"company_info_street": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations street.",
				MarkdownDescription: "The organizations street.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_street_number": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations street number.",
				MarkdownDescription: "The organizations street number.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_zip_code": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations zip code.",
				MarkdownDescription: "The organizations zip code.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_city": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations city.",
				MarkdownDescription: "The organizations city.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_country": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations country.",
				MarkdownDescription: "The organizations country.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_vat_id": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations vat ID.",
				MarkdownDescription: "The organizations vat ID,",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_preferred_billing_method": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations preferred billing method.",
				MarkdownDescription: "The organizations preferred billing method.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_phone": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations phone.",
				MarkdownDescription: "The organizations phone.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"company_info_accepted_tos": schema.BoolAttribute{
				Required:            true,
				Description:         "Whether the organization has accepted the terms of service or not.",
				MarkdownDescription: "Whether the organization has accepted the terms of service or not.",
				PlanModifiers:       []planmodifier.Bool{UseStateForUnknown()},
			},
			"company_info_company_name": schema.StringAttribute{
				Required:            true,
				Description:         "The organizations company name.",
				MarkdownDescription: "The organizations company name.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
		},
	}
}

type OrganizationModel struct {
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
