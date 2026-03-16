# Organization Team Membership Resource

The Organization Team Membership Resource enables the management of user and service account memberships within organization teams in SysEleven's IAM.

## Example Usage

### User Team Membership

```hcl
resource "sys11iam_organization_team_membership" "user_membership_test" {
  team_id = sys11iam_organization_team.test_team.id
  id = "user-uuid-here"
  org_id = data.sys11iam_organization.test_org.id
  membership_type = "user"

  membership = {
    user_team_membership = {
      user = {
        email = "user@example.com"
      }
      team_permissions = ["can_manage_team_in_team", "can_become_administrator_in_team"]
    }
  }
}
```

### Service Account Team Membership

```hcl
resource "sys11iam_organization_team_membership" "service_account_membership_test" {
  team_id = sys11iam_organization_team.test_team.id
  id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
  org_id = data.sys11iam_organization.test_org.id
  membership_type = "service_account"

  membership = {
    service_account_team_membership = {
      team_permissions = ["can_manage_team_in_team"]
    }
  }
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization_team_membership":

* **`team_id`** - The UUID of the team. (Required)
* **`id`** - The UUID of the user or service account. (Required)
* **`org_id`** - The UUID of the organization. (Required)
* **`membership_type`** - The type of the membership.
* **`membership`** - The membership configuration block. (Required)

### Membership Block

The `membership` block must contain either a `user_team_membership` or `service_account_team_membership` block, but not both.

#### User Team Membership Block

* **`user_team_membership.membership_type`** - The type of the membership. (Default: "user")
* **`user_team_membership.team_permissions`** - The team permissions the user has in the team.
* **`user_team_membership.user.email`** - The email address of the user.
* **`user_team_membership.user.id`** - The UUID of the user.

#### Service Account Team Membership Block

* **`service_account_team_membership.team_permissions`** - The team permissions the service account has in the team.

## Importing Organization Team Memberships

To import an organization team membership, your configuration would look like the following:

```hcl
resource "sys11iam_organization_team_membership" "test_membership" {
  team_id = "team-uuid-here"
  id = "member-uuid-here"
  org_id = "organization-uuid-here"
  membership_type = "user"

  membership = {
    user_team_membership = {
      membership_type = "user"
      user = {
        email = "user@example.com"
      }
      team_permissions = []
    }
  }
}
```

Then you execute:

```bash
terraform import sys11iam_organization_team_membership.test_membership <org_id,team_id,member_id>
```

Where `org_id` is the ID of the organization, `team_id` is the ID of the team, and `member_id` is the ID of the team member you want to import.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_organization_team_membership.test_membership
    id = "<org_id,team_id,member_id>"
}

resource "sys11iam_organization_team_membership" "test_membership" {
  team_id = "team-uuid-here"
  id = "member-uuid-here"
  org_id = "organization-uuid-here"
  membership_type = "user"

  membership = {
    user_team_membership = {
      membership_type = "user"
      user = {
        email = "user@example.com"
      }
      team_permissions = []
    }
  }
}
```

Now the resource to be imported can be managed with `terraform plan/apply`.
