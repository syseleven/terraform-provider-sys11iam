terraform {
  required_providers {
    ncs = {
      source = "hashicorp.com/syseleven/ncs"
    }
  }
}

provider "ncs" {
  #iam_url = "https://iam.stage-apis.syseleven.net/v1"
  oidc_url = "http://localhost:8181/realms/application/protocol/openid-connect/token"
  oidc_client_id = "pytest"
  oidc_client_secret = "YKjKvRHYtGjbxjsU2auNzcvt4FOaH5SK"
  oidc_client_scope = "pytest"
  iam_url = "http://127.0.0.1:9000"
}

resource "ncs_organization" "test_org" {
  name = "test_org"
  description = "testdescription"
  tags = ["testing"]
}

resource "ncs_project" "test_project" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  name = "test_project"
  description = "testdescription"
  tags = ["testing"]
  organization_id = ncs_organization.test_org.id
}

resource "ncs_organization_membership" "test_membership" {
  count = ncs_organization.test_org.is_active ? 1 : 0
  email = "test@syseleven.net"
  editable_permissions = ["can_become_project_administrator_in_org"]
  organization_id = ncs_organization.test_org.id
}

resource "ncs_project_membership" "test_project_membership" {
  depends_on = [ncs_project.test_project, ncs_organization_membership.test_membership]
  count = one(ncs_organization_membership.test_membership[*].is_active) == true ? 1 : 0
  email = "test@syseleven.net"
  permissions = ["can_become_administrator_in_project", "can_crud_permissions_in_project"]
  organization_id = ncs_organization.test_org.id
  project_id = ncs_project.test_project[0].id
}
