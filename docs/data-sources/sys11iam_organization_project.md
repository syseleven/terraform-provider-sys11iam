# sys11iam_organization_project

Get an Organization Project by its ID within an organization.

## Example Usage

Look up a project by ID:

```hcl
data "sys11iam_organization_project" "existing" {
  org_id = data.sys11iam_organization.testorg.id
  id     = "1234567890ab4cdef1234567890abcde"
}
```

The data source can be used with any project-scoped resource:

```hcl
resource "sys11iam_organization_project_s3_user" "existing_project_user" {
  org_id     = data.sys11iam_organization_project.existing.org_id
  project_id = data.sys11iam_organization_project.existing.id
  name       = "terraform-user"
}
```

## Argument Reference

The following arguments are supported:

* **`org_id`** - The UUID of the organization containing the project.
* **`id`** - The UUID of the project in hex format.

## Attributes Reference

The following attributes are exported:

* **`id`** - The UUID of project in hex format.
* **`org_id`** - The UUID of the organization containing the project.
* **`name`** - The name of the project.
* **`description`** - The description of the project.
* **`is_managed_by_s11`** - Whether the project is managed by S11.
* **`status`** - The status of the project.
* **`tags`** - The tags of the project.
* **`created_at`** - The timestamp of the project's creation.
* **`updated_at`** - The timestamp of the project's last update.
