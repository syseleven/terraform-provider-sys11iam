Organization Team Membership Resource

The Organization Team Membership Resource manages the membership of a team in Organization.

## Example Usage

```hcl
resource "sys11iam_organization_team_membership" "testorganization_team_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
  organization_id = data.sys11iam_organization.testorg.id
  team_id = sys11iam_organization_team.testorganization_team[0].id
}
```

## Argument Reference
The following arguments are supported for the resource "sys11iam_organization_team_membership":

* **`id`** - The UUID of the regular user or service account to add as a member.
* **`organization_id`** - The UUID of the organization.
* **`team_id`** - The UUID of the organization team.

## Importing Organization Team Memberships

To import an organization team membership, your configuration would look like the following:

```hcl
resource "sys11iam_organization_team_membership" "testorganization_team_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  id = "<member id>"
  organization_id = data.sys11iam_organization.testorg.id
  team_id = "<team id>"
}

```
Then you execute:

```bash
terraform import sys11iam_organization_team_membership.testorganization_team_membership[0] <organization_id,team_id,member_id>
```

Where `organization_id` is the ID of the organization, `team_id` is the ID of the team you want to import, and `member_id` is the ID of the team member (user/service account) to be imported.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_organization_team_membership.testorganization_team_membership[0] 
    id = "<organization_id,team_id,member_id>"
}

resource "sys11iam_organization_team_membership" "testorganization_team_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  id = "<member id>"
  organization_id = data.sys11iam_organization.testorg.id
  team_id = "<team id>"
}

```
Now the resource to be imported can be managed with `terraform plan/apply`.

