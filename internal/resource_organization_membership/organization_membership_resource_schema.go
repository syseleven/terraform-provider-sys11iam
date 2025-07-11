package resource_organization_membership

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func OrganizationMembershipResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"membership": schema.SingleNestedAttribute{
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(path.MatchRelative().AtName("user_membership"), path.MatchRelative().AtName("service_account_membership")),
				},
				Required: true,
				Attributes: map[string]schema.Attribute{
					"service_account_membership": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"affiliation": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "The affiliation of the service account to this organization. This is not to be understood as a role.",
								MarkdownDescription: "The affiliation of the service account to this organization. This is not to be understood as a role.",
								Validators: []validator.String{
									stringvalidator.OneOf(
										"member",
										"admin",
									),
								},
							},
							"permissions": schema.ListAttribute{
								ElementType:         types.StringType,
								Required:            true,
								Description:         "The editable permissions of the service account",
								MarkdownDescription: "The editable permissions of the service account",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"id": schema.StringAttribute{
								Computed:            true,
								Optional:            true,
								Description:         "The UUID of the service account",
								MarkdownDescription: "The UUID of the service account",
							},
							"membership_type": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "The type of the membership.",
								MarkdownDescription: "The type of the membership.",
								Default:             stringdefault.StaticString("service_account"),
							},
							"name": schema.StringAttribute{
								Computed:            true,
								Optional:            true,
								Description:         "The unique name of the service account.",
								MarkdownDescription: "The unique name of the service account.",
							},
						},
						Optional: true,
					},

					"user_membership": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"affiliation": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "The affiliation of the user to this organization. This is not to be understood as a role.",
								MarkdownDescription: "The affiliation of the user to this organization. This is not to be understood as a role.",
								Validators: []validator.String{
									stringvalidator.OneOf(
										"owner",
										"member",
										"admin",
									),
								},
							},
							"permissions": schema.ListAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
								Description:         "The editable permissions of the user",
								MarkdownDescription: "The editable permissions of the user",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"id": schema.StringAttribute{
								Computed:            true,
								Optional:            true,
								Description:         "The UUID of the user",
								MarkdownDescription: "The UUID of the user",
							},
							"email": schema.StringAttribute{
								Computed:            true,
								Optional:            true,
								Description:         "The email address of the user.",
								MarkdownDescription: "The email address of the user.",
							},
							"membership_type": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "The type of the membership.",
								MarkdownDescription: "The type of the membership.",
								Default:             stringdefault.StaticString("user"),
							},
						},
						Optional: true,
					},
				},
			},
		},
	}
}

type OrganizationMembershipModel struct {
	Id             types.String          `tfsdk:"id"`
	OrganizationId types.String          `tfsdk:"organization_id"`
	Membership     basetypes.ObjectValue `tfsdk:"membership"`
}

type MembershipValue struct {
	ServiceAccountMembership basetypes.ObjectValue `tfsdk:"service_account_membership"`
	UserMembership           basetypes.ObjectValue `tfsdk:"user_membership"`
}

type ServiceAccountMembershipValue struct {
	Affiliation    types.String `tfsdk:"affiliation"`
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	MembershipType types.String `tfsdk:"membership_type"`
	Permissions    types.List   `tfsdk:"permissions"`
}

type UserMembershipValue struct {
	Affiliation    types.String `tfsdk:"affiliation"`
	Id             types.String `tfsdk:"id"`
	Email          types.String `tfsdk:"email"`
	MembershipType types.String `tfsdk:"membership_type"`
	Permissions    types.List   `tfsdk:"permissions"`
}
