package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/clients/iam"
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/clients/keycloak"
)

var _ provider.Provider = (*ncsProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &ncsProvider{}
	}
}

type ncsProvider struct{}

type ncsProviderModel struct {
	OidcUrl          types.String `tfsdk:"oidc_url"`
	OidcClientId     types.String `tfsdk:"oidc_client_id"`
	OidcClientSecret types.String `tfsdk:"oidc_client_secret"`
	OidcClientScope  types.String `tfsdk:"oidc_client_scope"`
	IamUrl           types.String `tfsdk:"iam_url"`
}

func (p *ncsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"oidc_url": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"oidc_client_id": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_secret": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_scope": schema.StringAttribute{
				Optional: true,
			},
			"iam_url": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *ncsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config ncsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.OidcUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_url"),
			"Unknown NCS OIDC API Url",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API url. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_OIDC_URL environment variable.",
		)
	}

	if config.OidcClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client_id"),
			"Unknown NCS OIDC API client id",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API client id. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_OIDC_CLIENT_ID environment variable.",
		)
	}

	if config.OidcClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client_secret"),
			"Unknown NCS OIDC API client secret",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API client secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_OIDC_CLIENT_SECRET environment variable.",
		)
	}

	if config.OidcClientScope.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client_scope"),
			"Unknown NCS OIDC API client scope",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API client scope. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_OIDC_CLIENT_SCOPE environment variable.",
		)
	}

	if config.IamUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_url"),
			"Unknown NCS IAM API Url",
			"The provider cannot create the IAM API client as there is an unknown configuration value for the IAM API url. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_IAM_URL environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	oidcUrl := os.Getenv("NCS_OIDC_URL")
	oidcClientId := os.Getenv("NCS_OIDC_CLIENT_ID")
	oidcClientSecret := os.Getenv("NCS_OIDC_CLIENT_SECRET")
	oidcClientScope := os.Getenv("NCS_OIDC_CLIENT_SCOPE")
	iamUrl := os.Getenv("NCS_IAM_URL")

	if !config.OidcUrl.IsNull() {
		oidcUrl = config.OidcUrl.ValueString()
	}

	if !config.OidcClientId.IsNull() {
		oidcClientId = config.OidcClientId.ValueString()
	}

	if !config.OidcClientSecret.IsNull() {
		oidcClientSecret = config.OidcClientSecret.ValueString()
	}

	if !config.OidcClientScope.IsNull() {
		oidcClientScope = config.OidcClientScope.ValueString()
	}

	if !config.IamUrl.IsNull() {
		iamUrl = config.IamUrl.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if oidcClientId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client_id"),
			"Missing NCS OIDC API client id",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the OIDC API client id. "+
				"Set the client id value in the configuration or use the NCS_OIDC_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if oidcClientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client_secret"),
			"Missing NCS OIDC API client secret",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the OIDC API client secret. "+
				"Set the client secret value in the configuration or use the NCS_OIDC_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if oidcClientScope == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client_scope"),
			"Missing NCS OIDC API client scope",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the OIDC API client scope. "+
				"Set the client scope value in the configuration or use the NCS_OIDC_CLIENT_SCOPE environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if oidcUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_url"),
			"Missing NCS IAM OIDC url",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the IAM OIDC url. "+
				"Set the url value in the configuration or use the HASHICUPS_OIDCURL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if iamUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_url"),
			"Missing NCS IAM API Url",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the IAM API url. "+
				"Set the url value in the configuration or use the NCS_IAM_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new NCS Keystone client using the configuration values
	keycloakClient := keycloak.NewClient(oidcUrl, 10).
		WithClientConfig(oidcClientId, oidcClientSecret, oidcClientScope)
	token, err := keycloakClient.Login()
	if err != nil {
		resp.Diagnostics.AddError("Login error", err.Error())
		return
	}

	// Create a new NCS IAM client using the configuration values
	client := iam.NewClient(iamUrl, 10).WithBearerToken(token)

	// Make the NCS IAM client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ncsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ncs"
}

func (p *ncsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ncsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource, NewProjectResource, NewOrganizationMembershipResource, NewProjectMembershipResource,
	}
}
