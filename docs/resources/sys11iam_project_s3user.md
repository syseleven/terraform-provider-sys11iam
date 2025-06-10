Project S3 User Resource

The Project S3 User Resource manages an S3 User in a Project for SysEleven IAM.

## Example Usage

```hcl
resource "sys11iam_project_s3user" "test_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "tests3user"
  description = "test s3user"
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_project_s3user":
* **`name`** - The name of the S3 User.
* **`description`** - The description of the S3 User.
* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.

## Importing Organization Project Memberships

To import an organization project membership, your configuration would look like the following:

```hcl
resource "sys11iam_project_s3user" "test_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<name>"
  description = "<description>"
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}

```
Then you execute:

```bash
terraform import sys11iam_project_s3user.test_project_s3user[0] <organization_id,project_id,s3_user_id>
```

Where `organization_id` is the ID of the organization, `project_id` is the ID of the project you want to import, and `s3_user_id` is the ID of the S3 user to be imported.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_project_s3user.test_project_s3user[0]
    id = "<organization_id,project_id,s3_user_id>"
}

resource "sys11iam_project_s3user" "test_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<name>"
  description = "<description>"
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}

```

Now the resource to be imported can be managed with `terraform plan/apply`.

