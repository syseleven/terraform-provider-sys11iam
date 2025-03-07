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
	OidcUrl            types.String `tfsdk:"oidc_url"`
	IamUrl             types.String `tfsdk:"iam_url"`
	OidcClientUsername types.String `tfsdk:"oidc_client_username"`
	OidcClientPassword types.String `tfsdk:"oidc_client_password"`
	OidcClientSecret   types.String `tfsdk:"oidc_client_secret"`
	OidcClientId       types.String `tfsdk:"oidc_client_id"`
	OidcClientScope    types.String `tfsdk:"oidc_client_scope"`
}

func (p *ncsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"oidc_url": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"iam_url": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_username": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_password": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_secret": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_id": schema.StringAttribute{
				Optional: true,
			},
			"oidc_client_scope": schema.StringAttribute{
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

	if config.IamUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_url"),
			"Unknown NCS IAM API Url",
			"The provider cannot create the IAM API client as there is an unknown configuration value for the IAM API url. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NCS_IAM_URL environment variable.",
		)
	}

	if config.OidcClientSecret.IsUnknown() && (config.OidcClientUsername.IsUnknown() || config.OidcClientPassword.IsUnknown() || config.OidcClientId.IsUnknown() || config.OidcClientScope.IsUnknown()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client"),
			"Unknown NCS OIDC API client credentials. Provide either a secret (service account) or username+password+id+secret combination (regular account).",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API client. "+
				"Set the client secret value in the configuration or use the NCS_OIDC_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty."+
				"Set the client username value in the configuration or use the NCS_OIDC_CLIENT_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty."+
				"Set the client password value in the configuration or use the NCS_OIDC_CLIENT_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty."+
				"Set the client id value in the configuration or use the NCS_OIDC_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty."+
				"Set the client scope value in the configuration or use the NCS_OIDC_CLIENT_SCOPE environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	oidcUrl := os.Getenv("NCS_OIDC_URL")
	iamUrl := os.Getenv("NCS_IAM_URL")
	oidcClientUsername := os.Getenv("NCS_OIDC_CLIENT_USERNAME")
	oidcClientPassword := os.Getenv("NCS_OIDC_CLIENT_PASSWORD")
	oidcClientSecret := os.Getenv("NCS_OIDC_CLIENT_SECRET")
	oidcClientId := os.Getenv("NCS_OIDC_CLIENT_ID")
	oidcClientScope := os.Getenv("NCS_OIDC_CLIENT_SCOPE")

	if !config.OidcUrl.IsNull() {
		oidcUrl = config.OidcUrl.ValueString()
	}

	if !config.OidcClientUsername.IsNull() {
		oidcClientUsername = config.OidcClientUsername.ValueString()
	}

	if !config.OidcClientPassword.IsNull() {
		oidcClientPassword = config.OidcClientPassword.ValueString()
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

	if oidcClientSecret == "" && (oidcClientUsername == "" || oidcClientPassword == "" || oidcClientId == "" || oidcClientScope == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("oidc_client"),
			"Unknown NCS OIDC API client credentials. Provide either a secret (service account) or username+password+id+secret combination (regular account).",
			"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API client. "+
				"Set the client secret value in the configuration or use the NCS_OIDC_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty."+
				"Set the client username value in the configuration or use the NCS_OIDC_CLIENT_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty."+
				"Set the client password value in the configuration or use the NCS_OIDC_CLIENT_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty."+
				"Set the client id value in the configuration or use the NCS_OIDC_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty."+
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
	client := iam.NewClient(iamUrl, 10)
	if oidcClientId != "" {
		keycloakClient := keycloak.NewClient(oidcUrl, 10).
			WithClientConfig(oidcClientId, oidcClientSecret, oidcClientScope, oidcClientUsername, oidcClientPassword)
		token, err := keycloakClient.Login()
		if err != nil {
			resp.Diagnostics.AddError("Login error", err.Error())
			return
		}

		// Create a new NCS IAM client using the configuration values
		client.WithBearerToken(token)
	} else {
		client.WithServiceAccountToken(oidcClientSecret)
	}
	// Make the NCS IAM client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ncsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ncs"
}

func (p *ncsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
	}
}

func (p *ncsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource, NewProjectResource, NewOrganizationMembershipResource, NewProjectMembershipResource,
		NewOrganizationServiceaccountResource, NewOrganizationContactResource, NewOrganizationTeamResource,
		NewOrganizationTeamMembershipResource, NewProjectTeamMembershipResource, NewProjectS3UserResource,
		NewProjectTeamResource,
	}
}
