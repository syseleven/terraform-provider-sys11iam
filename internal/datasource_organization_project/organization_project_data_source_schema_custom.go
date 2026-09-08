package datasource_organization_project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OrganizationProjectDataSourceSchemaFull configures the generated project
// datasource for ID-only lookup scoped to an organization.
func OrganizationProjectDataSourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationProjectDataSourceSchema(ctx)
	delete(s.Attributes, "project_id")

	id := s.Attributes["id"].(schema.StringAttribute)
	id.Required = true
	id.Optional = false
	id.Computed = false
	id.Description = "The unique identifier of the project."
	id.MarkdownDescription = "The unique identifier of the project."
	s.Attributes["id"] = id

	name := s.Attributes["name"].(schema.StringAttribute)
	name.Required = false
	name.Optional = false
	name.Computed = true
	name.Description = "The name of the project."
	name.MarkdownDescription = "The name of the project."
	s.Attributes["name"] = name

	return s
}

type OrganizationProjectModelFull struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	Description    types.String `tfsdk:"description"`
	Id             types.String `tfsdk:"id"`
	IsManagedByS11 types.Bool   `tfsdk:"is_managed_by_s11"`
	Name           types.String `tfsdk:"name"`
	OrgId          types.String `tfsdk:"org_id"`
	Status         types.String `tfsdk:"status"`
	Tags           types.List   `tfsdk:"tags"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}
