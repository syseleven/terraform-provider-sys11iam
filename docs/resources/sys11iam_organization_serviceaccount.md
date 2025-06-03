# Organization Service Account Resource

The Organization Service Account Resource allows the management of a service account in SysEleven IAM.

## Example Usage

```hcl
resource "sys11iam_organization_serviceaccount" "test_serviceaccount" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "deploy"
  description = "deployment account"
  organization_id = data.sys11iam_organization.testorg.id
}
```

## Argument Reference
The following arguments are supported for the resource "sys11iam_organization_serviceaccount":

* **`name`** - The name of the service account.
* **`description`** - The description of the service account.
* **`organization_id`** - The UUID of the organization. Can be hardcoded or (recommended) passed in via the organization data source.

