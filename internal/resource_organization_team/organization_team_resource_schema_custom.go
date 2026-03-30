package resource_organization_team

// This file adds fields to the generated team schema that are managed via
// separate API endpoints (org-level team permissions and project-level team
// permissions). The generator only covers the team CRUD endpoints.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
)

// OrganizationTeamResourceSchemaFull wraps the generated schema and adds
// organization_permissions and projects attributes.
func OrganizationTeamResourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationTeamResourceSchema(ctx)

	// org_id must be required (generator marks path params as optional+computed)
	s.Attributes["org_id"] = schema.StringAttribute{
		Optional: true,
		Computed: true,
	}

	// team_id is redundant with id — remove it
	delete(s.Attributes, "team_id")

	// Deprecated organization_id alias for backwards compatibility
	s.Attributes["organization_id"] = compat.DeprecatedOrganizationIdAttribute()

	s.Version = 1

	// Organization-level team permissions (managed via /orgs/{org_id}/teams/{team_id}/permissions)
	s.Attributes["organization_permissions"] = schema.ListAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Description:         "The organization-level permissions for this team.",
		MarkdownDescription: "The organization-level permissions for this team.",
		Validators: []validator.List{
			listvalidator.UniqueValues(),
		},
	}

	// Project-level team permissions (managed via /orgs/{org_id}/projects/{project_id}/teams/{team_id}/permissions)
	s.Attributes["projects"] = schema.ListNestedAttribute{
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
	}

	return s
}

// OrganizationTeamModelFull extends the generated OrganizationTeamModel with
// fields managed via separate API endpoints.
type OrganizationTeamModelFull struct {
	Description             types.String        `tfsdk:"description"`
	Id                      types.String        `tfsdk:"id"`
	Name                    types.String        `tfsdk:"name"`
	OrgId                   types.String        `tfsdk:"org_id"`
	OrganizationId          types.String        `tfsdk:"organization_id"`
	Tags                    types.List          `tfsdk:"tags"`
	OrganizationPermissions types.List          `tfsdk:"organization_permissions"`
	Projects                basetypes.ListValue `tfsdk:"projects"`
}

// ProjectValue represents a project with its team-level permissions.
type ProjectValue struct {
	Id                 types.String `tfsdk:"id"`
	ProjectPermissions types.List   `tfsdk:"project_permissions"`
}
