package resource_organization_team

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func OrganizationTeamResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A description for the organization team.",
				MarkdownDescription: "A description for the organization team.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1000),
					stringvalidator.RegexMatches(regexp.MustCompile("^[^\u0000]*$"), ""),
				},
				Default: stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The UUID of the organization team",
				MarkdownDescription: "The UUID of the organization team",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "A unique name for the organization team.",
				MarkdownDescription: "A unique name for the organization team.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 62),
					stringvalidator.RegexMatches(regexp.MustCompile("^[a-z0-9]+(?:-[a-z0-9]+)*$"), ""),
				},
			},
			"organization_id": schema.StringAttribute{
				Required: true,
			},
			"organization_permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The permissions of the service account",
				MarkdownDescription: "The permissions of the service account",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "The tags of the organization team.",
				MarkdownDescription: "The tags of the organization team.",
			},
			"projects": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				Default: listdefault.StaticValue(types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{
					"id": types.StringType,
					"project_permissions": types.ListType{
						ElemType: types.StringType,
					},
				}}, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The UUID of the project",
							MarkdownDescription: "The UUID of the project",
						},
						"project_permissions": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Description:         "The permissions of the project",
							MarkdownDescription: "The permissions of the project",
							Validators: []validator.List{
								listvalidator.UniqueValues(),
							},
						},
					},
				},
			},
		},
	}
}

type OrganizationTeamModel struct {
	Id                      types.String        `tfsdk:"id"`
	Name                    types.String        `tfsdk:"name"`
	OrganizationId          types.String        `tfsdk:"organization_id"`
	Description             types.String        `tfsdk:"description"`
	Tags                    types.List          `tfsdk:"tags"`
	OrganizationPermissions types.List          `tfsdk:"organization_permissions"`
	Projects                basetypes.ListValue `tfsdk:"projects"`
}

type ProjectValue struct {
	Id                 types.String `tfsdk:"id"`
	ProjectPermissions types.List   `tfsdk:"project_permissions"`
}
