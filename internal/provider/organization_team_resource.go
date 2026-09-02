package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
	"github.com/syseleven/terraform-provider-sys11iam/internal/compat"
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_team"
)

var _ resource.Resource = (*OrganizationTeamResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationTeamResource)(nil)
var _ resource.ResourceWithMoveState = (*OrganizationTeamResource)(nil)
var _ resource.ResourceWithUpgradeState = (*OrganizationTeamResource)(nil)
var _ resource.ResourceWithValidateConfig = (*OrganizationTeamResource)(nil)

var projectAttrType = map[string]attr.Type{
	"id": types.StringType,
	"project_permissions": basetypes.ListType{
		ElemType: types.StringType,
	},
}
var projectObjectType = basetypes.ObjectType{
	AttrTypes: projectAttrType,
}

func NewOrganizationTeamResource() resource.Resource {
	return &OrganizationTeamResource{}
}

type OrganizationTeamResource struct {
	client *iam.Client
}

func (r *OrganizationTeamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_team"
}

func (r *OrganizationTeamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_organization_team.OrganizationTeamResourceSchemaFull(ctx)
}

func (r *OrganizationTeamResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: compat.TeamStateUpgrader(),
	}
}

func (r *OrganizationTeamResource) MoveState(ctx context.Context) []resource.StateMover {
	return []resource.StateMover{
		compat.RawStateMover("sys11iam_organization_team", func(ctx context.Context, rawState compat.RawState, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			orgId, organizationId, err := compat.RawOrgIDs(rawState)
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read organization identifier: "+err.Error())
				return
			}

			id, err := compat.RawString(rawState, "id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read team id: "+err.Error())
				return
			}

			description, err := compat.RawString(rawState, "description")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read team description: "+err.Error())
				return
			}

			name, err := compat.RawString(rawState, "name")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read team name: "+err.Error())
				return
			}

			tags, err := compat.RawStringList(ctx, rawState, "tags")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read team tags: "+err.Error())
				return
			}

			organizationPermissions, err := compat.RawStringList(ctx, rawState, "editable_permissions")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read team permissions: "+err.Error())
				return
			}

			data := resource_organization_team.OrganizationTeamModelFull{
				Description:             description,
				Id:                      id,
				Name:                    name,
				OrgId:                   orgId,
				OrganizationId:          organizationId,
				OrganizationPermissions: organizationPermissions,
				Projects:                types.ListNull(projectObjectType),
				Tags:                    tags,
			}

			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &data)...)
		}),
		compat.RawStateMover("sys11iam_project_team", func(ctx context.Context, rawState compat.RawState, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			orgId, organizationId, err := compat.RawOrgIDs(rawState)
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read organization identifier: "+err.Error())
				return
			}

			teamId, err := compat.RawString(rawState, "team_id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read team id: "+err.Error())
				return
			}
			if teamId.IsNull() || teamId.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Unsupported project team state",
					"Cannot safely migrate sys11iam_project_team state without a team_id value. Import the team as sys11iam_organization_team instead.",
				)
				return
			}

			projectId, err := compat.RawString(rawState, "project_id")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project id: "+err.Error())
				return
			}

			projectPermissions, err := compat.RawStringList(ctx, rawState, "editable_permissions")
			if err != nil {
				resp.Diagnostics.AddError("Error reading prior state", "Could not read project team permissions: "+err.Error())
				return
			}

			projects := basetypes.NewListValueMust(projectObjectType, []attr.Value{
				basetypes.NewObjectValueMust(projectObjectType.AttrTypes, map[string]attr.Value{
					"id":                  projectId,
					"project_permissions": projectPermissions,
				}),
			})

			data := resource_organization_team.OrganizationTeamModelFull{
				Description:             types.StringNull(),
				Id:                      teamId,
				Name:                    types.StringNull(),
				OrgId:                   orgId,
				OrganizationId:          organizationId,
				OrganizationPermissions: types.ListNull(types.StringType),
				Projects:                projects,
				Tags:                    types.ListNull(types.StringType),
			}

			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &data)...)
		}),
	}
}

func (r *OrganizationTeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationTeamResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	compat.ValidateOrgId(ctx, req.Config, resp)
}

func (r *OrganizationTeamResource) processProjectPermissions(
	ctx context.Context,
	projects []*resource_organization_team.ProjectValue,
	organizationId string,
	projectPermissionsClientRequest func(org_id string, project_id string, team_id string, permissions []string) ([]iam.IAMPermissionEntry, error),
	teamId string,
	maxWorkers int,
) ([]*resource_organization_team.ProjectValue, []error) {
	if len(projects) == 0 {
		return projects, nil
	}

	if maxWorkers <= 0 {
		maxWorkers = MaxWorkers
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	updatedProjects := make([]*resource_organization_team.ProjectValue, len(projects))
	copy(updatedProjects, projects)

	jobs := make(chan int, len(projects))

	// Start worker pool
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for idx := range jobs {
				project := projects[idx]

				projectPermissions := make([]string, 0, len(project.ProjectPermissions.Elements()))
				diags := project.ProjectPermissions.ElementsAs(ctx, &projectPermissions, false)
				if diags.HasError() {
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to extract project permissions for project %s", project.Id.ValueString()))
					mu.Unlock()
					continue
				}

				response, err := projectPermissionsClientRequest(organizationId, project.Id.ValueString(), teamId, projectPermissions)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to process permissions request for project %s: %w", project.Id.ValueString(), err))
					mu.Unlock()
					continue
				}

				permissions := iam.FilterActiveDirectPermissions(response)

				updatedProject := &resource_organization_team.ProjectValue{
					Id: project.Id,
					ProjectPermissions: types.ListValueMust(types.StringType, convertSliceToAttrValues(permissions, func(perm string) attr.Value {
						return types.StringValue(perm)
					})),
				}

				mu.Lock()
				updatedProjects[idx] = updatedProject
				mu.Unlock()
			}
		}()
	}

	// Send work to the workers
	for i := range projects {
		jobs <- i
	}
	close(jobs)

	wg.Wait()

	return updatedProjects, errors
}

func (r *OrganizationTeamResource) buildData(ctx context.Context, data resource_organization_team.OrganizationTeamModelFull, response *iam.IAMOrganizationTeam, orgPermissionsResponse []string, teamProjectPermissionsResponse []iam.IAMTeamProjectWithPermissions) (resource_organization_team.OrganizationTeamModelFull, diag.Diagnostics) {
	var diags diag.Diagnostics

	data.Id = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)

	tagsList, diags := types.ListValueFrom(ctx, types.StringType, response.Tags)
	if diags.HasError() {
		diags.Append(diags...)
		return data, diags
	} else {
		data.Tags = tagsList
	}

	data.OrganizationPermissions = types.ListValueMust(types.StringType, convertSliceToAttrValues(orgPermissionsResponse, func(perm string) attr.Value {
		return types.StringValue(perm)
	}))

	var stateProjects []*resource_organization_team.ProjectValue
	if !data.Projects.IsNull() && !data.Projects.IsUnknown() && len(data.Projects.Elements()) > 0 {
		diags := data.Projects.ElementsAs(ctx, &stateProjects, false)
		if diags.HasError() {
			stateProjects = nil
		}
	}

	if len(stateProjects) > 0 {
		// Merge projects in terraform state & from API, preserving order in favour of the terraform state
		orderedProjects := MergeSlicesWithKeys(stateProjects, teamProjectPermissionsResponse,
			func(p *resource_organization_team.ProjectValue) string {
				return p.Id.ValueString()
			}, func(p iam.IAMTeamProjectWithPermissions) string {
				return p.ID
			}, func(p iam.IAMTeamProjectWithPermissions) *resource_organization_team.ProjectValue {
				return &resource_organization_team.ProjectValue{
					Id: types.StringValue(p.ID),
					ProjectPermissions: types.ListValueMust(types.StringType, convertSliceToAttrValues(p.Permissions, func(perm string) attr.Value {
						return types.StringValue(perm)
					})),
				}
			})

		data.Projects = basetypes.NewListValueMust(projectObjectType, convertSliceToAttrValues(orderedProjects, func(project *resource_organization_team.ProjectValue) attr.Value {
			return basetypes.NewObjectValueMust(projectObjectType.AttrTypes, map[string]attr.Value{
				"id":                  project.Id,
				"project_permissions": project.ProjectPermissions,
			})
		}))
	} else {
		var projects []*resource_organization_team.ProjectValue
		for _, project := range teamProjectPermissionsResponse {
			projectPermissions := types.ListValueMust(types.StringType, convertSliceToAttrValues(project.Permissions, func(perm string) attr.Value {
				return types.StringValue(perm)
			}))

			projects = append(projects, &resource_organization_team.ProjectValue{
				Id:                 types.StringValue(project.ID),
				ProjectPermissions: projectPermissions,
			})
		}

		data.Projects = basetypes.NewListValueMust(projectObjectType, convertSliceToAttrValues(projects, func(project *resource_organization_team.ProjectValue) attr.Value {
			return basetypes.NewObjectValueMust(projectObjectType.AttrTypes, map[string]attr.Value{
				"id":                  project.Id,
				"project_permissions": project.ProjectPermissions,
			})
		}))
	}

	return data, diags
}

func (r *OrganizationTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_organization_team.OrganizationTeamModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve org_id / organization_id
	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationTeam resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", orgId.ValueString()))

	tags := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &tags, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.CreateOrganizationTeam(orgId.ValueString(), data.Name.ValueString(), data.Description.ValueString(), tags)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Create org permissions for the team if set in plan data.
	if len(data.OrganizationPermissions.Elements()) > 0 {
		tflog.Info(ctx, "Creating OrganizationTeam permissions.")

		orgPermissions := make([]string, 0, len(data.OrganizationPermissions.Elements()))

		diags = data.OrganizationPermissions.ElementsAs(ctx, &orgPermissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		orgPermissionsResponse, err := r.client.CreateOrganizationTeamPermission(orgId.ValueString(), response.ID, orgPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}

		data.OrganizationPermissions = types.ListValueMust(types.StringType, convertSliceToAttrValues(iam.FilterActiveDirectPermissions(orgPermissionsResponse), func(perm string) attr.Value {
			return types.StringValue(perm)
		}))
	}

	// Create project permissions for the team to a project if set in plan data.
	if len(data.Projects.Elements()) > 0 {
		tflog.Info(ctx, "Creating OrganizationTeam project permissions.")

		var projects []*resource_organization_team.ProjectValue
		diags := data.Projects.ElementsAs(ctx, &projects, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updatedProjects, errors := r.processProjectPermissions(
			ctx,
			projects,
			orgId.ValueString(),
			r.client.CreateProjectTeamPermissions,
			response.ID,
			MaxWorkers,
		)

		if len(errors) > 0 {
			for _, err := range errors {
				resp.Diagnostics.AddError("Project Permission Creation Failed: ", err.Error())
			}
			return
		}

		data.Projects = types.ListValueMust(projectObjectType, convertSliceToAttrValues(updatedProjects, func(project *resource_organization_team.ProjectValue) attr.Value {
			return types.ObjectValueMust(projectAttrType, map[string]attr.Value{
				"id":                  project.Id,
				"project_permissions": project.ProjectPermissions,
			})
		}))
	}

	data.Id = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Sync org_id and organization_id before saving state
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readProjectPermissions fetches per-project team permissions for each project
// known in state. The v3 API has no "list all projects for a team" endpoint, so
// we re-read permissions for the projects the Terraform state already tracks.
func (r *OrganizationTeamResource) readProjectPermissions(ctx context.Context, orgId string, teamId string, stateProjects basetypes.ListValue) ([]iam.IAMTeamProjectWithPermissions, error) {
	if stateProjects.IsNull() || stateProjects.IsUnknown() || len(stateProjects.Elements()) == 0 {
		return nil, nil
	}

	var projects []*resource_organization_team.ProjectValue
	diags := stateProjects.ElementsAs(ctx, &projects, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract projects from state")
	}

	var result []iam.IAMTeamProjectWithPermissions
	for _, p := range projects {
		projectId := p.Id.ValueString()
		entries, err := r.client.GetProjectTeamPermissions(orgId, projectId, teamId)
		if err != nil {
			return nil, err
		}
		perms := iam.FilterActiveDirectPermissions(entries)
		result = append(result, iam.IAMTeamProjectWithPermissions{
			ID:          projectId,
			Permissions: perms,
		})
	}
	return result, nil
}

func (r *OrganizationTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_team.OrganizationTeamModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve org_id / organization_id
	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationTeam resource.")
	response, err := r.client.GetOrganizationTeam(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	orgPermissionsResponse, err := r.client.GetOrganizationTeamPermissions(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	orgPermissions := iam.FilterActiveDirectPermissions(orgPermissionsResponse)

	teamProjectPermissionsResponse, err := r.readProjectPermissions(ctx, orgId.ValueString(), data.Id.ValueString(), data.Projects)
	if err != nil {
		resp.Diagnostics.AddError("", fmt.Sprintf("could not get TeamProjects: %s", err.Error()))
		return
	}

	// Data value setting
	data, diags := r.buildData(ctx, data, &response, orgPermissions, teamProjectPermissionsResponse)
	resp.Diagnostics.Append(diags...)

	// Sync org_id and organization_id before saving state
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_team.OrganizationTeamModelFull

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve org_id / organization_id
	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationTeam resource.")
	elements := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateOrganizationTeam(orgId.ValueString(), data.Id.ValueString(), data.Name.ValueString(), data.Description.ValueString(), elements)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	if len(data.OrganizationPermissions.Elements()) > 0 {
		tflog.Info(ctx, "Updating OrganizationTeam permissions.")

		orgPermissions := make([]string, 0, len(data.OrganizationPermissions.Elements()))

		diags = data.OrganizationPermissions.ElementsAs(ctx, &orgPermissions, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		orgPermissionsResponse, err := r.client.UpdateOrganizationTeamPermission(orgId.ValueString(), response.ID, orgPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}

		data.OrganizationPermissions = types.ListValueMust(types.StringType, convertSliceToAttrValues(iam.FilterActiveDirectPermissions(orgPermissionsResponse), func(perm string) attr.Value {
			return types.StringValue(perm)
		}))
	}

	if len(data.Projects.Elements()) > 0 {
		tflog.Info(ctx, "Updating OrganizationTeam project permissions.")

		var projects []*resource_organization_team.ProjectValue
		diags := data.Projects.ElementsAs(ctx, &projects, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updatedProjects, errors := r.processProjectPermissions(
			ctx,
			projects,
			orgId.ValueString(),
			r.client.UpdateProjectTeamPermissions,
			response.ID,
			MaxWorkers,
		)

		if len(errors) > 0 {
			for _, err := range errors {
				resp.Diagnostics.AddError("Project Permission Update Failed: ", err.Error())
			}
			return
		}

		// Merge projects in terraform state & from API, preserving order in favour of the terraform state
		orderedProjects := MergeSlicesWithKeys(projects, updatedProjects, func(p *resource_organization_team.ProjectValue) string {
			return p.Id.ValueString()
		}, func(p *resource_organization_team.ProjectValue) string {
			return p.Id.ValueString()
		}, func(p *resource_organization_team.ProjectValue) *resource_organization_team.ProjectValue {
			return p
		})

		data.Projects = types.ListValueMust(projectObjectType, convertSliceToAttrValues(orderedProjects, func(project *resource_organization_team.ProjectValue) attr.Value {
			return types.ObjectValueMust(projectObjectType.AttrTypes, map[string]attr.Value{
				"id":                  project.Id,
				"project_permissions": project.ProjectPermissions,
			})
		}))
	}

	// Data value setting
	data.Id = types.StringValue(response.ID)
	data.Name = types.StringValue(response.Name)
	data.Description = types.StringValue(response.Description)
	data.Tags, _ = types.ListValueFrom(ctx, types.StringType, response.Tags)

	// Sync org_id and organization_id before saving state
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(orgId.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_team.OrganizationTeamModelFull

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve org_id / organization_id
	orgId := compat.ResolveOrgId(data.OrgId, data.OrganizationId)

	// Delete API call logic
	tflog.Info(ctx, "Deleting OrganizationTeam resource.")
	err := r.client.DeleteOrganizationTeam(orgId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
}

func (r *OrganizationTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || (len(idParts) == 2 && (idParts[0] == "" || idParts[1] == "")) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: org_id,team_id. Got: %q", req.ID),
		)
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Importing OrganizationTeam resource.")
	response, err := r.client.GetOrganizationTeam(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	orgPermissionsResponse, err := r.client.GetOrganizationTeamPermissions(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	orgPermissions := iam.FilterActiveDirectPermissions(orgPermissionsResponse)

	// v3 API has no endpoint to list all projects for a team, so on import we
	// start with an empty project list. A subsequent plan/apply that includes
	// `projects` blocks will reconcile.
	var teamProjectPermissionsResponse []iam.IAMTeamProjectWithPermissions

	var data resource_organization_team.OrganizationTeamModelFull

	// Data value setting
	data.OrgId, data.OrganizationId = compat.SyncOrgIds(idParts[0])
	data, diags := r.buildData(ctx, data, &response, orgPermissions, teamProjectPermissionsResponse)
	resp.Diagnostics.Append(diags...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
