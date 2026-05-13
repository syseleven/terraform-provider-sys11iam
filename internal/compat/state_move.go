package compat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	sys11IAMProviderAddress      = "sys11iam"
	sys11IAMProviderAddressShort = "syseleven/sys11iam"
)

type RawState map[string]json.RawMessage

type RawStateMoveFunc func(context.Context, RawState, resource.MoveStateRequest, *resource.MoveStateResponse)

// RawStateMover returns a cautious StateMover for a single legacy resource type.
func RawStateMover(sourceTypeName string, transform RawStateMoveFunc) resource.StateMover {
	return resource.StateMover{
		StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			if !MatchesMoveSource(req, sourceTypeName) {
				return
			}

			rawState, ok := DecodeMoveRawState(req, &resp.Diagnostics)
			if !ok {
				return
			}

			transform(ctx, rawState, req, resp)
		},
	}
}

func MatchesMoveSource(req resource.MoveStateRequest, sourceTypeName string) bool {
	if req.SourceTypeName != sourceTypeName || req.SourceSchemaVersion != 0 {
		return false
	}

	if req.SourceProviderAddress == sys11IAMProviderAddress {
		return true
	}

	return req.SourceProviderAddress == sys11IAMProviderAddressShort ||
		strings.HasSuffix(req.SourceProviderAddress, "/"+sys11IAMProviderAddressShort)
}

func DecodeMoveRawState(req resource.MoveStateRequest, diagnostics *diag.Diagnostics) (RawState, bool) {
	if req.SourceRawState == nil || len(req.SourceRawState.JSON) == 0 {
		diagnostics.AddError(
			"Error reading prior state",
			"No prior state data available for migration.",
		)
		return nil, false
	}

	var rawState RawState
	if err := json.Unmarshal(req.SourceRawState.JSON, &rawState); err != nil {
		diagnostics.AddError(
			"Error reading prior state",
			"Could not unmarshal prior state data: "+err.Error(),
		)
		return nil, false
	}

	return rawState, true
}

func RawString(rawState RawState, name string) (types.String, error) {
	raw, ok := rawState[name]
	if !ok || string(raw) == "null" {
		return types.StringNull(), nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return types.StringNull(), err
	}

	return types.StringValue(value), nil
}

func RawStringFallback(rawState RawState, names ...string) (types.String, error) {
	for _, name := range names {
		value, err := RawString(rawState, name)
		if err != nil {
			return types.StringNull(), err
		}
		if !value.IsNull() {
			return value, nil
		}
	}

	return types.StringNull(), nil
}

func RawStringList(ctx context.Context, rawState RawState, name string) (types.List, error) {
	raw, ok := rawState[name]
	if !ok || string(raw) == "null" {
		return types.ListNull(types.StringType), nil
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return types.ListNull(types.StringType), err
	}

	listValue, diags := types.ListValueFrom(ctx, types.StringType, values)
	if diags.HasError() {
		errs := diags.Errors()
		return types.ListNull(types.StringType), fmt.Errorf("%s: %s", errs[0].Summary(), errs[0].Detail())
	}

	return listValue, nil
}

func RawOrgIDs(rawState RawState) (types.String, types.String, error) {
	orgID, err := RawStringFallback(rawState, "org_id", "organization_id")
	if err != nil {
		return types.StringNull(), types.StringNull(), err
	}

	if orgID.IsNull() {
		return types.StringNull(), types.StringNull(), nil
	}

	return orgID, orgID, nil
}
