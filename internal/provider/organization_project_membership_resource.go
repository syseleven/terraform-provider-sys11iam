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

func (r *ProjectMembershipResource) buildData(data *resource_organization_project_membership.OrganizationProjectMembershipModel, response *iam.IAMProjectMembership) {
	if response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(response.ServiceAccount.ID)
		data.Membership.ServiceAccountMembership = &resource_organization_project_membership.ServiceAccountMembershipValue{
			MembershipType: types.StringValue(response.MembershipType),
			Permissions: types.ListValueMust(types.StringType, convertSliceToAttrValues(response.Permissions, func(s string) attr.Value {
				return types.StringValue(s)
			})),
			ServiceAccount: &resource_organization_project_membership.ServiceAccountValue{
				Id:   types.StringValue(response.ServiceAccount.ID),
				Name: types.StringValue(response.ServiceAccount.Name),
			},
		}
	} else if response.User.ID != "" {
		data.Id = types.StringValue(response.User.ID)
		data.Membership.UserMembership = &resource_organization_project_membership.UserMembershipValue{
			MembershipType: types.StringValue(response.MembershipType),
			Permissions: types.ListValueMust(types.StringType, convertSliceToAttrValues(response.Permissions, func(s string) attr.Value {
				return types.StringValue(s)
			})),
			User: &resource_organization_project_membership.UserValue{
				Email: types.StringValue(response.User.Email),
				Id:    types.StringValue(response.User.ID),
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

	var permissions []string
	var membershipType string

	var org_membership_response iam.IAMOrganizationMembership

	if data.Membership.UserMembership != nil {
		diags := data.Membership.UserMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		membershipType = data.Membership.UserMembership.MembershipType.ValueString()

		// Is the e-mail already a member?
		email := data.Membership.UserMembership.User.Email.ValueString()

		org_membership_response, err = r.client.GetOrganizationMembershipByEmail(ctx, data.OrganizationId.ValueString(), email)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// Is the e-mail at least invited?
				_, err := r.client.GetOrganizationInvitationByEmail(data.OrganizationId.ValueString(), email)
				if err != nil {
					// Invite the e-mail
					_, err := r.client.CreateOrganizationInvitation(data.OrganizationId.ValueString(), email, permissions)
					if err != nil {
						resp.Diagnostics.AddError("", err.Error())
						return
					}

					// The email is invited, but has to be activated manually
					resp.Diagnostics.AddError("InvitationNotAcceptedError",
						fmt.Sprintf("Can not create ProjectMembership in project with id %s in organization with id %s as the user with the e-mail %s has not yet accepted the invitation. Invitation accepting is a manual step, please contact the invited user.",
							data.OrganizationId.ValueString(), data.ProjectId.ValueString(), email))
					return
				}
			} else {
				resp.Diagnostics.AddError("", err.Error())
				return
			}
		}

	} else if data.Membership.ServiceAccountMembership != nil {
		diags := data.Membership.ServiceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		membershipType = data.Membership.ServiceAccountMembership.MembershipType.ValueString()

		org_membership_response, err = r.client.GetOrganizationMembership(data.OrganizationId.ValueString(), data.Membership.ServiceAccountMembership.ServiceAccount.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}
	}

	if org_membership_response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(org_membership_response.ServiceAccount.ID)
	} else if org_membership_response.User.ID != "" {
		data.Id = types.StringValue(org_membership_response.User.ID)
	}

	response, err := r.client.CreateProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString(), membershipType, permissions)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.ProjectId = types.StringValue(response.Project.ID)
	data.ProjectName = types.StringValue(response.Project.Name)
	r.buildData(&data, &response)

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
	response, err := r.client.GetProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.ProjectId = types.StringValue(response.Project.ID)
	data.ProjectName = types.StringValue(response.Project.Name)
	r.buildData(&data, &response)

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

	var permissions []string
	var membershipType string
	if len(data.Membership.UserMembership.Permissions.Elements()) > 0 {
		diags := data.Membership.UserMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		membershipType = data.Membership.UserMembership.MembershipType.ValueString()
	} else if len(data.Membership.ServiceAccountMembership.Permissions.Elements()) > 0 {
		diags := data.Membership.ServiceAccountMembership.Permissions.ElementsAs(ctx, &permissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		membershipType = data.Membership.ServiceAccountMembership.MembershipType.ValueString()
	}

	response, err := r.client.UpdateProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString(), membershipType, permissions)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data.ProjectId = types.StringValue(response.Project.ID)
	data.ProjectName = types.StringValue(response.Project.Name)

	r.buildData(&data, &response)

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
	tflog.Info(ctx, "Deleting ProjectMembership resource.")
	err := r.client.DeleteProjectMembership(data.OrganizationId.ValueString(), data.ProjectId.ValueString(), data.Id.ValueString())
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
	response, err := r.client.GetProjectMembership(idParts[0], idParts[1], idParts[2])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_project_membership.OrganizationProjectMembershipModel
	data.Membership = &resource_organization_project_membership.MembershipValue{}

	// Data value setting
	data.ProjectId = types.StringValue(response.Project.ID)
	data.OrganizationId = types.StringValue(idParts[0])
	data.ProjectName = types.StringValue(response.Project.Name)

	r.buildData(&data, &response)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
