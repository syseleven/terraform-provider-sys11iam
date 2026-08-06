# Organization Project S3 User Resource

The Project S3 User Resource manages an S3 User in an Organization's Project for SysEleven IAM.

## Example Usage

```hcl
resource "sys11iam_organization_project_s3_user" "test_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "tests3user"
  description = "test s3user"
  org_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_organization_project.test_project.id
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization_project_s3_user":
* **`name`** - The name of the S3 User.
* **`description`** - The description of the S3 User.
* **`org_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`keys`** - List of key pairs for the S3 user (read-only)
* **`s3_user_id`** - The S3 user ID (read-only)
* **`id`** - The UUID of the S3 user (read-only)
* **`created_at`** - The time the S3 user was created (read-only)

## Importing Organization Project S3 Users

To import an organization project membership, your configuration would look like the following:

```hcl
resource "sys11iam_organization_project_s3_user" "test_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<name>"
  description = "<description>"
  org_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_organization_project.test_project.id
}

```
Then you execute:

```bash
terraform import sys11iam_organization_project_s3_user.test_project_s3user[0] <org_id,project_id,s3_user_id>
```

Where `org_id` is the ID of the organization, `project_id` is the ID of the project you want to import, and `s3_user_id` is the ID of the S3 user to be imported.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_organization_project_s3_user.test_project_s3user[0]
    id = "<org_id,project_id,s3_user_id>"
}

resource "sys11iam_organization_project_s3_user" "test_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<name>"
  description = "<description>"
  org_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_organization_project.test_project.id
}

```

Now the resource to be imported can be managed with `terraform plan/apply`.
