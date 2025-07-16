package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_membership"
)

var _ resource.Resource = (*ProjectMembershipResource)(nil)
var _ resource.ResourceWithConfigure = (*ProjectMembershipResource)(nil)

func NewProjectMembershipResource() resource.Resource {
	return &ProjectMembershipResource{}
}

type ProjectMembershipResource struct {
	client iam.IAMClient
}

func (r *ProjectMembershipResource) createMembershipAttrTypes(membershipType string) map[string]attr.Type {
	baseTypes := map[string]attr.Type{
		"affiliation":     types.StringType,
		"permissions":     types.ListType{ElemType: types.StringType},
		"id":              types.StringType,
		"membership_type": types.StringType,
		"organization_id": types.StringType,
	}

	memberAttrTypes := map[string]map[string]attr.Type{
		user: {
			"id":    types.StringType,
			"email": types.StringType,
		},
		serviceAccount: {
			"id":          types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		},
	}

	if attrs, exists := memberAttrTypes[membershipType]; exists {
		baseTypes[membershipType] = types.ObjectType{AttrTypes: attrs}
	}

	return baseTypes
}

func (r *ProjectMembershipResource) createMembershipAttrValues(membership *iam.IAMProjectMembership, membershipType string) map[string]attr.Value {
	memberAttrTypes := map[string]map[string]attr.Type{
		user: {
			"id":    types.StringType,
			"email": types.StringType,
		},
		serviceAccount: {
			"id":          types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		},
		"project": {
			"id":          types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"tags":        types.ListType{ElemType: types.StringType},
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		},
	}

	baseMembershipAttrValues := func(membership *iam.IAMProjectMembership, membershipType string, memberAttrTypes map[string]map[string]attr.Type) map[string]attr.Value {
		return map[string]attr.Value{
			"membership_type": types.StringValue(membershipType),
			"permissions": types.ListValueMust(types.StringType, convertSliceToAttrValues(membership.Permissions, func(permission string) attr.Value {
				return types.StringValue(permission)
			})),
			"project": resource_organization_project_membership.NewProjectValueMust(memberAttrTypes["project"], map[string]attr.Value{
				"id":          types.StringValue(membership.Project.ID),
				"name":        types.StringValue(membership.Project.Name),
				"description": types.StringValue(membership.Project.Description),
				"tags": types.ListValueMust(types.StringType, convertSliceToAttrValues(membership.Project.Tags, func(tag string) attr.Value {
					return types.StringValue(tag)
				})),
				"created_at": types.StringValue(membership.Project.CreatedAt),
				"updated_at": types.StringValue(membership.Project.UpdatedAt),
			}),
		}
	}

	switch membershipType {
	case user:
		memberValue := resource_organization_project_membership.NewUserValueMust(memberAttrTypes[membershipType], map[string]attr.Value{
			"id":    types.StringValue(membership.User.ID),
			"email": types.StringValue(membership.User.Email),
		})

		result := baseMembershipAttrValues(membership, membershipType, memberAttrTypes)
		result[user] = memberValue
		return result

	case serviceAccount:
		memberValue := resource_organization_project_membership.NewServiceAccountValueMust(memberAttrTypes[membershipType], map[string]attr.Value{
			"id":          types.StringValue(membership.ServiceAccount.ID),
			"name":        types.StringValue(membership.ServiceAccount.Name),
			"description": types.StringValue(membership.ServiceAccount.Description),
			"created_at":  types.StringValue(membership.ServiceAccount.CreatedAt),
			"updated_at":  types.StringValue(membership.ServiceAccount.UpdatedAt),
		})

		result := baseMembershipAttrValues(membership, membershipType, memberAttrTypes)
		result[serviceAccount] = memberValue
		return result
	default:
		return nil
	}
}

func (r *ProjectMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_project_membership"
}

func (r *ProjectMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_project_membership.OrganizationProjectMembershipResourceSchema(ctx)
}

func (r *ProjectMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_project_membership.OrganizationProjectMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Info(ctx, "Creating ProjectMembership resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))

	// Is the organization active?
	org_response, err := r.client.GetOrganization(data.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create ProjectMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				data.OrganizationId.ValueString()))
		return
	}

	var elements []string
	var org_membership_response iam.IAMOrganizationMembership

	if !data.UserMembership.IsNull() || !data.UserMembership.IsUnknown() {
		diags := data.UserMembership.Permissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Is the e-mail already a member?
		var user map[string]string
		diags = data.UserMembership.User.As(ctx, &user, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		email, ok := user["email"]
		if !ok || email == "" {
			resp.Diagnostics.AddError("InvalidUserEmailError",
				"User email is not set or invalid. Please provide a valid user email.")
			return
		}

		org_membership_response, err = r.client.GetOrganizationMembershipByEmail(data.OrganizationId.ValueString(), email)
		if err != nil {
			// Is the e-mail at least invited?
			_, err := r.client.GetOrganizationInvitationByEmail(data.OrganizationId.ValueString(), email)
			if err != nil {
				// Invite the e-mail
				_, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), email, elements)
				if err != nil {
					resp.Diagnostics.AddError("", err.Error())
					return
				}
			}
			// The email is invited, but has to be activated manually
			resp.Diagnostics.AddError("InvitationNotAcceptedError",
				fmt.Sprintf("Can not create ProjectMembership in project with id %s in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
					data.OrganizationId.ValueString(), data.ProjectId.ValueString(), email))
			return
		}
	} else if !data.ServiceAccountMembership.IsNull() || !data.ServiceAccountMembership.IsUnknown() {
		diags := data.ServiceAccountMembership.Permissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var serviceAccount map[string]string
		diags = data.ServiceAccountMembership.ServiceAccount.As(ctx, &serviceAccount, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		id, ok := serviceAccount["id"]
		if !ok || id == "" {
			resp.Diagnostics.AddError("InvalidServiceAccountIdError",
				"Service account ID is not set or invalid. Please provide a valid service account ID.")
			return
		}

		org_membership_response, err = r.client.GetOrganizationMembership(data.OrganizationId.ValueString(), id)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	if org_membership_response.ServiceAccount.ID != "" {
		data.MemberId = types.StringValue(org_membership_response.ServiceAccount.ID)
	}
	if org_membership_response.User.ID != "" {
		data.MemberId = types.StringValue(org_membership_response.User.ID)
	}

	response, err := r.client.CreateProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.MemberId.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.ProjectId = types.StringValue(response.ProjectId)
	if response.ServiceAccount.ID != "" {
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
		data.ServiceAccountMembership = resource_organization_project_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(&response, serviceAccount))
	}
	if response.User.ID != "" {
		data.MemberId = types.StringValue(response.User.ID)
		data.UserMembership = resource_organization_project_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(&response, user))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_project_membership.OrganizationProjectMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectMembership resource.")
	response, err := r.client.GetProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.MemberId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	if response.ServiceAccount.ID != "" {
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
		data.ServiceAccountMembership = resource_organization_project_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(&response, serviceAccount))
	}
	if response.User.ID != "" {
		data.MemberId = types.StringValue(response.User.ID)
		data.UserMembership = resource_organization_project_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(&response, user))
	}
	data.ProjectId = types.StringValue(response.ProjectId)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_project_membership.OrganizationProjectMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.MemberId)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Creating ProjectMembership resource.")

	var elements []string
	if !data.UserMembership.IsNull() || !data.UserMembership.IsUnknown() {
		diags := data.UserMembership.Permissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else if !data.ServiceAccountMembership.IsNull() || !data.ServiceAccountMembership.IsUnknown() {
		diags := data.ServiceAccountMembership.Permissions.ElementsAs(ctx, &elements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	response, err := r.client.UpdateProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.MemberId.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.ProjectId = types.StringValue(response.ProjectId)
	data.OrganizationId = types.StringValue("PLEASE SET ME")
	if response.ServiceAccount.ID != "" {
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
		data.ServiceAccountMembership = resource_organization_project_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(&response, serviceAccount))
	}
	if response.User.ID != "" {
		data.MemberId = types.StringValue(response.User.ID)
		data.UserMembership = resource_organization_project_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(&response, user))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_project_membership.OrganizationProjectMembershipModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Info(ctx, "Reading ProjectMembership resource.")
	err := r.client.DeleteProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.MemberId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *ProjectMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,project_id,member_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading ProjectMembership resource.")
	response, err := r.client.GetProjectMembership(idParts[0], idParts[1], idParts[2])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_project_membership.OrganizationProjectMembershipModel

	// Data value setting
	data.ProjectId = types.StringValue(response.ProjectId)
	data.OrganizationId = types.StringValue("PLEASE SET ME")
	if response.ServiceAccount.ID != "" {
		data.MemberId = types.StringValue(response.ServiceAccount.ID)
		data.ServiceAccountMembership = resource_organization_project_membership.NewServiceAccountMembershipValueMust(r.createMembershipAttrTypes(serviceAccount), r.createMembershipAttrValues(&response, serviceAccount))
	}
	if response.User.ID != "" {
		data.MemberId = types.StringValue(response.User.ID)
		data.UserMembership = resource_organization_project_membership.NewUserMembershipValueMust(r.createMembershipAttrTypes(user), r.createMembershipAttrValues(&response, user))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
