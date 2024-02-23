terraform {
  required_providers {
    ncs = {
      source = "hashicorp.com/syseleven/ncs"
    }
  }
}

provider "ncs" {
  #host = "https://iam.stage-apis.syseleven.net/v1"
  host = "http://127.0.0.1:9000"
  oidc_url = "http://127.0.0.1:8181/realms/application/protocol/openid-connect/token"
  username = "admin"
  password = "admin"
}

resource "ncs_organization" "syseleven_spielt6" {
  name = "syseleven_spielt6"
}
