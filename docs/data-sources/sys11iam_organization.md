# sys11iam_organization

Get an Organization by its ID and name.

## Example Usage

```hcl
data "sys11iam_organization" "testorg" {
  id = "12345678-90ab-4cde-f123-4567890abcde"
  name = "test_org"
}

# now the data source can be used with any resource
resource "sys11iam_project_s3user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  organization_id = data.sys11iam_organization.testorg.id
  # ...
}
```

## Argument Reference

The following arguments are supported for the data source "sys11iam_organization":
* **`name`** - A unique name for the organization.
* **`id`** - The UUID of the organization.

