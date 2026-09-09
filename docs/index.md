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
  If omitted, the `SYS11IAM_IAM_URL` environment variable is used; otherwise, it defaults to `https://iam.apis.syseleven.de`.
* **`serviceaccount_secret`** - The secret of a service account to authenticate with. If omitted, the `SYS11IAM_SERVICEACCOUNT_SECRET` environment variable is used.

## Upgrading to v3

Version 3 renames the project-scoped resource types so they are grouped under their organization. Add `moved`
blocks to your configuration before switching to the new type names and the provider will migrate the state
(including the `organization_id` state attribute to `org_id`):

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

In addition, `organization_id` is renamed to `org_id` everywhere (the old name is still accepted with a
deprecation warning), and two resource types were removed:

* `sys11iam_project_team` — replaced by the `projects` block of `sys11iam_organization_team`
* `sys11iam_project_team_membership` — replaced by `sys11iam_organization_team_membership`

Both removed types require a manual migration. See the [migration guide](MIGRATION_GUIDE.md) for the complete,
step-by-step procedure, including the attribute renames, the new membership schema, and resource name mapping.
