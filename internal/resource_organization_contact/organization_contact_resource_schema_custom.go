package resource_organization_contact

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
)

// OrganizationContactResourceSchemaFull wraps the generated schema and adds
// the deprecated organization_id alias for backwards compatibility.
func OrganizationContactResourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationContactResourceSchema(ctx)
	s.Attributes["organization_id"] = compat.DeprecatedOrganizationIdAttribute()
	s.Version = 1
	return s
}

// OrganizationContactModelFull extends the generated model with the
// deprecated organization_id field.
type OrganizationContactModelFull struct {
	ContactId      types.String `tfsdk:"contact_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	Email          types.String `tfsdk:"email"`
	FirstName      types.String `tfsdk:"first_name"`
	Id             types.String `tfsdk:"id"`
	LastName       types.String `tfsdk:"last_name"`
	Notes          types.String `tfsdk:"notes"`
	OrgId          types.String `tfsdk:"org_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Phone          types.String `tfsdk:"phone"`
	Roles          types.List   `tfsdk:"roles"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}
