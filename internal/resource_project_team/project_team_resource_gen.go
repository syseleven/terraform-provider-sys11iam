// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_project_team

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ProjectTeamResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"editable_permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				Description:         "The editable permissions of the user",
				MarkdownDescription: "The editable permissions of the user",
			},
			"organization_id": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"team_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

type ProjectTeamModel struct {
	Permissions    types.List   `tfsdk:"editable_permissions"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	TeamId         types.String `tfsdk:"team_id"`
}
