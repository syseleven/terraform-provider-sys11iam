# sys11iam_organization

Get an Organization by its ID or unique name. At least one of `id` or `name` must be provided. Both are accepted for compatibility, but configuring only one is recommended.

## Example Usage

Look up an organization by name:

```hcl
data "sys11iam_organization" "testorg" {
  name = "test_org"
}
```

Look up an organization by ID:

```hcl
data "sys11iam_organization" "testorg" {
  id = "12345678-90ab-4cde-f123-4567890abcde"
}
```

The data source can be used with any resource:

```hcl
resource "sys11iam_organization_project_s3_user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  org_id = data.sys11iam_organization.testorg.id
  # ...
}
```

## Argument Reference

The following arguments are supported:

* **`id`** - The UUID of the organization. At least one of `id` or `name` must be provided; both are accepted for compatibility.
* **`name`** - A unique name for the organization. At least one of `id` or `name` must be provided; both are accepted for compatibility.

## Attributes Reference

The following attributes are exported:

* **`id`** - The UUID of the organization.
* **`name`** - The name of the organization.
* **`is_active`** - Whether the organization is active.
* **`description`** - A description of the organization.
* **`tags`** - The tags of the organization.
* **`created_at`** - The timestamp of the organization's creation.
* **`updated_at`** - The timestamp of the organization's last update.

### Company Info

* **`company_info_street`** - The organization's street name.
* **`company_info_street_number`** - The organization's street number.
* **`company_info_zip_code`** - The organization's zip code.
* **`company_info_city`** - The organization's city.
* **`company_info_country`** - The organization's country.
* **`company_info_vat_id`** - The organization's VAT ID.
* **`company_info_preferred_billing_method`** - The organization's preferred billing method.
* **`company_info_phone`** - The organization's phone number.
* **`company_info_accepted_tos`** - Whether the organization has accepted the terms of service.
* **`company_info_company_name`** - The organization's company name.
