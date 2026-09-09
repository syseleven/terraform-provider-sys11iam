# Terraform Provider for the SysEleven Cloud IAM

This provider manages SysEleven Cloud IAM resources with [Terraform](https://www.terraform.io/): organizations,
projects, teams, memberships, service accounts, S3 users and keys, and organization contacts.

> **Upgrading from v1.x?** Version 3 targets the new SysEleven IAM API (glue v3) and renames or removes several
> resource types. Read the [migration guide](docs/MIGRATION_GUIDE.md) before upgrading your configuration.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.x
- [Go](https://go.dev/doc/install) 1.25 (for development)

## Authentication

The provider authenticates with a SysEleven IAM service account secret:

```hcl
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_..."
}
```

The following arguments are supported:

* **`serviceaccount_secret`** - The secret of a service account to authenticate with. If omitted, the
  `SYS11IAM_SERVICEACCOUNT_SECRET` environment variable is used.
* **`iam_url`** - The url to the IAM service. If omitted, the `SYS11IAM_IAM_URL` environment variable is used;
  otherwise, it defaults to `https://iam.apis.syseleven.de`.

## Usage

The user documentation is found at the [provider documentation](docs/index.md).

```hcl
terraform {
  required_providers {
    sys11iam = {
      source = "syseleven/sys11iam"
    }
  }
}

# Configure the sys11iam Provider for service account authentication
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_asdziuch-967s-aduc-123f-00asdasd8asd_9xjakshdkjOJPvk-36Fxxx"
}

# Look up an existing organization
data "sys11iam_organization" "testorg" {
  name = "test_org"
}
```

## Resources and Data Sources

| Type | Description |
| --- | --- |
| [`sys11iam_organization`](docs/resources/sys11iam_organization.md) | Manages a SysEleven IAM organization |
| [`sys11iam_organization_contact`](docs/resources/sys11iam_organization_contact.md) | Contact information of an organization |
| [`sys11iam_organization_membership`](docs/resources/sys11iam_organization_membership.md) | Organization membership of a user or service account |
| [`sys11iam_organization_project`](docs/resources/sys11iam_organization_project.md) | A project in an organization |
| [`sys11iam_organization_project_membership`](docs/resources/sys11iam_organization_project_membership.md) | Project membership of a user or service account |
| [`sys11iam_organization_project_s3_user`](docs/resources/sys11iam_organization_project_s3_user.md) | An S3 user in an organization project |
| [`sys11iam_organization_project_s3_user_key`](docs/resources/sys11iam_organization_project_s3_user_key.md) | An access key for an S3 user |
| [`sys11iam_organization_serviceaccount`](docs/resources/sys11iam_organization_serviceaccount.md) | A service account in an organization |
| [`sys11iam_organization_team`](docs/resources/sys11iam_organization_team.md) | A team in an organization, including its project permissions |
| [`sys11iam_organization_team_membership`](docs/resources/sys11iam_organization_team_membership.md) | Team membership of a user or service account |
| [`sys11iam_organization` (data source)](docs/data-sources/sys11iam_organization.md) | Look up an organization by its `id` or `name` |
| [`sys11iam_organization_project` (data source)](docs/data-sources/sys11iam_organization_project.md) | Look up a project by its `id` within an organization |

## Upgrading from v1.x to v3

Version 3 renames the project-scoped resource types so they are grouped under their organization. Add `moved`
blocks to your configuration before switching to the new type names and the provider will migrate the state:

```hcl
moved {
  from = sys11iam_project.example
  to   = sys11iam_organization_project.example
}

moved {
  from = sys11iam_project_membership.example
  to   = sys11iam_organization_project_membership.example
}

moved {
  from = sys11iam_project_s3user.example
  to   = sys11iam_organization_project_s3_user.example
}

moved {
  from = sys11iam_project_s3user_key.example
  to   = sys11iam_organization_project_s3_user_key.example
}
```

Additionally:

* `organization_id` is renamed to `org_id` (the old name is still accepted with a deprecation warning).
* `sys11iam_project_team` is replaced by the team's `projects` block and
  `sys11iam_project_team_membership` by `sys11iam_organization_team_membership` — both require a manual migration.

See the [migration guide](docs/MIGRATION_GUIDE.md) for the complete, step-by-step procedure including attribute
renames and the new membership schema.

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## Development

### Building

Run `make terraform-provider-sys11iam` to build. This results in the binary `terraform-provider-sys11iam`.

### Installing

Run `make dev` to build and install the binary to your local golang binary path. This should result in a
deployment to `~/go/bin/terraform-provider-sys11iam`, where `terraform` can find it.

### Running (e2e)

Copy `main.tf.example` to `main.tf` and adjust the example values to run it against a live environment, then run
`terraform init`, `terraform plan` and `terraform apply` from the repository root.

### Testing (unit)

Run `make unit-test` to run the unit tests including the `iam` and `rest` API clients.

### Code generation

The resource and data source schemas under `internal/resource_*/` and `internal/datasource_*/` are generated from
`openapi.json` with tfplugingen. Run `make tf-generate` to regenerate them or `make tf-generate-check` to verify
that the generated code is in sync with the OpenAPI spec (this check runs on every pull request).
