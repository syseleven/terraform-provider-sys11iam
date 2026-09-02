package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_project_membership"
)

var _ resource.Resource = (*ProjectMembershipResource)(nil)
var _ resource.ResourceWithConfigure = (*ProjectMembershipResource)(nil)
var _ resource.ResourceWithMoveState = (*ProjectMembershipResource)(nil)
var _ resource.ResourceWithUpgradeState = (*ProjectMembershipResource)(nil)
var _ resource.ResourceWithValidateConfig = (*ProjectMembershipResource)(nil)

func NewProjectMembershipResource() resource.Resource {
	return &ProjectMembershipResource{}
}

type ProjectMembershipResource struct {
	client iam.IAMClient
}

func (r *ProjectMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_project_membership"
}

func (r *ProjectMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_project_membership.OrganizationProjectMembershipResourceSchema(ctx)
}

func (r *ProjectMembershipResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: compat.OrgIdStateUpgrader(),
	}
}

func (r *ProjectMembershipResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		compat.RawStateMover("sys11iam_project_membership", func(ctx context.Context, rawState compat.RawState, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			orgId, organizationId, err := compat.RawOrgIDs(rawState)
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read organization identifier: "+err.Error())
				return
			}

			id, err := compat.RawString(rawState, "id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project membership id: "+err.Error())
				return
			}

			email, err := compat.RawString(rawState, "email")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project membership email: "+err.Error())
				return
			}
			if email.IsNull() || email.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Unsupported project membership state",
					"Cannot safely migrate sys11iam_project_membership state without an email value. This is likely a service account membership; import it as sys11iam_organization_project_membership instead.",
				)
				return
			}

			permissions, err := compat.RawStringList(ctx, rawState, "permissions")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project membership permissions: "+err.Error())
				return
			}

			projectId, err := compat.RawString(rawState, "project_id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project id: "+err.Error())
				return
			}

			user, diags := basetypes.NewObjectValue(resource_organization_project_membership.UserAttrTypes(), map[string]attr.Value{
				"email": email,
				"id":    id,
			})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}

			userMembership, diags := basetypes.NewObjectValue(resource_organization_project_membership.UserMembershipAttrTypes(), map[string]attr.Value{
				"membership_type": types.StringValue("user"),
				"permissions":     permissions,
				"user":            user,
			})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}

			membership, diags := basetypes.NewObjectValue(resource_organization_project_membership.MembershipAttrTypes(), map[string]attr.Value{
				"user_membership":            userMembership,
				"service_account_membership": basetypes.NewObjectNull(resource_organization_project_membership.ServiceAccountMembershipAttrTypes()),
			})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}

			data := resource_organization_project_membership.OrganizationProjectMembershipModel{
				Id:             id,
				OrgId:          orgId,
				OrganizationId: organizationId,
				ProjectId:      projectId,
				ProjectName:    types.StringNull(),
				Membership:     membership,
			}

			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &data)...)
		}),
	}
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

func (r *ProjectMembershipResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	compat.ValidateOrgId(ctx, req.Config, resp)
}

// permissionsAttrList converts a slice of permission names into a types.List.
func permissionsAttrList(permissions []string) types.List {
	return types.ListValueMust(types.StringType, convertSliceToAttrValues(permissions, func(s string) attr.Value {
		return types.StringValue(s)
	}))
}

// fetchProjectName returns the project name for state population. Failures
// degrade to a null name instead of failing the whole operation.
func (r *ProjectMembershipResource) fetchProjectName(ctx context.Context, orgId string, projectId string) types.String {
	project, err := r.client.GetProject(orgId, projectId)
	if err != nil {
		tflog.Error(ctx, "Could not fetch project for project_name", map[string]any{
			"project_id": projectId,
			"error":      err.Error(),
		})
		return types.StringNull()
	}
	return types.StringValue(project.Name)
}

// buildDataFromV3 populates the Terraform model from the v3 project membership list entry.
// The provided permissions value is stored verbatim: Read passes the API read-back while
// Create and Update pass the planned (configured) permissions so state matches the config.
func (r *ProjectMembershipResource) buildDataFromV3(ctx context.Context, data *resource_organization_project_membership.OrganizationProjectMembershipModel, member *iam.IAMProjectMembershipV3, projectName types.String, permissions types.List) diag.Diagnostics {
	var diags diag.Diagnostics

	userMembershipAttrTypes := resource_organization_project_membership.UserMembershipAttrTypes()
	serviceAccountMembershipAttrTypes := resource_organization_project_membership.ServiceAccountMembershipAttrTypes()
	membershipAttrTypes := resource_organization_project_membership.MembershipAttrTypes()

	if member.Type == "service_account" {
		serviceAccount, d := basetypes.NewObjectValue(resource_organization_project_membership.ServiceAccountAttrTypes(), map[string]attr.Value{
			"id":   types.StringValue(member.ID),
			"name": types.StringValue(member.DisplayName),
		})
		diags.Append(d...)

		serviceAccountMembership, d := basetypes.NewObjectValue(serviceAccountMembershipAttrTypes, map[string]attr.Value{
			"membership_type": types.StringValue(member.Type),
			"permissions":     permissions,
			"service_account": serviceAccount,
		})
		diags.Append(d...)

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"user_membership":            basetypes.NewObjectNull(userMembershipAttrTypes),
			"service_account_membership": serviceAccountMembership,
		})
	} else if member.Type == "user" {
		user, d := basetypes.NewObjectValue(resource_organization_project_membership.UserAttrTypes(), map[string]attr.Value{
			"email": types.StringValue(member.DisplayName),
			"id":    types.StringValue(member.ID),
		})
		diags.Append(d...)

		userMembership, d := basetypes.NewObjectValue(userMembershipAttrTypes, map[string]attr.Value{
			"membership_type": types.StringValue(member.Type),
			"permissions":     permissions,
			"user":            user,
		})
		diags.Append(d...)

		data.Membership = basetypes.NewObjectValueMust(membershipAttrTypes, map[string]attr.Value{
			"user_membership":            userMembership,
			"service_account_membership": basetypes.NewObjectNull(serviceAccountMembershipAttrTypes),
		})
	}

	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(member.ID)
	data.ProjectName = projectName
	return diags
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

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", orgId.ValueString()))

	// Is the organization active?
	org_response, err := r.client.GetOrganization(orgId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	if !org_response.IsActive {
		resp.Diagnostics.AddError("OrganizationNotActiveError",
			fmt.Sprintf("Can not create ProjectMembership in organization with id %s as it is not active. Organization activation is a manual step, please contact an IAM administrator.",
				orgId.ValueString()))
		return
	}

	var membership resource_organization_project_membership.MembershipValue
	resp.Diagnostics.Append(data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []string
	var membershipType string
	var memberId string

	if !membership.UserMembership.IsNull() && !membership.UserMembership.IsUnknown() {
		var userMembership resource_organization_project_membership.UserMembershipValue
		resp.Diagnostics.Append(membership.UserMembership.As(ctx, &userMembership, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(userMembership.Permissions.ElementsAs(ctx, &permissions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		membershipType = userMembership.MembershipType.ValueString()

		if userMembership.User.IsNull() || userMembership.User.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("membership").AtName("user_membership").AtName("user"),
				"Missing user",
				"The user block within user_membership is required to create a project membership.",
			)
			return
		}

		var user resource_organization_project_membership.UserValue
		resp.Diagnostics.Append(userMembership.User.As(ctx, &user, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		if user.Email.IsNull() || user.Email.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("membership").AtName("user_membership").AtName("user").AtName("email"),
				"Missing user email",
				"The email address of the user within user_membership.user is required to create a project membership.",
			)
			return
		}

		// Is the e-mail already a member?
		email := user.Email.ValueString()

		orgMember, err := r.client.GetOrgMembershipV3ByEmail(orgId.ValueString(), email)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// Is the e-mail at least invited?
				_, err := r.client.GetOrganizationInvitationByEmail(orgId.ValueString(), email)
				if err != nil {
					// Invite the e-mail
					_, err := r.client.CreateOrganizationInvitation(orgId.ValueString(), email, permissions)
					if err != nil {
						resp.Diagnostics.AddError("", err.Error())
						return
					}

					// The email is invited, but has to be activated manually
					resp.Diagnostics.AddError("InvitationNotAcceptedError",
						fmt.Sprintf("Can not create ProjectMembership in project with id %s in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
							orgId.ValueString(), data.ProjectId.ValueString(), email))
					return
				}
			} else {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
		}

		if !data.Id.IsNull() && !data.Id.IsUnknown() && data.Id.ValueString() != "" && data.Id.ValueString() != orgMember.ID {
			resp.Diagnostics.AddError("MismatchedMemberId",
				fmt.Sprintf("The configured id %s does not match the organization membership id %s for the user with the e-mail %s.",
					data.Id.ValueString(), orgMember.ID, email))
			return
		}
		if data.Id.IsNull() || data.Id.IsUnknown() || data.Id.ValueString() == "" {
			data.Id = types.StringValue(orgMember.ID)
		}
		memberId = data.Id.ValueString()
	} else if !membership.ServiceAccountMembership.IsNull() && !membership.ServiceAccountMembership.IsUnknown() {
		var serviceAccountMembership resource_organization_project_membership.ServiceAccountMembershipValue
		resp.Diagnostics.Append(membership.ServiceAccountMembership.As(ctx, &serviceAccountMembership, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(serviceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		membershipType = serviceAccountMembership.MembershipType.ValueString()

		if serviceAccountMembership.ServiceAccount.IsNull() || serviceAccountMembership.ServiceAccount.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("membership").AtName("service_account_membership").AtName("service_account"),
				"Missing service account",
				"The service_account block within service_account_membership is required to create a project membership.",
			)
			return
		}

		var serviceAccount resource_organization_project_membership.ServiceAccountValue
		resp.Diagnostics.Append(serviceAccountMembership.ServiceAccount.As(ctx, &serviceAccount, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		if serviceAccount.Id.IsNull() || serviceAccount.Id.IsUnknown() || serviceAccount.Id.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("membership").AtName("service_account_membership").AtName("service_account").AtName("id"),
				"Missing service account id",
				"The service account UUID within service_account_membership.service_account is required to create a project membership.",
			)
			return
		}

		memberId = serviceAccount.Id.ValueString()

		// Verify the service account exists in the org
		_, err = r.client.GetOrgMembershipV3(orgId.ValueString(), memberId)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}

		if !data.Id.IsNull() && !data.Id.IsUnknown() && data.Id.ValueString() != "" && data.Id.ValueString() != memberId {
			resp.Diagnostics.AddError("MismatchedMemberId",
				fmt.Sprintf("The configured id %s does not match the service account id %s.",
					data.Id.ValueString(), memberId))
			return
		}
		if data.Id.IsNull() || data.Id.IsUnknown() || data.Id.ValueString() == "" {
			data.Id = types.StringValue(memberId)
		}
	} else {
		resp.Diagnostics.AddError("MissingMembership",
			"The membership block must contain either a user_membership or a service_account_membership block.")
		return
	}

	// Write permissions via v3 endpoints
	projectId := data.ProjectId.ValueString()

	if membershipType == "user" {
		_, err = r.client.PutUserProjectPermissions(orgId.ValueString(), projectId, memberId, permissions)
	} else if membershipType == "service_account" {
		_, err = r.client.PutServiceAccountProjectPermissions(orgId.ValueString(), projectId, memberId, permissions)
	}
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Read back to populate state
	member, err := r.client.GetProjectMembershipV3(orgId.ValueString(), projectId, memberId)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting — the configured permissions win over the API read-back
	data.ProjectId = types.StringValue(projectId)
	resp.Diagnostics.Append(r.buildDataFromV3(ctx, &data, &member, r.fetchProjectName(ctx, orgId.ValueString(), projectId), permissionsAttrList(permissions))...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

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

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	member, err := r.client.GetProjectMembershipV3(orgId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// The API read-back is the source of truth during refresh
	apiPermissions := permissionsAttrList(iam.FilterActiveDirectPermissions(member.Permissions))

	// Data value setting
	resp.Diagnostics.Append(r.buildDataFromV3(ctx, &data, &member, r.fetchProjectName(ctx, orgId.ValueString(), data.ProjectId.ValueString()), apiPermissions)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_project_membership.OrganizationProjectMembershipModel
	var stateData resource_organization_project_membership.OrganizationProjectMembershipModel

	// Read Terraform plan data and prior state into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating ProjectMembership resource.")

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	var stateMembership resource_organization_project_membership.MembershipValue
	resp.Diagnostics.Append(stateData.Membership.As(ctx, &stateMembership, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planMembership resource_organization_project_membership.MembershipValue
	resp.Diagnostics.Append(data.Membership.As(ctx, &planMembership, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []string
	var membershipType string

	// The member identity cannot change in place; it is taken from the prior
	// state while the new permissions are taken from the plan.
	switch {
	case !stateMembership.UserMembership.IsNull() && !stateMembership.UserMembership.IsUnknown():
		membershipType = "user"

		if planMembership.UserMembership.IsNull() || planMembership.UserMembership.IsUnknown() {
			resp.Diagnostics.AddError("MembershipTypeChanged",
				"The member type of a project membership cannot be changed in place. Remove the resource and create a new membership for the other member type.")
			return
		}

		var userMembership resource_organization_project_membership.UserMembershipValue
		resp.Diagnostics.Append(planMembership.UserMembership.As(ctx, &userMembership, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(userMembership.Permissions.ElementsAs(ctx, &permissions, false)...)
	case !stateMembership.ServiceAccountMembership.IsNull() && !stateMembership.ServiceAccountMembership.IsUnknown():
		membershipType = "service_account"

		if planMembership.ServiceAccountMembership.IsNull() || planMembership.ServiceAccountMembership.IsUnknown() {
			resp.Diagnostics.AddError("MembershipTypeChanged",
				"The member type of a project membership cannot be changed in place. Remove the resource and create a new membership for the other member type.")
			return
		}

		var serviceAccountMembership resource_organization_project_membership.ServiceAccountMembershipValue
		resp.Diagnostics.Append(planMembership.ServiceAccountMembership.As(ctx, &serviceAccountMembership, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(serviceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)...)
	default:
		resp.Diagnostics.AddError("MissingMembership",
			"The prior state of the project membership does not contain a user or service account membership; nothing to update.")
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Write permissions via v3 endpoints
	projectId := data.ProjectId.ValueString()
	memberId := data.Id.ValueString()

	var err error
	if membershipType == "user" {
		_, err = r.client.PutUserProjectPermissions(orgId.ValueString(), projectId, memberId, permissions)
	} else if membershipType == "service_account" {
		_, err = r.client.PutServiceAccountProjectPermissions(orgId.ValueString(), projectId, memberId, permissions)
	}
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Read back to populate state
	member, err := r.client.GetProjectMembershipV3(orgId.ValueString(), projectId, memberId)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting — the configured permissions win over the API read-back
	data.ProjectId = types.StringValue(projectId)
	resp.Diagnostics.Append(r.buildDataFromV3(ctx, &data, &member, r.fetchProjectName(ctx, orgId.ValueString(), projectId), permissionsAttrList(permissions))...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

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

	// Delete API call logic — PUT empty permissions
	tflog.Info(ctx, "Deleting ProjectMembership resource.")

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	var membership resource_organization_project_membership.MembershipValue
	resp.Diagnostics.Append(data.Membership.As(ctx, &membership, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectId := data.ProjectId.ValueString()
	memberId := data.Id.ValueString()

	var err error
	if !membership.UserMembership.IsNull() && !membership.UserMembership.IsUnknown() {
		_, err = r.client.PutUserProjectPermissions(orgId.ValueString(), projectId, memberId, []string{})
	} else if !membership.ServiceAccountMembership.IsNull() && !membership.ServiceAccountMembership.IsUnknown() {
		_, err = r.client.PutServiceAccountProjectPermissions(orgId.ValueString(), projectId, memberId, []string{})
	}
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
	tflog.Info(ctx, "Importing ProjectMembership resource.")
	member, err := r.client.GetProjectMembershipV3(idParts[0], idParts[1], idParts[2])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_project_membership.OrganizationProjectMembershipModel
	data.Membership = basetypes.NewObjectNull(resource_organization_project_membership.MembershipAttrTypes())

	// Data value setting
	data.ProjectId = types.StringValue(idParts[1])
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(idParts[0])

	// Fetch project to populate project_name (not included in membership response)
	project, err := r.client.GetProject(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", fmt.Sprintf("could not get project: %s", err.Error()))
		return
	}
	data.ProjectName = types.StringValue(project.Name)

	apiPermissions := permissionsAttrList(iam.FilterActiveDirectPermissions(member.Permissions))

	resp.Diagnostics.Append(r.buildDataFromV3(ctx, &data, &member, data.ProjectName, apiPermissions)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
