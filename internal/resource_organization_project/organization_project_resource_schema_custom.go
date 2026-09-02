package resource_organization_project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
)

// OrganizationProjectResourceSchemaFull wraps the generated schema and adds
// the deprecated organization_id alias for backwards compatibility.
func OrganizationProjectResourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationProjectResourceSchema(ctx)
	s.Attributes["organization_id"] = compat.DeprecatedOrganizationIdAttribute()
	s.Version = 1
	return s
}

// OrganizationProjectModelFull extends the generated model with the
// deprecated organization_id field.
type OrganizationProjectModelFull struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	Description    types.String `tfsdk:"description"`
	Id             types.String `tfsdk:"id"`
	IsManagedByS11 types.Bool   `tfsdk:"is_managed_by_s11"`
	Name           types.String `tfsdk:"name"`
	OrgId          types.String `tfsdk:"org_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	Status         types.String `tfsdk:"status"`
	Tags           types.List   `tfsdk:"tags"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}
