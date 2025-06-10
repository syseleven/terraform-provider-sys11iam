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

## Importing Organization Service Accounts

To import an organization service account, your configuration would look like the following:

```hcl
resource "sys11iam_organization_serviceaccount" "test_serviceaccount" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<service account name>"
  description = "<service account description>"
  organization_id = data.sys11iam_organization.testorg.id
}

```
Then you execute:

```bash
terraform import sys11iam_organization_serviceaccount.test_serviceaccount[0] <organization_id,service_account_id>
```

Where `organization_id` is the ID of the organization and `service_account_id` is the ID of the organization service account you want to import.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
    to = sys11iam_organization_serviceaccount.test_serviceaccount[0] 
    id = "<organization_id,service_account_id>"
}

resource "sys11iam_organization_serviceaccount" "test_serviceaccount" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "<service account name>"
  description = "<service account description>"
  organization_id = data.sys11iam_organization.testorg.id
}
```
Now the resource to be imported can be managed with `terraform plan/apply`.

