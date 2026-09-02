package compat

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const OrgIdDeprecatedMessage = "Use org_id instead. This field will be removed in a future version."

// DeprecatedOrganizationIdAttribute returns the deprecated organization_id schema attribute.
func DeprecatedOrganizationIdAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:           true,
		Computed:           true,
		DeprecationMessage: OrgIdDeprecatedMessage,
	}
}

// ResolveOrgId returns the effective org_id, preferring org_id over the deprecated organization_id.
func ResolveOrgId(orgId, organizationId types.String) types.String {
	if !orgId.IsNull() && !orgId.IsUnknown() {
		return orgId
	}
	return organizationId
}

// SyncOrgIds returns both org_id and organization_id set to the same value.
func SyncOrgIds(orgIdValue string) (orgId, organizationId types.String) {
	v := types.StringValue(orgIdValue)
	return v, v
}

// ValidateOrgId checks that at least one of org_id or organization_id is set
// in the resource configuration. Call this from a resource's ValidateConfig method.
func ValidateOrgId(ctx context.Context, config tfsdk.Config, diags *resource.ValidateConfigResponse) {
	var orgId, organizationId types.String
	diags.Diagnostics.Append(config.GetAttribute(ctx, path.Root("org_id"), &orgId)...)
	diags.Diagnostics.Append(config.GetAttribute(ctx, path.Root("organization_id"), &organizationId)...)
	if diags.Diagnostics.HasError() {
		return
	}
	// Unknown means a value was provided but isn't resolved yet (e.g. from a
	// data source); only null means the attribute was not set at all.
	orgIdMissing := orgId.IsNull()
	organizationIdMissing := organizationId.IsNull()
	if orgIdMissing && organizationIdMissing {
		diags.Diagnostics.AddAttributeError(
			path.Root("org_id"),
			"Missing organization identifier",
			"Either org_id or the deprecated organization_id must be set.",
		)
	}
}

// OrgIdStateUpgrader returns a StateUpgrader that migrates organization_id to org_id
// in the raw state JSON. It copies the organization_id value into org_id when
// org_id is not already present.
func OrgIdStateUpgrader() resource.StateUpgrader {
	return resource.StateUpgrader{
		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			if req.RawState == nil || len(req.RawState.JSON) == 0 {
				resp.Diagnostics.AddError(
					"Error reading prior state",
					"No prior state data available for migration.",
				)
				return
			}

			var rawState map[string]json.RawMessage
			if err := json.Unmarshal(req.RawState.JSON, &rawState); err != nil {
				resp.Diagnostics.AddError(
					"Error reading prior state",
					"Could not unmarshal prior state data: "+err.Error(),
				)
				return
			}

			// Copy organization_id to org_id if org_id is absent
			if _, hasOrgId := rawState["org_id"]; !hasOrgId {
				if orgIdRaw, ok := rawState["organization_id"]; ok {
					rawState["org_id"] = orgIdRaw
				}
			}

			upgradedState, err := json.Marshal(rawState)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error writing upgraded state",
					"Could not marshal upgraded state data: "+err.Error(),
				)
				return
			}

			resp.DynamicValue = &tfprotov6.DynamicValue{
				JSON: upgradedState,
			}
		},
	}
}

// TeamStateUpgrader returns a StateUpgrader for the organization team resource.
// In addition to the organization_id to org_id migration performed by
// OrgIdStateUpgrader, it copies the legacy editable_permissions value into
// organization_permissions when organization_permissions is not already
// present, and removes the legacy editable_permissions key.
func TeamStateUpgrader() resource.StateUpgrader {
	return resource.StateUpgrader{
		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			if req.RawState == nil || len(req.RawState.JSON) == 0 {
				resp.Diagnostics.AddError(
					"Error reading prior state",
					"No prior state data available for migration.",
				)
				return
			}

			var rawState map[string]json.RawMessage
			if err := json.Unmarshal(req.RawState.JSON, &rawState); err != nil {
				resp.Diagnostics.AddError(
					"Error reading prior state",
					"Could not unmarshal prior state data: "+err.Error(),
				)
				return
			}

			// Copy organization_id to org_id if org_id is absent
			if _, hasOrgId := rawState["org_id"]; !hasOrgId {
				if orgIdRaw, ok := rawState["organization_id"]; ok {
					rawState["org_id"] = orgIdRaw
				}
			}

			// Copy editable_permissions to organization_permissions if the latter is absent
			if _, hasOrgPermissions := rawState["organization_permissions"]; !hasOrgPermissions {
				if permissionsRaw, ok := rawState["editable_permissions"]; ok {
					rawState["organization_permissions"] = permissionsRaw
				}
			}
			delete(rawState, "editable_permissions")

			upgradedState, err := json.Marshal(rawState)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error writing upgraded state",
					"Could not marshal upgraded state data: "+err.Error(),
				)
				return
			}

			resp.DynamicValue = &tfprotov6.DynamicValue{
				JSON: upgradedState,
			}
		},
	}
}
