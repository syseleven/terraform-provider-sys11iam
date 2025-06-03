Organization Team Resource

The Organization Team Resource enables the management of a team in an Organization for SysEleven's IAM.

## Example Usage

```hcl
resource "sys11iam_organization_team" "testorganization_team" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "testteam"
  description = "test team"
  tags = ["testing2"]
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = data.sys11iam_organization.testorg.id
}
```

## Argument Reference
The following arguments are supported for the resource "sys11iam_organization_team":

* **`name`** - The name of the organization team.
* **`description`** - The description of the organization team.
* **`tags`** - The tags of the organization team.
* **`editable_permissions`** - The editable permissions of the organization team.
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

* **`organization_id`** - The UUID of the organization.
* **`id`** - The UUID of the organization team. (read-only)

