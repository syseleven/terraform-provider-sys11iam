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

