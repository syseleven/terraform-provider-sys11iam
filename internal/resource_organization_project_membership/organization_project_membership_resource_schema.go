package resource_organization_project_membership

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func OrganizationProjectMembershipResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
				Description:         "The unique identifier of the member(user, service account) in the project.",
				MarkdownDescription: "The unique identifier of the member(user, service account) in the project.",
			},
			"org_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"organization_id": compat.DeprecatedOrganizationIdAttribute(),
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
								Required: true,
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Required:            true,
										PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
										Description:         "The UUID of the service account.",
										MarkdownDescription: "The UUID of the service account.",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
										Description:         "The unique name of the service account.",
										MarkdownDescription: "The unique name of the service account.",
									},
								},
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
								Required: true,
								Attributes: map[string]schema.Attribute{
									"email": schema.StringAttribute{
										Required:            true,
										PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
										Description:         "The email address of the user.",
										MarkdownDescription: "The email address of the user.",
									},
									"id": schema.StringAttribute{
										Computed:            true,
										PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
										Description:         "The UUID of the user.",
										MarkdownDescription: "The UUID of the user.",
									},
								},
							},
						},
						Optional: true,
					},
				},
			},
		},
	}
}

func UserAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"email": types.StringType,
		"id":    types.StringType,
	}
}

func UserMembershipAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"membership_type": types.StringType,
		"permissions":     types.ListType{ElemType: types.StringType},
		"user":            basetypes.ObjectType{AttrTypes: UserAttrTypes()},
	}
}

func ServiceAccountAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
}

func ServiceAccountMembershipAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"membership_type": types.StringType,
		"permissions":     types.ListType{ElemType: types.StringType},
		"service_account": basetypes.ObjectType{AttrTypes: ServiceAccountAttrTypes()},
	}
}

// MembershipAttrTypes returns the attribute types of the membership object.
func MembershipAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"service_account_membership": basetypes.ObjectType{AttrTypes: ServiceAccountMembershipAttrTypes()},
		"user_membership":            basetypes.ObjectType{AttrTypes: UserMembershipAttrTypes()},
	}
}

type OrganizationProjectMembershipModel struct {
	Id             types.String          `tfsdk:"id"`
	OrgId          types.String          `tfsdk:"org_id"`
	OrganizationId types.String          `tfsdk:"organization_id"`
	ProjectId      types.String          `tfsdk:"project_id"`
	ProjectName    types.String          `tfsdk:"project_name"`
	Membership     basetypes.ObjectValue `tfsdk:"membership"`
}

// MembershipValue is the decoded representation of the membership object.
// The nested blocks are ObjectValue so that partially unknown plans (for
// example a computed user.id that is "known after apply") can be decoded
// without a Value Conversion Error.
type MembershipValue struct {
	ServiceAccountMembership basetypes.ObjectValue `tfsdk:"service_account_membership"`
	UserMembership           basetypes.ObjectValue `tfsdk:"user_membership"`
}

type ServiceAccountMembershipValue struct {
	MembershipType types.String          `tfsdk:"membership_type"`
	Permissions    types.List            `tfsdk:"permissions"`
	ServiceAccount basetypes.ObjectValue `tfsdk:"service_account"`
}

type ServiceAccountValue struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type UserMembershipValue struct {
	MembershipType types.String          `tfsdk:"membership_type"`
	Permissions    types.List            `tfsdk:"permissions"`
	User           basetypes.ObjectValue `tfsdk:"user"`
}

type UserValue struct {
	Email types.String `tfsdk:"email"`
	Id    types.String `tfsdk:"id"`
}
