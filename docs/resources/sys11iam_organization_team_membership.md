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

