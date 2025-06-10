# Organization Contact resource

The Organization Contact Resource defines the corresponding contact information in SysEleven IAM.

## Example Usage

```hcl
resource "sys11iam_organization_contact" "testorganization_contact" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  first_name = "Test"
  last_name = "Contact"
  notes = "test notes"
  email = "test@example.com"
  phone = "+491684941254823"
  roles = ["Technical"]
  organization_id = data.sys11iam_organization.testorg.id
}

```

## Argument Reference

The following arguments are supported:

* **`first_name`** - The first name of the organization contact.
* **`last_name`** - The last name of the organization contact.
* **`notes`** - The notes concerning the organization contact.
* **`email`** - The email of the organization contact.
* **`phone`** - The phone number of the organization contact.
* **`roles`** - The roles of the organization contact.
* **`organization_id`** - The UUID of the organization.
* **`id`** - The UUID of the organization contact. (read-only)

## Importing Contacts

To import an organization contact, your configuration would look like the following:

```hcl
resource "sys11iam_organization_contact" "testorganization_contact" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  first_name = "<first name>"
  last_name = "<last name>"
  notes = "<notes>"
  email = "<email>"
  phone = "<phone>"
  roles = []
  organization_id = data.sys11iam_organization.testorg.id
}
```
Then you execute:

```bash
terraform import sys11iam_organization_contact.testorganization_contact[0] <organization_id,contact_id>
```

Where `organization_id` is the ID of the organization and `contact_id` is the ID of the contact you want to import.

A programmatic alternative involves using the [import block](https://developer.hashicorp.com/terraform/language/import#syntax):

```hcl
import {
  to = sys11iam_organization_contact.testorganization_contact[0] 
  id = "<organization_id,contact_id>"
}

resource "sys11iam_organization_contact" "testorganization_contact" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  first_name = "<first name>"
  last_name = "<last name>"
  notes = "<notes>"
  email = "<email>"
  phone = "<phone>"
  roles = []
  organization_id = data.sys11iam_organization.testorg.id
}

```

Now the resource to be imported can be managed with `terraform plan/apply`.



