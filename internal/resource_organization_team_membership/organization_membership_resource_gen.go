// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_organization_team_membership

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func OrganizationTeamMembershipResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The UUID of the user",
				MarkdownDescription: "The UUID of the user",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required: true,
			},
			"team_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

type OrganizationTeamMembershipModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	TeamId         types.String `tfsdk:"team_id"`
}
