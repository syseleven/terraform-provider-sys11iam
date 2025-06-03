# Organization Membership Resource

The Organization Membership Resource defines a way to manage the membership of a user within an organization in SysEleven IAM.

## Example Usage 

```hcl
resource "sys11iam_organization_membership" "test_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  email = "test@example.com"
  affiliation = "member"
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = data.sys11iam_organization.testorg.id
}

```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization_membership":

* **`email`** - The email of the user.
* **`affiliation`** - The affiliation of the user to the organization. This is not to be understood as a role. The member affiliation can be ("member" | "admin" | "owner")
* **`editable_permissions`** - The editable permissions of the user in an organization. 

    Supported permissions: 
    * `can_become_project_administrator_in_org` 
    * `can_create_projects_in_org`
    * `can_invite_members_in_org`
    * `can_crud_permissions_in_org`
    * `can_read_members_in_org`
    * `can_delete_members_in_org`
    * `can_manage_contact_persons_in_org`
    * `can_read_contact_persons_in_org`
    * `can_create_teams_in_org`
    * `can_create_service_accounts_in_org`
* **`organization_id`** - The UUID of the organization.
* **`is_active`** - Whether the organization membership is active or not. Organization membership activation is a manual step executed by the invited user. An invitation is issued by creating this resource. (default: false)
* **`id`** - The UUID of the organization membership. (read-only)


