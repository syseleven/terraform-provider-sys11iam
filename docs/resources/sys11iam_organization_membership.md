# Organization Membership Resource

The Organization Membership Resource defines a way to manage the membership of a user or service account within an organization in SysEleven IAM.

## Example Usage 

### User Membership

```hcl
resource "sys11iam_organization_membership" "test_user_membership" {
  count = data.sys11iam_organization.test_org.is_active ? 1 : 0
  id = "user-uuid-here"
  organization_id = data.sys11iam_organization.test_org.id

  membership = {
    user_membership = {
      permissions = ["can_crud_permissions_in_org"]
      affiliation = "member"
      membership_type = "user"
      email = "test@example.com"
    }
  }
}
```

### Service Account Membership

```hcl
resource "sys11iam_organization_membership" "test_service_account_membership" {
  count = data.sys11iam_organization.test_org.is_active ? 1 : 0
  id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
  organization_id = data.sys11iam_organization.test_org.id
  
  membership = {
    service_account_membership = {
      permissions = ["can_crud_permissions_in_org"]
      affiliation = "member"
      membership_type = "service_account"
    }
  }
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization_membership":

* **`id`** - The UUID of the user or service account. (Required)
* **`organization_id`** - The UUID of the organization.
* **`membership`** - The membership configuration block. (Required)

### Membership Block

The `membership` block must contain either a `user_membership` or `service_account_membership` block, but not both.

#### User Membership Block

* **`user_membership.affiliation`** - The affiliation of the user to the organization. This is not to be understood as a role. The member affiliation can be ("member" | "admin" | "owner")
* **`user_membership.permissions`** - The editable permissions of the user in an organization. 

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
* **`user_membership.membership_type`** - The type of the membership. (Default: "user")
* **`user_membership.user.email`** - The email address of the user.
* **`user_membership.user.id`** - The UUID of the user.

#### Service Account Membership Block

* **`service_account_membership.affiliation`** - The affiliation of the service account to the organization. This is not to be understood as a role. The member affiliation can be ("member" | "admin")
* **`service_account_membership.permissions`** - The editable permissions of the service account in an organization. (Required)
* **`service_account_membership.membership_type`** - The type of the membership. (Default: "service_account")

## Importing Organization Memberships

To import an organization membership, your configuration would look like the following:

```hcl
resource "sys11iam_organization_membership" "test_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  id = "user-uuid-here"
  organization_id = data.sys11iam_organization.testorg.id
  
  membership = {
    user_membership = {
      permissions = []
      affiliation = "member"
      membership_type = "user"
      email = "test@example.com"
    }
  }
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
  
  membership = {
    user_membership = {
      permissions = []
      affiliation = "member"
      membership_type = "user"
      email = "test@example.com"
    }
  }
}
```

Now the resource to be imported can be managed with `terraform plan/apply`.
