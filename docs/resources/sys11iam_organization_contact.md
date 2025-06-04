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
  first_name = ""
  last_name = ""
  notes = ""
  email = ""
  phone = ""
  roles = []
  organization_id = data.sys11iam_organization.testorg.id # or ""
}
```
Then you execute:

```bash
terraform import sys11iam_organization_contact.testorganization_contact <organization_id,contact_id>
```

Where `organization_id` is the ID of the organization and `contact_id` is the ID of the contact you want to import.


