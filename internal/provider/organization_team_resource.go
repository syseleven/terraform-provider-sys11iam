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
	"github.com/syseleven/terraform-provider-sys11iam/internal/resource_organization_team"
)

var _ resource.Resource = (*OrganizationTeamResource)(nil)
var _ resource.ResourceWithConfigure = (*OrganizationTeamResource)(nil)

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
	resp.Schema = resource_organization_team.OrganizationTeamResourceSchema(ctx)
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

func (r *OrganizationTeamResource) processProjectPermissions(
	ctx context.Context,
	projects []*resource_organization_team.ProjectValue,
	organizationId string,
	projectPermissionsClientRequest func(org_id string, project_id string, team_id string, permissions []string) (iam.IAMProjectTeamPermissions, error),
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

				var permissions []string
				if resp, ok := any(response).(iam.IAMProjectTeamPermissions); ok {
					permissions = resp.UpdatedPermissions
				} else if resp, ok := any(response).([]string); ok {
					permissions = resp
				} else {
					mu.Lock()
					errors = append(errors, fmt.Errorf("unexpected response type for project %s: %T", project.Id.ValueString(), response))
					mu.Unlock()
					continue
				}

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

func (r *OrganizationTeamResource) buildData(ctx context.Context, data resource_organization_team.OrganizationTeamModel, response *iam.IAMOrganizationTeam, orgPermissionsResponse []string, teamProjectPermissionsResponse []iam.IAMTeamProjectWithPermissions) (resource_organization_team.OrganizationTeamModel, diag.Diagnostics) {
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
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	tflog.Info(ctx, "Creating OrganizationTeam resource.")
	tflog.Info(ctx, fmt.Sprintf("Checking if organization with id %s is active.", data.OrganizationId.ValueString()))

	tags := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &tags, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.CreateOrganizationTeam(data.OrganizationId.ValueString(), data.Name.ValueString(), data.Description.ValueString(), tags)
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

		orgPermissionsResponse, err := r.client.CreateOrganizationTeamPermission(data.OrganizationId.ValueString(), response.ID, orgPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}

		data.OrganizationPermissions = types.ListValueMust(types.StringType, convertSliceToAttrValues(orgPermissionsResponse.UpdatedPermissions, func(perm string) attr.Value {
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
			data.OrganizationId.ValueString(),
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Info(ctx, "Reading OrganizationTeam resource.")
	response, err := r.client.GetOrganizationTeam(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	orgPermissionsResponse, err := r.client.GetOrganizationTeamPermissions(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	teamProjectPermissionsResponse, err := r.client.GetTeamProjects(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	// Data value setting
	data, diags := r.buildData(ctx, data, &response, orgPermissionsResponse, teamProjectPermissionsResponse)
	resp.Diagnostics.Append(diags...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &data.Id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	tflog.Info(ctx, "Updating OrganizationTeam resource.")
	elements := make([]string, 0, len(data.Tags.Elements()))
	diags := data.Tags.ElementsAs(ctx, &elements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateOrganizationTeam(data.OrganizationId.ValueString(), data.Id.ValueString(), data.Name.ValueString(), data.Description.ValueString(), elements)
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

		orgPermissionsResponse, err := r.client.UpdateOrganizationTeamPermission(data.OrganizationId.ValueString(), response.ID, orgPermissions)
		if err != nil {
			resp.Diagnostics.AddError("", err.Error())
			return
		}

		data.OrganizationPermissions = types.ListValueMust(types.StringType, convertSliceToAttrValues(orgPermissionsResponse.UpdatedPermissions, func(perm string) attr.Value {
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
			data.OrganizationId.ValueString(),
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_organization_team.OrganizationTeamModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Info(ctx, "Deleting OrganizationTeam resource.")
	err := r.client.DeleteOrganizationTeam(data.OrganizationId.ValueString(), data.Id.ValueString())
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

	teamProjectPermissionsResponse, err := r.client.GetTeamProjects(idParts[0], idParts[1])
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	var data resource_organization_team.OrganizationTeamModel

	// Data value setting
	data.OrganizationId = types.StringValue(idParts[0])
	data, diags := r.buildData(ctx, data, &response, orgPermissionsResponse, teamProjectPermissionsResponse)
	resp.Diagnostics.Append(diags...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
