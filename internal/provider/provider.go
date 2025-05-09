package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/keycloak"
)

var _ provider.Provider = (*sys11IamProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &sys11IamProvider{}
	}
}

type sys11IamProvider struct{}

type sys11IamProviderModel struct {
	OidcUrl              types.String `tfsdk:"oidc_url"`
	IamUrl               types.String `tfsdk:"iam_url"`
	OidcClientUsername   types.String `tfsdk:"oidc_client_username"`
	OidcClientPassword   types.String `tfsdk:"oidc_client_password"`
	OidcClientSecret     types.String `tfsdk:"oidc_client_secret"`
	OidcClientId         types.String `tfsdk:"oidc_client_id"`
	OidcClientScope      types.String `tfsdk:"oidc_client_scope"`
	ServiceAccountSecret types.String `tfsdk:"serviceaccount_secret"`
}

func (p *sys11IamProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
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
			"serviceaccount_secret": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *sys11IamProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config sys11IamProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.IamUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_url"),
			"Unknown NCS IAM API Url.",
			"The provider cannot create the IAM API client as there is an unknown configuration value for the IAM API url. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SYS11IAM_IAM_URL environment variable.",
		)
	}

	if config.ServiceAccountSecret.IsUnknown() {
		if !(config.OidcUrl.IsUnknown() || config.OidcClientSecret.IsUnknown() || config.OidcClientUsername.IsUnknown() || config.OidcClientPassword.IsUnknown() || config.OidcClientId.IsUnknown() || config.OidcClientScope.IsUnknown()) {
			if config.OidcUrl.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("oidc_url"),
					"Unknown NCS OIDC API Url",
					"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API url. "+
						"Either target apply the source of the value first, set the value statically in the configuration, or use the SYS11IAM_OIDC_URL environment variable.",
				)
			}

			if config.OidcClientSecret.IsUnknown() && config.OidcClientUsername.IsUnknown() || config.OidcClientPassword.IsUnknown() || config.OidcClientId.IsUnknown() || config.OidcClientScope.IsUnknown() {
				resp.Diagnostics.AddAttributeError(
					path.Root("oidc_client"),
					"Unknown NCS OIDC API client credentials. Provide username+password+id+secret combination (regular account).",
					"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API client. "+
						"Set the client secret value in the configuration or use the SYS11IAM_OIDC_CLIENT_SECRET environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client username value in the configuration or use the SYS11IAM_OIDC_CLIENT_USERNAME environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client password value in the configuration or use the SYS11IAM_OIDC_CLIENT_PASSWORD environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client id value in the configuration or use the SYS11IAM_OIDC_CLIENT_ID environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client scope value in the configuration or use the SYS11IAM_OIDC_CLIENT_SCOPE environment variable. "+
						"If either is already set, ensure the value is not empty.",
				)
			}
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("serviceaccount_secret"),
				"Unknown NCS service account secret. Alternatively provide regular account authentication details as described below.",
				"Set the client secret value in the configuration or use the SYS11IAM_SERVICEACCOUNT_SECRET environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	oidcUrl := os.Getenv("SYS11IAM_OIDC_URL")
	iamUrl := os.Getenv("SYS11IAM_IAM_URL")
	oidcClientUsername := os.Getenv("SYS11IAM_OIDC_CLIENT_USERNAME")
	oidcClientPassword := os.Getenv("SYS11IAM_OIDC_CLIENT_PASSWORD")
	oidcClientSecret := os.Getenv("SYS11IAM_OIDC_CLIENT_SECRET")
	oidcClientId := os.Getenv("SYS11IAM_OIDC_CLIENT_ID")
	oidcClientScope := os.Getenv("SYS11IAM_OIDC_CLIENT_SCOPE")
	serviceAccountSecret := os.Getenv("SYS11IAM_SERVICEACCOUNT_SECRET")

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

	if !config.ServiceAccountSecret.IsNull() {
		serviceAccountSecret = config.ServiceAccountSecret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if iamUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("iam_url"),
			"Unknown NCS IAM API Url.",
			"The provider cannot create the IAM API client as there is an unknown configuration value for the IAM API url. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SYS11IAM_IAM_URL environment variable.",
		)
	}

	if serviceAccountSecret == "" {
		if !(oidcUrl == "" || oidcClientSecret == "" || oidcClientUsername == "" || oidcClientPassword == "" || oidcClientId == "" || oidcClientScope == "") {
			if oidcUrl == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("oidc_url"),
					"Unknown NCS OIDC API Url",
					"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API url. "+
						"Either target apply the source of the value first, set the value statically in the configuration, or use the SYS11IAM_OIDC_URL environment variable.",
				)
			}

			if oidcClientSecret == "" && oidcClientUsername == "" || oidcClientPassword == "" || oidcClientId == "" || oidcClientScope == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("oidc_client"),
					"Unknown NCS OIDC API client credentials. Provide username+password+id+secret combination (regular account).",
					"The provider cannot create the OIDC API client as there is an unknown configuration value for the OIDC API client. "+
						"Set the client secret value in the configuration or use the SYS11IAM_OIDC_CLIENT_SECRET environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client username value in the configuration or use the SYS11IAM_OIDC_CLIENT_USERNAME environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client password value in the configuration or use the SYS11IAM_OIDC_CLIENT_PASSWORD environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client id value in the configuration or use the SYS11IAM_OIDC_CLIENT_ID environment variable. "+
						"If either is already set, ensure the value is not empty."+
						"Set the client scope value in the configuration or use the SYS11IAM_OIDC_CLIENT_SCOPE environment variable. "+
						"If either is already set, ensure the value is not empty.",
				)
			}
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("serviceaccount_secret"),
				"Unknown NCS service account secret. Alternatively provide regular account authentication details as described below.",
				"Set the client secret value in the configuration or use the SYS11IAM_SERVICEACCOUNT_SECRET environment variable. "+
					"If either is already set, ensure the value is not empty.",
			)
		}
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
		client.WithServiceAccountToken(serviceAccountSecret)
	}
	// Make the NCS IAM client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *sys11IamProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sys11iam"
}

func (p *sys11IamProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
	}
}

func (p *sys11IamProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource, NewProjectResource, NewOrganizationMembershipResource, NewProjectMembershipResource,
		NewOrganizationServiceaccountResource, NewOrganizationContactResource, NewOrganizationTeamResource,
		NewOrganizationTeamMembershipResource, NewProjectTeamMembershipResource, NewProjectS3UserResource,
		NewProjectTeamResource, NewProjectS3UserKeyResource,
	}
}
