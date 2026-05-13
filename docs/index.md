# sys11iam Provider

The sys11iam provider is used to interact with the SysEleven IAM. The provider needs to be configured with the proper credentials (a service account) before it can be used.

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
* **`serviceaccount_secret`** - The secret of an service account to authenticate with. If omitted, the `SYS11IAM_SERVICEACCOUNT_SECRET` environment variable is used.

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

The old `sys11iam_project_team` resource is now represented by project permission entries in `sys11iam_organization_team.projects`, which can combine multiple old project-team resources into a single team resource. The old `sys11iam_project_team_membership` state also does not contain enough information to safely infer the v3 nested membership shape in all cases. Migrate those resources by updating configuration to the v3 shape and importing the resulting resources instead of using a `moved` block.
