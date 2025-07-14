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
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_membership_permission"
)

func convertSliceToAttrValues[T any](slice []T, converter func(T) attr.Value) []attr.Value {
	values := make([]attr.Value, len(slice))
	for i, item := range slice {
		values[i] = converter(item)
	}
	return values
}

const user string = "user"
const serviceAccount string = "service_account"

func createMembershipAttrTypes(membershipType string) map[string]attr.Type {
    memberAttrTypes := map[string]attr.Type{
        "id":          types.StringType,
        "name":        types.StringType,
        "description": types.StringType,
        "created_at":  types.StringType,
        "updated_at":  types.StringType,
    }

    baseTypes := map[string]attr.Type{
        "affiliation": types.StringType,
        "permissions": types.ListType{
            ElemType: types.StringType,
        },
        "id":              types.StringType,
        "membership_type": types.StringType,
        "organization_id": types.StringType,
    }

    switch membershipType {
    case user:
        baseTypes[user] = types.ObjectType{
            AttrTypes: memberAttrTypes,
        }
    case serviceAccount:
        baseTypes[serviceAccount] = types.ObjectType{
            AttrTypes: memberAttrTypes,
        }
    }

    return baseTypes
}

func createMembershipAttrValues(membership *iam.IAMOrganizationMembershipPermission, membershipType string) map[string]attr.Value {
    memberAttrTypes := map[string]attr.Type{
        "id":          types.StringType,
        "name":        types.StringType,
        "description": types.StringType,
        "created_at":  types.StringType,
        "updated_at":  types.StringType,
    }
    
    switch membershipType {
    case user:
        return map[string]attr.Value{
            "affiliation": types.StringValue(membership.User.Affiliation),
            "permissions": types.ListValueMust(types.StringType, convertSliceToAttrValues(membership.User.Permissions, func(s string) attr.Value {
                return types.StringValue(s)
            })),
            "id":              types.StringValue(membership.User.Id),
            "membership_type": types.StringValue(membership.User.MembershipType),
            "organization_id": types.StringValue(membership.User.OrganizationId),
            "user": resource_organization_membership_permission.NewUserValueMust(memberAttrTypes, map[string]attr.Value{
                "id":          types.StringValue(membership.User.User.ID),
                "name":        types.StringValue(membership.User.User.Name),
                "description": types.StringValue(membership.User.User.Description),
                "created_at":  types.StringValue(membership.User.User.CreatedAt),
                "updated_at":  types.StringValue(membership.User.User.UpdatedAt),
            }),
        }
    case serviceAccount:
        return map[string]attr.Value{
            "affiliation": types.StringValue(membership.ServiceAccount.Affiliation),
            "permissions": types.ListValueMust(types.StringType, convertSliceToAttrValues(membership.ServiceAccount.Permissions, func(s string) attr.Value {
                return types.StringValue(s)
            })),
            "id":              types.StringValue(membership.ServiceAccount.Id),
            "membership_type": types.StringValue(membership.ServiceAccount.MembershipType),
            "organization_id": types.StringValue(membership.ServiceAccount.OrganizationId),
            "service_account": resource_organization_membership_permission.NewServiceAccountValueMust(memberAttrTypes, map[string]attr.Value{
                "id":          types.StringValue(membership.ServiceAccount.ServiceAccount.ID),
                "name":        types.StringValue(membership.ServiceAccount.ServiceAccount.Name),
                "description": types.StringValue(membership.ServiceAccount.ServiceAccount.Description),
                "created_at":  types.StringValue(membership.ServiceAccount.ServiceAccount.CreatedAt),
                "updated_at":  types.StringValue(membership.ServiceAccount.ServiceAccount.UpdatedAt),
            }),
        }
    default:
        return nil
    }
}

var _ resource.Resource = (*OrganizationMembershipPermissionResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationMembershipPermissionResource)(nil)

func NewOrganizationMembershipPermissionResource() resource.Resource {
	return &OrganizationMembershipPermissionResource{}
}

type OrganizationMembershipPermissionResource struct {
	client *iam.Client
}

func (r *OrganizationMembershipPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_membership_permission"
}

func (r *OrganizationMembershipPermissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_membership_permission.OrganizationMembershipPermissionResourceSchema(ctx)
}

func (r *OrganizationMembershipPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*iam.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *iam.Client, got: %T. Please report this issue to the provider developers.",
			req.ProviderData,
		)

		return
	}

	r.client = client
}

func (r *OrganizationMembershipPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_membership_permission.OrganizationMembershipPermissionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating organization membership permission", map[string]interface{}{
		"organization_id": data.OrganizationId.ValueString(),
		"membership_type": data.Discriminator.ValueString(),
		"member_id":       data.MemberId.ValueString(),
	})

	// Create API call logic
	var permissions []string
	if !data.User.Permissions.IsNull() || !data.User.Permissions.IsUnknown() {
		permissions = data.User.EditablePermissions.ElementsAs(ctx, &permissions, false)
	} else if !data.ServiceAccount.Permissions.IsNull() || !data.ServiceAccount.Permissions.IsUnknown() {
		permissions = data.ServiceAccount.Permissions.Elements()
	}

	elements := make([]string, 0, len(data.))
	response, err := r.client.CreateOrUpdateOrganizationMembershipPermission(ctx, data.MemberId.String(), data.OrganizationId.String(), permissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Organization Membership Permission",
			err.Error(),
		)
		return
	}

	if response.ServiceAccount.MembershipType != "" {
		data.User = resource_organization_membership_permission.NewUserValueMust(createMembershipAttrTypes(user), createMembershipAttrValues(&response, user))
	} else if response.User.MembershipType != "" {
		data.ServiceAccount = resource_organization_membership_permission.NewServiceAccountValueMust(createMembershipAttrTypes(serviceAccount), createMembershipAttrValues(&response, serviceAccount))
	} else {
		resp.Diagnostics.AddError(
			"Invalid Membership Type",
			"Membership type must be either 'user' or 'service_account'.",
		)
		return
	}

	data.MemberId = types.StringValue(response.MemberId)
	data.OrganizationId = types.StringValue(response.OrganizationId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembershipPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_membership_permission.OrganizationMembershipPermissionModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading organization membership permission", map[string]interface{}{
		"organization_id": data.OrganizationId.ValueString(),
		"membership_id":   data.MemberId.ValueString(),
	})

	// Read API call logic
	response, err := r.client.GetOrganizationMembership(data.OrganizationId.String(), data.MemberId.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Organization Membership Permission",
			err.Error(),
		)
		return
	}

	if response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(response.ServiceAccount.ID)
	}
	if response.User.ID != "" {
		data.Id = types.StringValue(response.User.ID)
	}
	data.OrganizationId = types.StringValue(response.Organisation.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembershipPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_membership_permission.OrganizationMembershipPermissionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.MemberId)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating organization membership permission", map[string]interface{}{
		"organization_id": data.OrganizationId.ValueString(),
		"membership_id":   data.MemberId.ValueString(),
	})

	// Update API call logic
	var permissions []string
	if !data.User.Permissions.IsNull() || !data.User.Permissions.IsUnknown() {
		permissions = data.User.EditablePermissions.ElementsAs(ctx, &permissions, false)
	} else if !data.ServiceAccount.Permissions.IsNull() || !data.ServiceAccount.Permissions.IsUnknown() {
		permissions = data.ServiceAccount.Permissions.Elements()
	}

	response, err := r.client.CreateOrUpdateOrganizationMembershipPermission(ctx, data.MemberId.String(), data.OrganizationId.String(), permissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Organization Membership Permission",
			err.Error(),
		)
		return
	}

	if response.ServiceAccount.MembershipType != "" {
		data.User = resource_organization_membership_permission.NewUserValueMust(createMembershipAttrTypes(user), createMembershipAttrValues(&response, user))
	} else if response.User.MembershipType != "" {
		data.ServiceAccount = resource_organization_membership_permission.NewServiceAccountValueMust(createMembershipAttrTypes(serviceAccount), createMembershipAttrValues(&response, serviceAccount))
	} else {
		resp.Diagnostics.AddError(
			"Invalid Membership Type",
			"Membership type must be either 'user' or 'service_account'.",
		)
		return
	}

	data.MemberId = types.StringValue(response.MemberId)
	data.OrganizationId = types.StringValue(response.OrganizationId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembershipPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_membership_permission.OrganizationMembershipPermissionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting organization membership permission", map[string]interface{}{
		"organization_id": data.OrganizationId.ValueString(),
		"membership_id":   data.MemberId.ValueString(),
	})

	// Delete API call logic
	err := r.client.DeleteOrganizationMembershipPermission(ctx, data.OrganizationId.String(), data.MemberId.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Organization Membership Permission",
			err.Error(),
		)
		return
	}
}

func (r *OrganizationMembershipPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || (len(idParts) == 2 && (idParts[0] == "" || idParts[1] == "")) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,member_id. Got: %q", req.ID),
		)
		return
	}

	//Read API call logic
	tflog.Info(ctx, "Reading organization membership permission")
	response, err := r.client.GetOrganizationMembership(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Organization Membership Permission",
			err.Error(),
		)
		return
	}

	var data resource_organization_membership_permission.OrganizationMembershipPermissionModel
	data.OrganizationId = types.StringValue(idParts[0])
	data.MemberId = types.StringValue(idParts[1])
	if response.ServiceAccount.ID != "" {
		data.Id = types.StringValue(response.ServiceAccount.ID)
		data.ServiceAccount = resource_organization_membership_permission.NewServiceAccountValueMust(createMembershipAttrTypes(serviceAccount), createMembershipAttrValues(&response, serviceAccount))
	}
	if response.User.ID != "" {
		data.Id = types.StringValue(response.User.ID)
		data.User = resource_organization_membership_permission.NewUserValueMust(createMembershipAttrTypes(user), createMembershipAttrValues(&response, user))
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
