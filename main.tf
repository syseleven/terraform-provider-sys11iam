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

resource "ncs_organization" "syseleven_spielt6" {
  name = "syseleven_spielt6"
}
