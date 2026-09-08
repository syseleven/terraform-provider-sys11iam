package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/datasource_organization_project"
)

var _ datasource.DataSource = &projectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client *iam.Client
}

func (r *projectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_project"
}

func (r *projectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_organization_project.OrganizationProjectDataSourceSchemaFull(ctx)
}

func (r *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*iam.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *iam.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_organization_project.OrganizationProjectModelFull
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading project by ID.")
	response, err := r.client.GetProject(data.OrgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if response.ID == "" {
		resp.Diagnostics.AddError(
			"Project not found",
			"The project lookup did not return a project.",
		)
		return
	}

	data.Id = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.IsManagedByS11 = types.BoolValue(response.IsManagedByS11)
	data.Status = types.StringValue(response.Status)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)
	data.CreatedAt = types.StringValue(response.CreatedAt)
	data.UpdatedAt = types.StringValue(response.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
