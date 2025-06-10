Project Membership Resource

The Project Membership Resource manages an organization project's membership in SysEleven IAM.

## Example Usage

```hcl
resource "sys11iam_project_membership" "test_project_membership" {
  depends_on = [sys11iam_project.test_project, sys11iam_organization_membership.test_membership]
  count = one(sys11iam_organization_membership.test_membership[*].is_active) == true ? 1 : 0
  email = "test@example.com"
  permissions = ["can_become_administrator_in_project", "can_crud_permissions_in_project"]
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_project_membership":

* **`email`** - The email of the user.
* **`permissions`** - The editable permissions of the user.
* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`id`** - The UUID of the project membership. (read-only)

## Importing Organization Project Memberships

To import an organization project membership, your configuration would look like the following:

```hcl
resource "sys11iam_project_membership" "test_project_membership" {
  count = one(sys11iam_organization_membership.test_membership[*].is_active) == true ? 1 : 0
  email = "<email>"
  permissions = []
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}

```
Then you execute:

```bash
terraform import sys11iam_project_membership.test_project_membership[0] <organization_id,project_id,member_id>
```

Where `organization_id` is the ID of the organization, `project_id` is the ID of the project you want to import, and `member_id` is the ID of the member (user/service account) to be imported.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_project_membership.test_project_membership[0]
    id = "<organization_id,project_id,member_id>"
}

resource "sys11iam_project_membership" "test_project_membership" {
  count = one(sys11iam_organization_membership.test_membership[*].is_active) == true ? 1 : 0
  email = "<email>"
  permissions = []
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}

```

Now the resource to be imported can be managed with `terraform plan/apply`.

