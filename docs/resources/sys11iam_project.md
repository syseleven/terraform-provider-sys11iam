Project Resource

The Project Resource manages a SysEleven IAM project in an Organization.

## Example Usage

```hcl
resource "sys11iam_project" "test_project" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "testproject"
  description = "testdescription"
  tags = ["testing"]
  organization_id = data.sys11iam_organization.testorg.id
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_project":
* **`name`** - The name of the project.
* **`description`** - The description of the project.
* **`tags`** - The tags of the project.
* **`organization_id`** - The UUID of the organization.
* **`id`** - The UUID of the project. (read-only)

