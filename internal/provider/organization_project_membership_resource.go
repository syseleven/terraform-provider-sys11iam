package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

			data := resource_organization_project_membership.OrganizationProjectMembershipModel{
				Id:             id,
				OrgId:          orgId,
				OrganizationId: organizationId,
				ProjectId:      projectId,
				ProjectName:    types.StringNull(),
				Membership: &resource_organization_project_membership.MembershipValue{
					UserMembership: &resource_organization_project_membership.UserMembershipValue{
						MembershipType: types.StringValue("user"),
						Permissions:    permissions,
						User: &resource_organization_project_membership.UserValue{
							Email: email,
							Id:    id,
						},
					},
				},
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

// buildDataFromV3 populates the Terraform model from the v3 project membership list entry.
func (r *ProjectMembershipResource) buildDataFromV3(data *resource_organization_project_membership.OrganizationProjectMembershipModel, member *iam.IAMProjectMembershipV3) {
	permissions := iam.FilterActiveDirectPermissions(member.Permissions)

	if member.Type == "service_account" {
		data.Id = types.StringValue(member.ID)
		data.Membership.ServiceAccountMembership = &resource_organization_project_membership.ServiceAccountMembershipValue{
			MembershipType: types.StringValue(member.Type),
			Permissions: types.ListValueMust(types.StringType, convertSliceToAttrValues(permissions, func(s string) attr.Value {
				return types.StringValue(s)
			})),
			ServiceAccount: &resource_organization_project_membership.ServiceAccountValue{
				Id:   types.StringValue(member.ID),
				Name: types.StringValue(member.DisplayName),
			},
		}
	} else if member.Type == "user" {
		data.Id = types.StringValue(member.ID)
		data.Membership.UserMembership = &resource_organization_project_membership.UserMembershipValue{
			MembershipType: types.StringValue(member.Type),
			Permissions: types.ListValueMust(types.StringType, convertSliceToAttrValues(permissions, func(s string) attr.Value {
				return types.StringValue(s)
			})),
			User: &resource_organization_project_membership.UserValue{
				Email: types.StringValue(member.DisplayName),
				Id:    types.StringValue(member.ID),
			},
		}
	}
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

	var permissions []string
	var membershipType string

	if data.Membership.UserMembership != nil {
		diags := data.Membership.UserMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		membershipType = data.Membership.UserMembership.MembershipType.ValueString()

		// Is the e-mail already a member?
		email := data.Membership.UserMembership.User.Email.ValueString()

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

		data.Id = types.StringValue(orgMember.ID)

	} else if data.Membership.ServiceAccountMembership != nil {
		diags := data.Membership.ServiceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		membershipType = data.Membership.ServiceAccountMembership.MembershipType.ValueString()

		// Verify the service account exists in the org
		_, err = r.client.GetOrgMembershipV3(orgId.ValueString(), data.Membership.ServiceAccountMembership.ServiceAccount.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
		data.Id = types.StringValue(data.Membership.ServiceAccountMembership.ServiceAccount.Id.ValueString())
	}

	// Write permissions via v3 endpoints
	projectId := data.ProjectId.ValueString()
	memberId := data.Id.ValueString()

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

	// Data value setting
	data.ProjectId = types.StringValue(projectId)
	r.buildDataFromV3(&data, &member)
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

	// Data value setting
	r.buildDataFromV3(&data, &member)
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_project_membership.OrganizationProjectMembershipModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating ProjectMembership resource.")

	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	var permissions []string
	var membershipType string
	if data.Membership.UserMembership != nil && len(data.Membership.UserMembership.Permissions.Elements()) > 0 {
		diags := data.Membership.UserMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		membershipType = data.Membership.UserMembership.MembershipType.ValueString()
	} else if data.Membership.ServiceAccountMembership != nil && len(data.Membership.ServiceAccountMembership.Permissions.Elements()) > 0 {
		diags := data.Membership.ServiceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		membershipType = data.Membership.ServiceAccountMembership.MembershipType.ValueString()
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

	// Data value setting
	data.ProjectId = types.StringValue(projectId)
	r.buildDataFromV3(&data, &member)
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

	projectId := data.ProjectId.ValueString()
	memberId := data.Id.ValueString()

	var err error
	if data.Membership.UserMembership != nil {
		_, err = r.client.PutUserProjectPermissions(orgId.ValueString(), projectId, memberId, []string{})
	} else if data.Membership.ServiceAccountMembership != nil {
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
	data.Membership = &resource_organization_project_membership.MembershipValue{}

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

	r.buildDataFromV3(&data, &member)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
