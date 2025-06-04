Project Team Resource

The Project Team Resource manages a team with Project permissions in SysEleven IAM.

## Example Usage

```hcl
resource "sys11iam_project_team" "test_project_team" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
  team_id = sys11iam_organization_team.testorganization_team[0].id
  editable_permissions = ["can_become_administrator_in_project"]
}
```

## Argument Reference 

The following arguments are supported for the resource "sys11iam_project_team":

* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`team_id`** - The UUID of the organization team.
* **`editable_permissions`** - The editable permissions of the project team.

    Supported Permissions:
    * `can_become_administrator_in_project`
    * `can_crud_permissions_in_project`
    * `can_read_project_in_project`
    * `can_delete_project_in_project`
    * `can_create_api_keys_in_project`
    * `can_read_api_keys_in_project`
    * `can_become_viewer_in_openstack`
    * `can_become_editor_in_openstack`
    * `can_create_clusters_in_metakube`
    * `can_read_clusters_in_metakube`
    * `can_update_clusters_in_metakube`
    * `can_delete_clusters_in_metakube`
    * `can_create_ssh_keys_in_metakube`
    * `can_read_ssh_keys_in_metakube`
    * `can_update_ssh_keys_in_metakube`
    * `can_delete_ssh_keys_in_metakube`
    * `can_create_databases_in_dbaas`
    * `can_read_databases_in_dbaas`
    * `can_update_databases_in_dbaas`
    * `can_delete_databases_in_dbaas`
    * `can_read_backups_from_databases_in_dbaas`
    * `can_become_reader_in_observability`
    * `can_become_editor_in_observability`
    * `can_become_admin_in_observability`

## Importing Organization Project Teams

To import an organization project team, your configuration would look like the following:

```hcl
resource "sys11iam_project_team" "test_project_team" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  organization_id = data.sys11iam_organization.testorg.id # or ""
  project_id = sys11iam_project.test_project[0].id # or ""
  team_id = sys11iam_organization_team.testorganization_team[0].id # or ""
  editable_permissions = []
}

```
Then you execute:

```bash
terraform import sys11iam_project_team.test_project_team <organization_id,project_id,team_id>
```

Where `organization_id` is the ID of the organization, `project_id` is the ID of the project you want to import, and `team_id` is the ID of the team to be imported.
