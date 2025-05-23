// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_project_s3user

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ProjectS3UserResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A description for the S3User.",
				MarkdownDescription: "A description for the S3User.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]*$"), ""),
				},
				Default: stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The UUID of the S3User",
				MarkdownDescription: "The UUID of the S3User",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "A unique name for the S3User.",
				MarkdownDescription: "A unique name for the S3User.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[ -~]{1,62}$"), ""),
				},
			},
			"organization_id": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

type ProjectS3UserModel struct {
	Description    types.String `tfsdk:"description"`
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
}
