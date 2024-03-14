terraform {
  required_providers {
    ncs = {
      source = "hashicorp.com/syseleven/ncs"
    }
  }
}

provider "ncs" {
  #iam_url = "https://iam.stage-apis.syseleven.net/v1"
  oidc_url = "http://127.0.0.1:8181/realms/application/protocol/openid-connect/token"
  oidc_username = "admin"
  oidc_password = "admin"
  iam_url = "http://127.0.0.1:9000"
}

resource "ncs_organization" "test_org" {
  name = "test_org"
  description = "testdescription"
  tags = ["testing"]
}

resource "ncs_project" "test_project" {
  name = "test_project"
  description = "testdescription"
  tags = ["testing"]
  organization_id = ncs_organization.test_org.id
}

resource "ncs_organization_membership" "test_membership" {
  email = "test@syseleven.net"
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = ncs_organization.test_org.id
}

resource "ncs_project_membership" "test_project_membership" {
  email = "test@syseleven.net"
  permissions = ["can_become_administrator_in_project", "can_crud_permissions_in_project"]
  organization_id = ncs_organization.test_org.id
  project_id = ncs_project.test_project.id
}
