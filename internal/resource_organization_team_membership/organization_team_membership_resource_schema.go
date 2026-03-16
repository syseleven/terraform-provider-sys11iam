// Permissions of a user or service account within an organization team.
package resource_organization_team_membership

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func OrganizationTeamMembershipResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Required: true,
			},
			"org_id": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The UUID of the organization team membership, user or service account.",
			},
			"team_name": schema.StringAttribute{
				Computed: true,
			},
			"membership_type": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Description:         "The type of the membership.",
				MarkdownDescription: "The type of the membership.",
			},
			"membership": schema.SingleNestedAttribute{
				// Validators: []validator.Object{
				// 	objectvalidator.ExactlyOneOf(path.MatchRelative().AtName("user_team_membership"), path.MatchRelative().AtName("service_account_team_membership")),
				// },
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"user_team_membership": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"team_permissions": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
								Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
								Description:         "The team permissions the user has in the team",
								MarkdownDescription: "The team permissions the user has in the team",
							},
							"user": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"email": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Description:         "The email address of the user.",
										MarkdownDescription: "The email address of the user.",
									},
								},
								Optional: true,
								Computed: true,
							},
						},
						Optional: true,
						Computed: true,
					},
					"service_account_team_membership": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"team_permissions": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
								Description:         "The team permissions the user has in the team",
								MarkdownDescription: "The team permissions the user has in the team",
								Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
							},
						},
						Optional: true,
						Computed: true,
					},
				},
			},
		},
	}
}

type OrganizationTeamMembershipModel struct {
	Id             types.String          `tfsdk:"id"`
	Membership     basetypes.ObjectValue `tfsdk:"membership"`
	MembershipType types.String          `tfsdk:"membership_type"`
	OrganizationId types.String          `tfsdk:"org_id"`
	TeamID         types.String          `tfsdk:"team_id"`
	TeamName       types.String          `tfsdk:"team_name"`
}

type MembershipValue struct {
	ServiceAccountMembership basetypes.ObjectValue `tfsdk:"service_account_team_membership"`
	UserTeamMembership       basetypes.ObjectValue `tfsdk:"user_team_membership"`
}

type ServiceAccountMembershipValue struct {
	TeamPermissions types.List `tfsdk:"team_permissions"`
}

type OrganizationValue struct {
	Id       types.String `tfsdk:"id"`
	IsActive types.Bool   `tfsdk:"is_active"`
	Name     types.String `tfsdk:"name"`
}

type UserTeamMembershipValue struct {
	User            basetypes.ObjectValue `tfsdk:"user"`
	TeamPermissions types.List            `tfsdk:"team_permissions"`
}

type UserValue struct {
	Email types.String `tfsdk:"email"`
}
