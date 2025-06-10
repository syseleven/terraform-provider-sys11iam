Project S3 User Key Resource

The Project S3 User Key Resource manages an S3 Key for an S3 User. The access and secret key pairs are managed by this resource.

## Example Usage

```hcl
resource "sys11iam_project_s3user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  s3_user_id = sys11iam_project_s3user.test_terraform_project_s3user[0].id
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.terraform_test_project[0].id
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_project_s3user_access_key":

* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`s3_user_id`** - The UUID of the S3 User.

## Importing Organization Project S3 User Keys

To import an organization project S3 User key, your configuration would look like the following:

```hcl
resource "sys11iam_project_s3user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  s3_user_id = sys11iam_project_s3user.test_terraform_project_s3user[0].id
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.terraform_test_project[0].id
}

```
Then you execute:

```bash
terraform import sys11iam_project_s3user_key.test_terraform_project_s3_user_key[0] <organization_id,project_id,s3_user_id,s3_access_key>
```

Where `organization_id` is the ID of the organization, `project_id` is the ID of the project you want to import, `s3_user_id` is the ID of the S3 user to be imported, and `s3_access_key` is the access key of the S3 credential to be imported. The access and secret key of the S3 credential will be added to the Terraform state.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_project_s3user_key.test_terraform_project_s3_user_key[0]
    id = "<organization_id,project_id,s3_user_id,s3_access_key>"
}

resource "sys11iam_project_s3user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  s3_user_id = sys11iam_project_s3user.test_terraform_project_s3user[0].id
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.terraform_test_project[0].id
}

```

Now the resource to be imported can be managed with `terraform plan/apply`.

