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

## Importing Organization Memberships

To import an organization membership, your configuration would look like the following:

```hcl
resource "sys11iam_organization_membership" "test_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  email = "test@example.com"
  affiliation = "member"
  editable_permissions = []
  organization_id = data.sys11iam_organization.testorg.id
}

```
Then you execute:

```bash
terraform import sys11iam_organization_membership.test_membership[0] <organization_id,member_id>
```

Where `organization_id` is the ID of the organization and `member_id` is the ID of the organization member you want to import.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_organization_membership.test_membership[0]
    id = "<organization_id,member_id>"
}

resource "sys11iam_organization_membership" "test_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  email = "test@example.com"
  affiliation = "member"
  editable_permissions = []
  organization_id = data.sys11iam_organization.testorg.id
}
```
Now the resource to be imported can be managed with `terraform plan/apply`.

