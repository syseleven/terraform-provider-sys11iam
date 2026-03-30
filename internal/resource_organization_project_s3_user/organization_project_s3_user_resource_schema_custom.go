package resource_organization_project_s3_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
)

// OrganizationProjectS3UserResourceSchemaFull wraps the generated schema and adds
// the deprecated organization_id alias for backwards compatibility.
func OrganizationProjectS3UserResourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationProjectS3UserResourceSchema(ctx)
	s.Attributes["organization_id"] = compat.DeprecatedOrganizationIdAttribute()
	s.Version = 1
	return s
}

// OrganizationProjectS3UserModelFull extends the generated model with the
// deprecated organization_id field.
type OrganizationProjectS3UserModelFull struct {
	Description    types.String `tfsdk:"description"`
	Id             types.String `tfsdk:"id"`
	Keys           types.List   `tfsdk:"keys"`
	Name           types.String `tfsdk:"name"`
	OrgId          types.String `tfsdk:"org_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	S3UserId       types.String `tfsdk:"s3_user_id"`
}
