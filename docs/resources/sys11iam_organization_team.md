# Organization Team Resource

The Organization Team Resource enables the management of a team in an Organization for SysEleven's IAM.

## Example Usage

```hcl
resource "sys11iam_organization_team" "test" {
  name        = "test-team"
  org_id = data.sys11iam_organization.test_org.id
  description = "Test team for acceptance testing"
  organization_permissions = ["can_become_project_administrator_in_org"]
  tags = ["test", "acceptance-testing"]

  projects = [
    {
      id = sys11iam_organization_project.test_project_1.id
      project_permissions = ["can_become_administrator_in_project"]
    },
    {
      id = sys11iam_organization_project.test_project_2.id
      project_permissions = ["can_become_administrator_in_project"]
    }
  ]
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization_team":

* **`name`** - The name of the organization team. Must be between 3-62 characters and contain only lowercase letters, numbers, and hyphens.
* **`description`** - The description of the organization team. Maximum 1000 characters.
* **`tags`** - The tags of the organization team.
* **`organization_permissions`** - The permissions of the team within the organization.
    Supported Permissions:
    * `can_become_project_administrator_in_org`
    * `can_create_projects_in_org`
    * `can_invite_members_in_org`
    * `can_crud_permissions_in_org`
    * `can_read_members_in_org`
    * `can_delete_members_in_org`
    * `can_manage_contact_persons_in_org`
    * `can_read_contact_persons_in_org`
    * `can_create_service_accounts_in_org`

* **`org_id`** - The UUID of the organization.
* **`projects`** - A list of project configurations for the team.
    * **`id`** - The UUID of the project.
    * **`project_permissions`** - The permissions of the team within the project.Supported Project Permissions:
        * `can_become_administrator_in_project`
        * `can_create_projects_in_project`
        * `can_invite_members_in_project`
        * `can_crud_permissions_in_project`
        * `can_read_members_in_project`
        * `can_delete_members_in_project`
        * `can_manage_contact_persons_in_project`
        * `can_read_contact_persons_in_project`
        * `can_create_service_accounts_in_project`

* **`id`** - The UUID of the organization team. (read-only)

## Importing Organization Teams

To import an organization team, your configuration would look like the following:

```hcl
resource "sys11iam_organization_team" "test" {
  name = "<team name>"
  description = "<description>"
  tags = []
  organization_permissions = []
  org_id = data.sys11iam_organization.test_org.id
  projects = []
}
```

Then you execute:

```bash
terraform import sys11iam_organization_team.test <org_id,team_id>
```

Where `org_id` is the ID of the organization and `team_id` is the ID of the team you want to import.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
  to = sys11iam_organization_team.test
  id = "<org_id,team_id>"
}

resource "sys11iam_organization_team" "test" {
  name = "<team name>"
  description = "<description>"
  tags = []
  organization_permissions = []
  org_id = data.sys11iam_organization.test_org.id
  projects = []
}
```

Now the resource to be imported can be managed with `terraform plan/apply`.
