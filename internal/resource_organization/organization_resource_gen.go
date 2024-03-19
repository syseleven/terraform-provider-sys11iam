// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_organization

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func OrganizationResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The time the resource was created.",
				MarkdownDescription: "The time the resource was created.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
				Description:         "The UUID of the organization",
				MarkdownDescription: "The UUID of the organization",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether the organization is active or not.",
				MarkdownDescription: "Whether the organization is active or not.",
				PlanModifiers: []planmodifier.Bool{UseStateForUnknown()},
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
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

type OrganizationModel struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	Description    types.String `tfsdk:"description"`
	Id             types.String `tfsdk:"id"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	Name           types.String `tfsdk:"name"`
	Tags           types.List   `tfsdk:"tags"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}
