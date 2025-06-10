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

## Importing Organization Projects

To import an organization project, your configuration would look like the following:

```hcl
resource "sys11iam_project" "test_project" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<project name>"
  description = "<description>"
  tags = []
  organization_id = data.sys11iam_organization.testorg.id
}

```
Then you execute:

```bash
terraform import sys11iam_project.test_project[0] <organization_id,project_id>
```

Where `organization_id` is the ID of the organization and `project_id` is the ID of the project you want to import.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_project.test_project[0] 
    id = <organization_id,project_id>
}

resource "sys11iam_project" "test_project" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<project name>"
  description = "<description>"
  tags = []
  organization_id = data.sys11iam_organization.testorg.id
}

```
Now the resource to be imported can be managed with `terraform plan/apply`.

