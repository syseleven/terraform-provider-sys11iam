package resource_organization_project_membership

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func OrganizationProjectMembershipResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Description:         "The unique identifier of the member(user, service account) in the project.",
				MarkdownDescription: "The unique identifier of the member(user, service account) in the project.",
			},
			"organization_id": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[a-f0-9]{8}[a-f0-9]{4}4[a-f0-9]{3}[89ab][a-f0-9]{3}[a-f0-9]{12}$"), ""),
				},
			},
			"project_name": schema.StringAttribute{
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
							"membership_type": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "The type of the membership.",
								MarkdownDescription: "The type of the membership.",
								Default:             stringdefault.StaticString("service_account"),
							},
							"permissions": schema.ListAttribute{
								ElementType:         types.StringType,
								Required:            true,
								Description:         "The permissions of the service account",
								MarkdownDescription: "The permissions of the service account",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"service_account": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:            true,
										Optional:            true,
										Description:         "The UUID of the service account.",
										MarkdownDescription: "The UUID of the service account.",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										Description:         "The unique name of the service account.",
										MarkdownDescription: "The unique name of the service account.",
									},
								},
								Computed: true,
							},
						},
						Optional: true,
					},
					"user_membership": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"membership_type": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "The type of the membership.",
								MarkdownDescription: "The type of the membership.",
								Default:             stringdefault.StaticString("user"),
							},
							"permissions": schema.ListAttribute{
								ElementType:         types.StringType,
								Required:            true,
								Description:         "The permissions of the user",
								MarkdownDescription: "The permissions of the user",
								Validators: []validator.List{
									listvalidator.UniqueValues(),
								},
							},
							"user": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"email": schema.StringAttribute{
										Computed:            true,
										Optional:            true,
										Description:         "The email address of the user.",
										MarkdownDescription: "The email address of the user.",
									},
									"id": schema.StringAttribute{
										Computed:            true,
										Description:         "The UUID of the user.",
										MarkdownDescription: "The UUID of the user.",
									},
								},
								Computed: true,
							},
						},
						Optional: true,
					},
				},
			},
		},
	}
}

type OrganizationProjectMembershipModel struct {
	Id             types.String     `tfsdk:"id"`
	OrganizationId types.String     `tfsdk:"organization_id"`
	ProjectId      types.String     `tfsdk:"project_id"`
	ProjectName    types.String     `tfsdk:"project_name"`
	Membership     *MembershipValue `tfsdk:"membership"`
}

type MembershipValue struct {
	ServiceAccountMembership *ServiceAccountMembershipValue `tfsdk:"service_account_membership"`
	UserMembership           *UserMembershipValue           `tfsdk:"user_membership"`
}

type ServiceAccountMembershipValue struct {
	MembershipType types.String         `tfsdk:"membership_type"`
	Permissions    types.List           `tfsdk:"permissions"`
	ServiceAccount *ServiceAccountValue `tfsdk:"service_account"`
}
type ServiceAccountValue struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
type UserMembershipValue struct {
	MembershipType types.String `tfsdk:"membership_type"`
	Permissions    types.List   `tfsdk:"permissions"`
	User           *UserValue   `tfsdk:"user"`
}
type UserValue struct {
	Email types.String `tfsdk:"email"`
	Id    types.String `tfsdk:"id"`
}
