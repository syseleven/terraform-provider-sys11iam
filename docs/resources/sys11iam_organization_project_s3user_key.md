# Organization Project S3 User Key Resource

The Project S3 User Key Resource manages an S3 Key for an S3 User. The access and secret key pairs are managed by this resource.

## Example Usage

```hcl
resource "sys11iam_organization_project_s3_user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  s3_user_id = sys11iam_organization_project_s3_user.test_terraform_project_s3user[0].id
  org_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_organization_project.terraform_test_project.id
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization_project_s3_user_key":

* **`org_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`s3_user_id`** - The UUID of the S3 User.

## Importing Organization Project S3 User Keys

To import an organization project S3 User key, your configuration would look like the following:

```hcl
resource "sys11iam_organization_project_s3_user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  s3_user_id = sys11iam_organization_project_s3_user.test_terraform_project_s3user[0].id
  org_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_organization_project.terraform_test_project.id
}

```
Then you execute:

```bash
terraform import sys11iam_organization_project_s3_user_key.test_terraform_organization_project_s3_user_key[0] <org_id,project_id,s3_user_id,access_key>
```

Where `org_id` is the ID of the organization, `project_id` is the ID of the project you want to import, `s3_user_id` is the ID of the S3 user to be imported, and `access_key` is the access key of the S3 credential to be imported. The access and secret key of the S3 credential will be added to the Terraform state.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_organization_project_s3_user_key.test_terraform_project_s3_user_key[0]
    id = "<org_id,project_id,s3_user_id,access_key>"
}

resource "sys11iam_organization_project_s3_user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  s3_user_id = sys11iam_organization_project_s3_user.test_terraform_project_s3user[0].id
  org_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_organization_project.terraform_test_project.id
}
```

Now the resource to be imported can be managed with `terraform plan/apply`.
