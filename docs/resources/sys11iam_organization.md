# Organization Resource

The Organization Resource manages a SysEleven IAM organization, including the onboarding (company) information.

> Creating an organization starts the SysEleven onboarding process. Activating a newly created organization is a
> manual step (contact the SysEleven sales team or use the SysEleven dashboard). The provider emits a warning
> while the organization is not active yet.

## Example Usage

```hcl
resource "sys11iam_organization" "testorg" {
  name        = "test_org"
  description = "my organization"
  tags        = ["testing"]

  company_info = {
    accepted_tos             = true
    city                     = "Munich"
    company_name             = "Example GmbH"
    country                  = "Germany"
    preferred_billing_method = "example"
    street                   = "Examplestreet"
    street_number            = "1"
    vat_id                   = "DE123456789"
    zip_code                 = "80331"
  }
}
```

## Argument Reference

The following arguments are supported for the resource "sys11iam_organization":

* **`name`** - A unique name for the organization (3-62 characters, lowercase letters, digits and hyphens).
* **`description`** - A description for the organization.
* **`tags`** - The tags of the organization.
* **`company_info`** - Required block with the onboarding information of the organization (see
  [Company Info](#company-info) below).

### Company Info

The following arguments are supported in the `company_info` block:

* **`accepted_tos`** - Whether the organization creator has accepted the terms of service. (required)
* **`city`** - The city where the company resides. (required)
* **`company_name`** - The legal name of the company the organization belongs to. (required)
* **`country`** - The country where the company resides. (required)
* **`phone_number`** - The phone number of the organization. (optional)
* **`preferred_billing_method`** - The preferred billing method. (required)
* **`street`** - The street where the company resides. (required)
* **`street_number`** - The street number of the company's address. (required)
* **`vat_id`** - The VAT ID of the company. (required)
* **`zip_code`** - The zip code of the company. (required)

## Attributes Reference

The following attributes are exported:

* **`id`** - The UUID of the organization. (read-only)
* **`org_id`** - The UUID of the organization. (read-only)
* **`is_active`** - Whether the organization is active. (read-only)
* **`created_at`** - The timestamp of the organization's creation. (read-only)
* **`updated_at`** - The timestamp of the organization's last update. (read-only)

## Importing Organizations

To import an existing organization, your configuration would look like the following:

```hcl
resource "sys11iam_organization" "testorg" {
  name = "<organization name>"

  company_info = {
    accepted_tos             = true
    city                     = "<city>"
    company_name             = "<company name>"
    country                  = "<country>"
    preferred_billing_method = "<billing method>"
    street                   = "<street>"
    street_number            = "<street number>"
    vat_id                   = "<vat id>"
    zip_code                 = "<zip code>"
  }
}
```

Then you execute:

```bash
terraform import sys11iam_organization.testorg <org_id>
```

Where `org_id` is the ID of the organization you want to import. The `company_info` values are not read back
from the API, so set them in your configuration after importing.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
  to = sys11iam_organization.testorg
  id = "<org_id>"
}

resource "sys11iam_organization" "testorg" {
  name = "<organization name>"

  company_info = {
    accepted_tos             = true
    city                     = "<city>"
    company_name             = "<company name>"
    country                  = "<country>"
    preferred_billing_method = "<billing method>"
    street                   = "<street>"
    street_number            = "<street number>"
    vat_id                   = "<vat id>"
    zip_code                 = "<zip code>"
  }
}
```

Now the resource to be imported can be managed with `terraform plan/apply`.
