# NCS Provider

The NCS provider is used to interact with the SysEleven NCS IAM. The provider needs to be configured with the proper credentials before it can be used.

## Example Usage

```hcl
# Define required providers
terraform {
  required_providers {
    ncs = {
      source = "hashicorp.com/syseleven/ncs"
    }
  }
}

# Configure the NCS Provider for regular user authentication (see configuration for service accounts below)
provider "ncs" {
  oidc_url = "http://127.0.0.1:8181/realms/application/protocol/openid-connect/token"
  oidc_client_username = "admin"
  oidc_client_password = "admin"
  oidc_client_id = "pytest"
  oidc_client_secret = "YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK"
  oidc_client_scope = "pytest"
  iam_url = "http://127.0.0.1:9000"
}

# Create an NCS organization
resource "ncs_organization" "test_org" {
  name = "test_org"
  description = "testdescription"
  tags = ["testing"]
  company_info_street = "teststreet"
  company_info_street_number = "1"
  company_info_zip_code = "12345"
  company_info_city = "testcity"
  company_info_country = "testland"
  company_info_vat_id = "42069"
  company_info_preferred_billing_method = "SEPA"
  company_info_phone = "+49123456789"
  company_info_accepted_tos = true
  company_info_company_name = "testcompany"
}

# Create an NCS project
resource "ncs_project" "test_project" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  name = "test_project"
  description = "testdescription"
  tags = ["testing"]
  organization_id = ncs_organization.test_org.id
}

# Create an NCS organization membership
resource "ncs_organization_membership" "test_membership" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  email = "test@syseleven.net"
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = ncs_organization.test_org.id
}

# Create an NCS project membership
resource "ncs_project_membership" "test_project_membership" {
  depends_on = [ncs_project.test_project, ncs_organization_membership.test_membership]
  count = one(ncs_organization_membership.test_membership[*].is_active) == true ? 1 : 0
  email = "test@syseleven.net"
  permissions = ["can_become_administrator_in_project", "can_crud_permissions_in_project"]
  organization_id = ncs_organization.test_org.id
  project_id = ncs_project.test_project[0].id
}

# Create an NCS service account
resource "ncs_organization_serviceaccount" "test_serviceaccount" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  name = "deploy"
  description = "deployment account"
  organization_id = ncs_organization.test_org.id
}

# Create an NCS organization contact
resource "ncs_organization_contact" "test_organization_contact" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  first_name = "Test"
  last_name = "Contact"
  notes = "test notes"
  email = "test@syseleven.net"
  phone = "+491684941254823"
  roles = ["Technical"]
  organization_id = ncs_organization.test_org.id
}

# Create an NCS organization team
resource "ncs_organization_team" "test_organization_team" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  name = "Testteam"
  description = "test team"
  tags = ["testing2"]
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = ncs_organization.test_org.id
}

# Create an NCS project eam
resource "ncs_project_team" "test_project_team" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  organization_id = ncs_organization.test_org.id
  project_id = ncs_project.test_project[0].id
  team_id = ncs_organization_team.test_organization_team[0].id
  editable_permissions = ["can_become_administrator_in_project"]
}

# Create an NCS organization team membership
resource "ncs_organization_team_membership" "test_organization_team_membership" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  id = ncs_organization_serviceaccount.test_serviceaccount[0].id
  organization_id = ncs_organization.test_org.id
  team_id = ncs_organization_team.test_organization_team[0].id
  editable_permissions = ["can_become_project_administrator_in_org"]
}

# Create an NCS project S3User
resource "ncs_project_s3user" "test_project_s3user" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  name = "tests3user"
  description = "test s3user"
  organization_id = ncs_organization.test_org.id
  project_id = ncs_project.test_project[0].id
}

```

Replacing above provider configuration:
```
# Configure the NCS Provider for regular user authentication
provider "ncs" {
  oidc_url = "http://127.0.0.1:8181/realms/application/protocol/openid-connect/token"
  oidc_client_username = "admin"
  oidc_client_password = "admin"
  oidc_client_id = "pytest"
  oidc_client_secret = "YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK"
  oidc_client_scope = "pytest"
  iam_url = "http://127.0.0.1:9000"
}
```

with the following, allows you to authenticate with a service account instead of a regular user account, but disallows organization creation:

```
# Configure the NCS Provider for service account user authentication
provider "ncs" {
  oidc_url = "http://127.0.0.1:8181/realms/application/protocol/openid-connect/token"
  oidc_client_secret = "s11_orgsa_my-s-e-c-r_e_tserviceaccountcredential"
  iam_url = "http://127.0.0.1:9000"
}
````

## Configuration Reference

The following arguments are supported for the provider "ncs":

* **`oidc_url`** - The identity authentication URL.
  If omitted, the `NCS_OIDC_URL` environment variable is used.
* **`oidc_client_id`** - The ID of an application credential to authenticate with. An`oidc_client_secret` and `oidc_client_scope` has to bet set along with this parameter.
  If omitted, the `NCS_OIDC_CLIENT_ID` environment variable is used.
* **`oidc_client_secret`** - The secret of an application credential to authenticate with. Required by `oidc_client_id`. When `oidc_client_id` is not set, the value is used for a service account authentication credential.
  If omitted, the `NCS_OIDC_CLIENTSECRET` environment variable is used.
* **`oidc_client_scope`** - The scope of an application credential to authenticate with. Required by `oidc_client_id`.
* **`oidc_client_username`** - The regular user username to authenticate with. Required by `oidc_client_id`.
* **`oidc_client_password`** - The regular user password to authenticate with. Required by `oidc_client_id`.
  If omitted, the `NCS_OIDC_CLIENT_SCOPE` environment variable is used.
* **`iam_url`** - The url to the IAM service for creating organization, project, organization membership and project membership resources.
  If omitted, the `NCS_IAM_URL` environment variable is used.
  
The following arguments are supported for the resource "ncs_organization":
* **`name`** - A unique name for the organization.
* **`description`** - A description for the organization.
* **`tags`** - The tags of the organization.
* **`is_active`** - Whether the organization is active or not. Organization activation is manual step executed by an IAM administrator. (default: false)
* **`id`** - The UUID of the organization. (read-only)
* **`created_at`** - The time the resource was created. (read-only)
* **`updated_at`** - The time the resource was last updated. (read-only)


The following arguments are supported for the resource "ncs_project":
* **`name`** - The name of the project.
* **`description`** - The description of the project.
* **`tags`** - The tags of the project.
* **`organization_id`** - The UUID of the organization.
* **`id`** - The UUID of the project. (read-only)

The following arguments are supported for the resource "ncs_organization_membership":

* **`email`** - The email of the user.
* **`editable_permissions`** - The editable permissions of the user. Choose from:
  * `can_become_project_administrator_in_org`
  * `can_create_projects_in_org`
  * `can_invite_members_in_org`
  * `can_crud_permissions_in_org`
  * `can_read_members_in_org`
  * `can_delete_members_in_org`
* **`organization_id`** - The UUID of the organization.
* **`is_active`** - Whether the organization membership is active or not. Organization membership activation is a manual step executed by the invited user. An invitation is issued by creating this resource. (default: false)
* **`id`** - The UUID of the organization membership. (read-only)

The following arguments are supported for the resource "ncs_project_membership":

* **`email`** - The email of the user.
* **`permissions`** - The editable permissions of the user. Choose from:
  * `can_become_administrator_in_project`
  * `can_crud_permissions_in_project`
  * `can_become_viewer_in_openstack`
  * `can_become_editor_in_openstack`
  * `can_read_project_in_project`
  * `can_delete_project_in_project`
  * `can_create_api_keys_in_project`
  * `can_create_clusters_in_metakube`
  * `can_read_clusters_in_metakube`
  * `can_update_clusters_in_metakube`
  * `can_delete_clusters_in_metakube`
  * `can_create_ssh_keys_in_metakube`
  * `can_read_ssh_keys_in_metakube`
  * `can_update_ssh_keys_in_metakube`
  * `can_delete_ssh_keys_in_metakube`
  * `can_create_databases_in_dbaas`
  * `can_read_databases_in_dbaas`
  * `can_update_databases_in_dbaas`
  * `can_delete_databases_in_dbaas`
  * `can_read_backups_from_databases_in_dbaas`
  * `can_become_reader_in_horus`
  * `can_become_editor_in_horus`
  * `can_become_admin_in_horus`
* **`organization_id`** - The UUID of the organization.
* **`project_id`** - The UUID of the project.
* **`id`** - The UUID of the project membership. (read-only)
