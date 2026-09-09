# Changelog

Notable changes to the `syseleven/sys11iam` Terraform provider are documented in this file.

## [3.0.0]

Version 3 targets the new SysEleven IAM API (glue v3) and restructures the project-scoped resources under their
organization. If you are upgrading from a v1.x release, read the full [migration guide](docs/MIGRATION_GUIDE.md)
before upgrading.

### Breaking changes

* Renamed `sys11iam_project` to `sys11iam_organization_project` (v1 state migrates via `moved` blocks).
* Renamed `sys11iam_project_membership` to `sys11iam_organization_project_membership` (user memberships migrate
  via `moved` blocks; service account memberships must be imported manually).
* Renamed `sys11iam_project_s3user` to `sys11iam_organization_project_s3_user` (state migrates via `moved`
  blocks).
* Renamed `sys11iam_project_s3user_key` to `sys11iam_organization_project_s3_user_key` (state migrates via
  `moved` blocks).
* Removed `sys11iam_project_team`: project-level team permissions are now managed by the nested `projects` block
  of `sys11iam_organization_team`.
* Removed `sys11iam_project_team_membership`: team membership is now managed exclusively via
  `sys11iam_organization_team_membership` (manual migration).
* Renamed the `organization_id` attribute to `org_id` on all resources. The old name is deprecated but still
  accepted.
* Renamed the `s3_access_key` attribute to `access_key` on `sys11iam_organization_project_s3_user_key`.
* Renamed the team attribute `editable_permissions` to `organization_permissions`.
* The membership resources (`sys11iam_organization_membership`, `sys11iam_organization_project_membership`,
  `sys11iam_organization_team_membership`) now use a nested `membership` block instead of the previous flat
  attributes.

### Features

* The provider now runs against the glue v3 IAM API (rewritten `iam` client, regenerated schemas).
* State migration via Terraform `moved` blocks for all renamed resource types, including adopting a team from
  the state of a `sys11iam_project_team` resource.
* `sys11iam_organization_team` manages per-project permissions through a nested `projects` block and exposes
  organization-level permissions through `organization_permissions`.
* Membership resources use a nested `membership` block with user and service account variants, plus new computed
  attributes (`membership_type`, `team_name`, `project_name`).
* New `sys11iam_organization_project` data source to look up a project within an organization (#11).
* The `sys11iam_organization` data source can now be looked up by `id` **or** `name`.
* The deprecated `organization_id` attribute is still accepted as an alternative to `org_id` on all resources.
* New computed `keys` attribute on `sys11iam_organization_project_s3_user` exposing the user's S3 keys.
* New computed `is_managed_by_s11` attribute on `sys11iam_organization_project` indicating whether the project
  is managed by SysEleven.

### Fixes

* Fixed broken user state after state moves caused by a computed/non-computed field mismatch.
* Fixed `iam_url` not falling back to `https://iam.apis.syseleven.de` when no custom URL was configured (#10).
* Added the missing `is_managed_by_s11` attribute to the project resource and data source schemas.

### CI

* The pull request workflow now verifies that the generated code is in sync with `openapi.json`
  (`make tf-generate-check`); the OpenAPI spec is committed to the repository so the check works out of the box.
