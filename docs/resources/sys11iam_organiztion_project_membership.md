# Organization Project Membership Resource

The Organization Project Membership Resource manages an organization project's membership in SysEleven IAM, allowing users and service accounts to be added to projects with specific permissions.

## Example Usage

### User Membership

```hcl
resource "sys11iam_organization_project_membership" "test_user_membership" {
  organization_id = data.sys11iam_organization.test_org.id
  project_id      = sys11iam_organization_project.test_project.id
  id              = sys11iam_organization_membership.test_user_membership[0].id

  membership = {
    user_membership = {
      permissions = ["can_crud_permissions_in_project"]
      membership_type = "user"
      user = {
        email = "test@example.com"
      }
    }
  }
}
```

### Service Account Membership

```hcl
resource "sys11iam_organization_project_membership" "test_service_account_membership" {
  organization_id = data.sys11iam_organization.test_org.id
  project_id      = sys11iam_organization_project.test_project.id
  id              = sys11iam_organization_membership.test_service_account_membership[0].id

  membership = {
    service_account_membership = {
      permissions = ["can_crud_permissions_in_project"]
      membership_type = "service_account"
      service_account = {
        id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization_project_membership":

* **`organization_id`** - The UUID of the organization. (Required)
* **`project_id`** - The UUID of the project.
* **`id`** - The unique identifier of the member (user or service account) in the project.
* **`membership`** - The membership configuration block.

### Membership Block

The `membership` block must contain either a `user_membership` or `service_account_membership` block, but not both.

#### User Membership Block

* **`user_membership.permissions`** - The editable permissions of the user in the project. (Required)
* **`user_membership.membership_type`** - The type of the membership. (Default: "user")
* **`user_membership.user.email`** - The email address of the user.
* **`user_membership.user.id`** - The UUID of the user.

#### Service Account Membership Block

* **`service_account_membership.permissions`** - The editable permissions of the service account in the project. (Required)
* **`service_account_membership.membership_type`** - The type of the membership. (Default: "service_account")
* **`service_account_membership.service_account.id`** - The UUID of the service account.
* **`service_account_membership.service_account.name`** - The unique name of the service account.

## Importing Organization Project Memberships

To import an organization project membership, your configuration would look like the following:

```hcl
resource "sys11iam_organization_project_membership" "test_membership" {
  organization_id = data.sys11iam_organization.test_org.id
  project_id      = "project-uuid"
  id              = "member-uuid"
  
  membership = {
    user_membership = {
      permissions = []
      membership_type = "user"
      user = {
        email = "user@example.com"
      }
    }
  }
}
```

Then you execute:

```bash
terraform import sys11iam_organization_project_membership.test_membership <organization_id,project_id,member_id>
```

Where `organization_id` is the ID of the organization, `project_id` is the ID of the project you want to import, and `member_id` is the ID of the member (user/service account) to be imported.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_organization_project_membership.test_membership
    id = "<organization_id,project_id,member_id>"
}

resource "sys11iam_organization_project_membership" "test_membership" {
  organization_id = data.sys11iam_organization.test_org.id
  project_id      = "project-uuid"
  id              = "member-uuid"
  
  membership = {
    user_membership = {
      permissions = []
      membership_type = "user"
      user = {
        email = "user@example.com"
      }
    }
  }
}
```

Now the resource to be imported can be managed with `terraform plan/apply`.
