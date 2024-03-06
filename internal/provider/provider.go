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
	OidcUrl      types.String `tfsdk:"oidc_url"`
	OidcUsername types.String `tfsdk:"oidc_username"`
	OidcPassword types.String `tfsdk:"oidc_password"`
	IamUrl       types.String `tfsdk:"iam_url"`
}

func (p *ncsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"oidc_url": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"oidc_username": schema.StringAttribute{
				Optional: true,
			},
			"oidc_password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
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

	if config.OidcUsername.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_username"),
			"Unknown NCS OIDC API Username",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_OIDC_USERNAME environment variable.",
		)
	}

	if config.OidcPassword.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_password"),
			"Unknown NCS OIDC API Password",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_OIDC_PASSWORD environment variable.",
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
	oidcUsername := os.Getenv("NCS_OIDC_USERNAME")
	oidcPassword := os.Getenv("NCS_OIDC_PASSWORD")
	iamUrl := os.Getenv("NCS_IAM_URL")

	if !config.OidcUrl.IsNull() {
		oidcUrl = config.OidcUrl.ValueString()
	}

	if !config.OidcUsername.IsNull() {
		oidcUsername = config.OidcUsername.ValueString()
	}

	if !config.OidcPassword.IsNull() {
		oidcPassword = config.OidcPassword.ValueString()
	}

	if !config.IamUrl.IsNull() {
		iamUrl = config.IamUrl.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if oidcUsername == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_username"),
			"Missing NCS OIDC API Username",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the OIDC API username. "+
				"Set the username value in the configuration or use the NCS_OIDC_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if oidcPassword == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_password"),
			"Missing NCS OIDC API Password",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the OIDC API password. "+
				"Set the password value in the configuration or use the NCS_OIDC_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if oidcUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_url"),
			"Missing NCS IAM OIDC url",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the IAM OIDC url. "+
				"Set the password value in the configuration or use the HASHICUPS_OIDCURL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if iamUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_url"),
			"Missing NCS IAM API Url",
			"The provider cannot create the NCS IAM API client as there is a missing or empty value for the IAM API url. "+
				"Set the host value in the configuration or use the NCS_IAM_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new NCS Keystone client using the configuration values
	keycloakClient := keycloak.NewClient(oidcUrl, 10).WithLogin(oidcUsername, oidcPassword)
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
		NewOrganizationResource, NewProjectResource, NewOrganizationMembershipResource,
	}
}
