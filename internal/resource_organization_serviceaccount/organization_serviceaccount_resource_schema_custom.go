package resource_organization_serviceaccount

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
)

// OrganizationServiceaccountResourceSchemaFull wraps the generated schema and adds
// the deprecated organization_id alias for backwards compatibility.
func OrganizationServiceaccountResourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationServiceaccountResourceSchema(ctx)
	s.Attributes["organization_id"] = compat.DeprecatedOrganizationIdAttribute()
	s.Version = 1
	return s
}

// OrganizationServiceaccountModelFull extends the generated model with the
// deprecated organization_id field.
type OrganizationServiceaccountModelFull struct {
	CreatedAt        types.String `tfsdk:"created_at"`
	Description      types.String `tfsdk:"description"`
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	OrgId            types.String `tfsdk:"org_id"`
	OrganizationId   types.String `tfsdk:"organization_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}
