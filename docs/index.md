# sys11iam Provider

The sys11iam provider is used to interact with the SysEleven IAM. The provider needs to be configured with the proper credentials before it can be used.

## Example Usage

```hcl
# Define required providers
terraform {
  required_providers {
    sys11iam = {
      source = "syseleven/sys11iam"
    }
  }
}
```

Using the `sys11iam` provider above:

```hcl
# Configure the sys11iam Provider for service account user authentication
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_asdziuch-967s-aduc-123f-00asdasd8asd_9xjakshdkjOJPvk-36Fxxx"
  iam_url = "https://iam.apis.syseleven.de"
}
```

## Configuration Reference

The following arguments are supported for the provider "sys11iam":

* **`iam_url`** - The url to the IAM service for creating organization, project, organization membership and project membership resources.
  If omitted, the `SYS11IAM_IAM_URL` environment variable is used.
* **`serviceaccount_secret`** - The secret of a service account to authenticate with. If omitted, the `SYS11IAM_SERVICEACCOUNT_SECRET` environment variable is used.

## Upgrading project resources to v3

Version 3 renames the old project-scoped resource types so they are grouped under their organization:

```hcl
moved {
  from = sys11iam_project.example
  to   = sys11iam_organization_project.example
}

moved {
  from = sys11iam_project_membership.example
  to   = sys11iam_organization_project_membership.example
}

moved {
  from = sys11iam_project_s3user.example
  to   = sys11iam_organization_project_s3_user.example
}

moved {
  from = sys11iam_project_s3user_key.example
  to   = sys11iam_organization_project_s3_user_key.example
}
```

The provider supports state moves for these direct renames and migrates `organization_id` state to `org_id`.

## Upgrading organization team resources to v3

`sys11iam_organization_team` keeps its type name in v3, so no `moved` block is needed for the team itself. On the first plan/apply with the v3 provider the state is upgraded in place: `organization_id` becomes `org_id` and `editable_permissions` becomes `organization_permissions`.

If a team has exactly one old `sys11iam_project_team` resource, its project permissions can be folded into the team's new `projects` block automatically:

```hcl
moved {
  from = sys11iam_project_team.example
  to   = sys11iam_organization_team.example
}
```

Because the state move adopts the team through the old `project_team` state, the target address must not already have state: run `terraform state rm sys11iam_organization_team.example` before applying the move. The team's name, description, tags and organization permissions are re-read from the API. Teams with multiple old `sys11iam_project_team` resources cannot be moved this way (Terraform does not allow several `moved` blocks to the same target); combine their `editable_permissions` into the team's `projects` block manually and remove the old resources from state.

The old `sys11iam_project_team_membership` state also does not contain enough information to safely infer the v3 nested membership shape in all cases. Migrate those resources by updating configuration to the v3 shape and importing the resulting resources instead of using a `moved` block.
