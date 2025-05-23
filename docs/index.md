# sys11iam Provider

The sys11iam provider is used to interact with the SysEleven NCS IAM. The provider needs to be configured with the proper credentials before it can be used.

## Example Usage

```hcl
# Define required providers
terraform {
  required_providers {
    sys11iam = {
      source = "syseleven/sys11iam"
    }
  }
}

# Configure the sys11iam Provider for service account authentication (see configuration for regular accounts below)
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_asdziuch-967s-aduc-123f-00asdasd8asd_9xjakshdkjOJPvk-36FJqasdmalkwaksnkajc"
  iam_url = "https://iam.apis.syseleven.de"
}

# Get an sys11iam organization
data "sys11iam_organization" "testorg" {
  id = "12345678-90ab-4cde-f123-4567890abcde"
  name = "test_org"
}

# Create an sys11iam project
resource "sys11iam_project" "test_project" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "testproject"
  description = "testdescription"
  tags = ["testing"]
  organization_id = data.sys11iam_organization.testorg.id
}

# Create an sys11iam organization membership
resource "sys11iam_organization_membership" "test_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  email = "test@example.com"
  affiliation = "member"
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = data.sys11iam_organization.testorg.id
}

# Create an sys11iam project membership
resource "sys11iam_project_membership" "test_project_membership" {
  depends_on = [sys11iam_project.test_project, sys11iam_organization_membership.test_membership]
  count = one(sys11iam_organization_membership.test_membership[*].is_active) == true ? 1 : 0
  email = "test@example.com"
  permissions = ["can_become_administrator_in_project", "can_crud_permissions_in_project"]
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}

# Create an sys11iam service account
resource "sys11iam_organization_serviceaccount" "test_serviceaccount" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "deploy"
  description = "deployment account"
  organization_id = data.sys11iam_organization.testorg.id
}

# Create an sys11iam organization contact
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

# Create an sys11iam organization team
resource "sys11iam_organization_team" "testorganization_team" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "testteam"
  description = "test team"
  tags = ["testing2"]
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = data.sys11iam_organization.testorg.id
}

# Create an sys11iam project team
resource "sys11iam_project_team" "test_project_team" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
  team_id = sys11iam_organization_team.testorganization_team[0].id
  editable_permissions = ["can_become_administrator_in_project"]
}

# Create an sys11iam organization team membership
resource "sys11iam_organization_team_membership" "testorganization_team_membership" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
  organization_id = data.sys11iam_organization.testorg.id
  team_id = sys11iam_organization_team.testorganization_team[0].id
}

# Create an sys11iam project S3User
resource "sys11iam_project_s3user" "test_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "tests3user"
  description = "test s3user"
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.test_project[0].id
}

# Create an SysEleven IAM  project S3 User
resource "sys11iam_project_s3user" "test_terraform_project_s3user" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  name = "terraform-tests3user-new-name"
  description = "terraform test s3user- new description"
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.terraform_test_project[0].id
}

# Create a SysEleven IAM project S3 User key
resource "sys11iam_project_s3user_key" "test_terraform_project_s3_user_key" {
  count = data.sys11iam_organization.testorg.is_active ? 1 : 0
  s3_user_id = sys11iam_project_s3user.test_terraform_project_s3user[0].id
  organization_id = data.sys11iam_organization.testorg.id
  project_id = sys11iam_project.terraform_test_project[0].id
}

```

Replacing above provider configuration:

```
# Configure the sys11iam Provider for service account user authentication
provider "ncs" {
  serviceaccount_secret = "s11_orgsa_asdziuch-967s-aduc-123f-00asdasd8asd_9xjakshdkjOJPvk-36FJqasdmalkwaksnkajc"
  iam_url = "https://iam.apis.syseleven.de"
}
````

## Configuration Reference

The following arguments are supported for the provider "ncs":

* **`iam_url`** - The url to the IAM service for creating organization, project, organization membership and project membership resources.
  If omitted, the `SYS11IAM_IAM_URL` environment variable is used.
* **`serviceaccount_secret`** - The secret of an service account to authenticate with. If omitted, the `SYS11IAM_SERVICEACCOUNT_SECRET` environment variable is used.
  
The following arguments are supported for the data source "sys11iam_organization":
* **`name`** - A unique name for the organization.
* **`id`** - The UUID of the organization.

The following arguments are supported for the resource "sys11iam_project":
* **`name`** - The name of the project.
* **`description`** - The description of the project.
* **`tags`** - The tags of the project.
* **`organization_id`** - The UUID of the organization.
* **`id`** - The UUID of the project. (read-only)

The following arguments are supported for the resource "sys11iam_organization_membership":

* **`email`** - The email of the user.
* **`affiliation`** - The member affiliation can be ("member" | "admin" | "owner")
* **`editable_permissions`** - The editable permissions of the user.
* **`organization_id`** - The UUID of the organization.
* **`is_active`** - Whether the organization membership is active or not. Organization membership activation is a manual step executed by the invited user. An invitation is issued by creating this resource. (default: false)
* **`id`** - The UUID of the organization membership. (read-only)

The following arguments are supported for the resource "sys11iam_project_membership":

* **`email`** - The email of the user.
* **`permissions`** - The editable permissions of the user.
* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`id`** - The UUID of the project membership. (read-only)

The following arguments are supported for the resource "sys11iam_organization_serviceaccount":

* **`name`** - The name of the service account.
* **`description`** - The description of the service account.
* **`organization_id`** - The UUID of the organization.

The following arguments are supported for the resource "sys11iam_organization_contact":

* **`first_name`** - The first name of the organization contact.
* **`last_name`** - The last name of the organization contact.
* **`notes`** - The notes concerning the organization contact.
* **`email`** - The email of the organization contact.
* **`phone`** - The phone number of the organization contact.
* **`roles`** - The roles of the organization contact.
* **`organization_id`** - The UUID of the organization.
* **`id`** - The UUID of the organization contact. (read-only)

The following arguments are supported for the resource "sys11iam_organization_team":

* **`name`** - The name of the organization team.
* **`description`** - The description of the organization team.
* **`tags`** - The tags of the organization team.
* **`editable_permissions`** - The editable permissions of the organization team.
* **`organization_id`** - The UUID of the organization.
* **`id`** - The UUID of the organization team. (read-only)

The following arguments are supported for the resource "sys11iam_project_team":

* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`team_id`** - The UUID of the organization team.
* **`editable_permissions`** - The editable permissions of the project team.

The following arguments are supported for the resource "sys11iam_organization_team_membership":

* **`id`** - The UUID of the regular user or service account to add as a member.
* **`organization_id`** - The UUID of the organization.
* **`team_id`** - The UUID of the organization team.

The following arguments are supported for the resource "sys11iam_project_s3user":

* **`name`** - The name of the S3User.
* **`description`** - The description of the S3User.
* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.

The following arguments are supported for the resource "sys11iam_project_s3user":
* **`name`** - The name of the S3 User.
* **`description`** - The description of the S3 User.
* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.

The following arguments are supported for the resource "sys11iam_project_s3user_access_key":

* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`s3_user_id`** - The UUID of the S3 User.

