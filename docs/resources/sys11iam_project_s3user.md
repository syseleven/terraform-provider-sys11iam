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

