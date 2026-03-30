package resource_organization_project_s3_user_key

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
)

// OrganizationProjectS3UserKeyResourceSchemaFull wraps the generated schema and adds
// the deprecated organization_id alias for backwards compatibility.
// Note: The generated model has a bug where the tfsdk tag is "organization_id"
// but the schema attribute is "org_id". This custom schema and model fix that.
func OrganizationProjectS3UserKeyResourceSchemaFull(ctx context.Context) schema.Schema {
	s := OrganizationProjectS3UserKeyResourceSchema(ctx)
	s.Attributes["organization_id"] = compat.DeprecatedOrganizationIdAttribute()
	s.Attributes["org_id"] = schema.StringAttribute{
		Optional: true,
		Computed: true,
	}
	s.Version = 1
	return s
}

// OrganizationProjectS3UserKeyModelFull replaces the generated model with correct
// tfsdk tags and adds the deprecated organization_id field.
type OrganizationProjectS3UserKeyModelFull struct {
	AccessKey      types.String `tfsdk:"access_key"`
	OrgId          types.String `tfsdk:"org_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	S3UserId       types.String `tfsdk:"s3_user_id"`
	SecretKey      types.String `tfsdk:"secret_key"`
}
