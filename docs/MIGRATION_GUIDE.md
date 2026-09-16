# Migration Guide: terraform-provider-sys11iam v1.5.4 → v3

This guide covers the breaking changes when migrating from the last released v1.x version (v1.5.4) to v3.

---

## Summary

Three categories of changes affect your Terraform configuration:

| # | Category | Scope | Migration |
|---|---|---|---|
| 1 | **Resource renames** | 4 resource types | Automatic via `moved` blocks |
| 2 | **Attribute renames** | All resources | `organization_id` → `org_id` |
| 3 | **Removed resources** | 2 resource types | Manual migration required |
| 4 | **In-place schema changes** | `sys11iam_organization_membership`, `sys11iam_organization` | Automatic state upgrade, config restructure required |

---

## 1. Resource Renames (Automatic State Migration)

The provider implements `ResourceWithMoveState` for the following renames. Add `moved` blocks to your configuration and the provider will migrate your state automatically.

### 1.1 `sys11iam_project` → `sys11iam_organization_project`

**Old configuration:**
```hcl
resource "sys11iam_project" "test_project" {
  name            = "project"
  description     = "my project"
  tags            = ["prod"]
  organization_id = data.sys11iam_organization.testorg.id
}
```

**Migration:**
```hcl
moved {
  from = sys11iam_project.test_project
  to   = sys11iam_organization_project.test_project
}

resource "sys11iam_organization_project" "test_project" {
  name        = "project"
  description = "my project"
  tags        = ["prod"]
  org_id      = data.sys11iam_organization.testorg.id
}
```

**State migration details:**
- `id` → `id` + `project_id` (both set to the project ID)
- `organization_id` → `org_id` (deprecated `organization_id` preserved)
- `description`, `name`, `tags` → unchanged
- Computed fields (`created_at`, `status`, `updated_at`) set to null, populated on next read

---

### 1.2 `sys11iam_project_s3user` → `sys11iam_organization_project_s3_user`

**Old configuration:**
```hcl
resource "sys11iam_project_s3user" "test_s3user" {
  name            = "s3user"
  description     = "my s3 user"
  organization_id = data.sys11iam_organization.testorg.id
  project_id      = sys11iam_project.test_project.id
}
```

**Migration:**
```hcl
moved {
  from = sys11iam_project_s3user.test_s3user
  to   = sys11iam_organization_project_s3_user.test_s3user
}

resource "sys11iam_organization_project_s3_user" "test_s3user" {
  name        = "s3user"
  description = "my s3 user"
  org_id      = data.sys11iam_organization.testorg.id
  project_id  = sys11iam_organization_project.test_project.id
}
```

**State migration details:**
- `id` → `id` + `s3_user_id`
- `organization_id` → `org_id`
- New `keys` attribute initialized as null (populated on next `terraform apply` from API)
- Computed fields set to null, populated on next read

---

### 1.3 `sys11iam_project_s3user_key` → `sys11iam_organization_project_s3_user_key`

**Old configuration:**
```hcl
resource "sys11iam_project_s3user_key" "test_key" {
  s3_access_key   = "AKIA..."
  s3_user_id      = sys11iam_project_s3user.test_s3user.id
  organization_id = data.sys11iam_organization.testorg.id
  project_id      = sys11iam_project.test_project.id
}
```

**Migration:**
```hcl
moved {
  from = sys11iam_project_s3user_key.test_key
  to   = sys11iam_organization_project_s3_user_key.test_key
}

resource "sys11iam_organization_project_s3_user_key" "test_key" {
  s3_user_id  = sys11iam_organization_project_s3_user.test_s3user.id
  org_id      = data.sys11iam_organization.testorg.id
  project_id  = sys11iam_organization_project.test_project.id
}
```

**State migration details:**
- `s3_access_key` → `access_key` (the state mover handles both old and new names)
- `secret_key` → `secret_key` (unchanged)
- `organization_id` → `org_id`
- `s3_user_id` → `s3_user_id` (unchanged)

> **Config attribute rename:** `s3_access_key` is now `access_key` in the HCL configuration. The state mover handles the old state attribute name (`s3_access_key` or `access_key`), but update your config to use `access_key`.

---

### 1.4 `sys11iam_project_membership` → `sys11iam_organization_project_membership`

This resource underwent a **significant schema change**. The old flat attributes (`email`, `permissions`) are now inside a nested `membership` block.

**Old configuration (user membership):**
```hcl
resource "sys11iam_project_membership" "user_project" {
  email           = "dev@example.com"
  permissions     = ["can_become_administrator_in_project"]
  organization_id = data.sys11iam_organization.testorg.id
  project_id      = sys11iam_project.test_project.id
}
```

**Migration:**
```hcl
moved {
  from = sys11iam_project_membership.user_project
  to   = sys11iam_organization_project_membership.user_project
}

resource "sys11iam_organization_project_membership" "user_project" {
  org_id     = data.sys11iam_organization.testorg.id
  project_id = sys11iam_organization_project.test_project.id
  id         = sys11iam_organization_membership.test_user[0].id

  membership = {
    user_membership = {
      permissions     = ["can_become_administrator_in_project"]
      membership_type = "user"
      user = {
        email = "dev@example.com"
      }
    }
  }
}
```

**State migration — user memberships:**

The built-in state mover migrates user memberships automatically if the old state contains a non-empty `email` attribute:
- `id` → `id`
- `organization_id` → `org_id`
- `project_id` → `project_id`
- `email` → `membership.user_membership.user.email`
- `permissions` → `membership.user_membership.permissions`
- `membership_type` set to `"user"`

**State migration — service account memberships (NOT automatic):**

If the old state is for a service account membership (no `email` attribute), the state mover will **reject the migration** with an error. Migrate these manually:

```bash
# 1. Remove the old resource from state
terraform state rm sys11iam_project_membership.sa_project

# 2. Update your config to the new v3 shape
resource "sys11iam_organization_project_membership" "sa_project" {
  org_id     = "org-id"
  project_id = "project-id"
  id         = "service-account-id"

  membership = {
    service_account_membership = {
      permissions     = ["can_become_administrator_in_project"]
      membership_type = "service_account"
      service_account = {
        id = "service-account-id"
      }
    }
  }
}

# 3. Import into state
terraform import sys11iam_organization_project_membership.sa_project <org_id>,<project_id>,<service_account_id>
```

**New v3 schema attributes:**

| Old attribute | New location | Notes |
|---|---|---|
| `email` | `membership.user_membership.user.email` | User memberships only |
| `permissions` | `membership.*.permissions` | Moved into membership block |
| `organization_id` | `org_id` | Renamed |
| `id` | `id` | Unchanged |
| *(new)* | `project_name` | Computed, populated on read |
| *(new)* | `membership_type` | Computed, derived from membership data |
| *(new)* | `membership.service_account_membership` | New block for service accounts |

---

## 2. Attribute Renames (All Resources)

### `organization_id` → `org_id`

Every resource that previously used `organization_id` now uses `org_id`. The old `organization_id` attribute is **deprecated** but still accepted for backward compatibility.

**Affected resources:**
- `sys11iam_organization_contact`
- `sys11iam_organization_membership`
- `sys11iam_organization_project`
- `sys11iam_organization_project_membership`
- `sys11iam_organization_project_s3_user`
- `sys11iam_organization_project_s3_user_key`
- `sys11iam_organization_serviceaccount`
- `sys11iam_organization_team`
- `sys11iam_organization_team_membership`

**Migration:**
```diff
-  organization_id = data.sys11iam_organization.testorg.id
+  org_id = data.sys11iam_organization.testorg.id
```

You can continue using `organization_id` temporarily — the provider will accept it and show a deprecation warning. Update to `org_id` to prepare for when `organization_id` is removed in a future release.

The provider's state upgrader automatically copies `organization_id` → `org_id` in existing Terraform state on the first apply after upgrading.

---

## 3. Removed Resources

### 3.1 `sys11iam_project_team` → `sys11iam_organization_team.projects`

The `sys11iam_project_team` resource no longer exists. Its functionality is now a nested `projects` block inside `sys11iam_organization_team`.

**Old configuration:**
```hcl
resource "sys11iam_organization_team" "my_team" {
  name            = "deployers"
  organization_id = data.sys11iam_organization.testorg.id
  description     = "Deployment team"
  project         = []
}

resource "sys11iam_project_team" "team_project_1" {
  organization_id      = data.sys11iam_organization.testorg.id
  project_id           = sys11iam_project.project_1.id
  team_id              = sys11iam_organization_team.my_team.id
  editable_permissions = ["can_become_administrator_in_project"]
}
```

**New configuration:**
```hcl
resource "sys11iam_organization_team" "my_team" {
  name                     = "deployers"
  org_id                   = data.sys11iam_organization.testorg.id
  description              = "Deployment team"
  organization_permissions = []
  tags                     = []

  projects = [
    {
      id                  = sys11iam_organization_project.project_1.id
      project_permissions = ["can_become_administrator_in_project"]
    }
  ]
}
```

**Automatic migration (one `sys11iam_project_team` per team):**

The provider supports a state move that adopts the team through the old `project_team` state and folds its `editable_permissions` into a single `projects` entry (`id` = `team_id`, so `terraform state rm` the team from state first if the target address already has state). The team's name, description, tags and organization permissions are re-read from the API on the first refresh:

```hcl
moved {
  from = sys11iam_project_team.team_project_1
  to   = sys11iam_organization_team.my_team
}
```

```bash
terraform state rm sys11iam_organization_team.my_team
```

**Manual migration (multiple `sys11iam_project_team` per team):**

Terraform does not allow several `moved` blocks targeting the same resource, so teams that had more than one old `sys11iam_project_team` resource must be migrated manually:

1. For each `sys11iam_project_team`, note the `project_id` and `editable_permissions`
2. Remove all of them from state: `terraform state rm sys11iam_project_team.<name>`
3. Remove the `sys11iam_project_team` resources from your configuration
4. Update your `sys11iam_organization_team` resource with a `projects` block containing one entry per old project-team resource
5. Rename `editable_permissions` to `organization_permissions` and `organization_id` to `org_id` in the team config
6. `terraform apply`

---

### 3.2 `sys11iam_project_team_membership` → `sys11iam_organization_team_membership`

The `sys11iam_project_team_membership` resource is no longer available. Team membership is now managed exclusively through `sys11iam_organization_team_membership`.

**Migration steps:**
1. Collect `team_id` and `member_id` from state
2. `terraform state rm sys11iam_project_team_membership.<name>`
3. Create `sys11iam_organization_team_membership` in the new schema
4. Import: `terraform import sys11iam_organization_team_membership.<name> <org_id>,<team_id>,<member_id>`

> **Why no automatic migration?** The old state lacks the member type (user vs. service account) needed to infer the v3 nested `membership` block shape.

---

## 4. Additional Schema Changes

### `sys11iam_organization_membership` — flat attributes to nested `membership` block

The v1.5.4 attributes `email`, `affiliation` and `editable_permissions` are now inside a nested `membership`
block. The in-place state upgrade (no `moved` block needed, the type name is unchanged) converts the old state
automatically:

| Old attribute | New location |
|---|---|
| `email` | `membership.user_membership.email` |
| `affiliation` | `membership.user_membership.affiliation` |
| `editable_permissions` | `membership.user_membership.permissions` |
| `organization_id` | `org_id` (deprecated `organization_id` preserved) |
| `id` | `id` |
| `is_active` | *(removed)* |
| *(new)* | `membership.user_membership.membership_type` set to `"user"` |

**Migration:**
```diff
-  email                = "test@example.com"
-  affiliation          = "member"
-  editable_permissions = ["can_become_project_administrator_in_org"]
-  organization_id      = data.sys11iam_organization.testorg.id
+  org_id = data.sys11iam_organization.testorg.id
+  membership = {
+    user_membership = {
+      email             = "test@example.com"
+      affiliation       = "member"
+      permissions       = ["can_become_project_administrator_in_org"]
+      membership_type   = "user"
+    }
+  }
```

> If the old state has no `email` value, the state upgrade is rejected with an error. Remove the resource from
> state and import it instead: `terraform import sys11iam_organization_membership.<name> <org_id>,<member_id>`

### `sys11iam_organization` — flat `company_info_*` attributes to nested `company_info` block

The v1.5.4 flat `company_info_*` attributes are now a nested `company_info` block. The in-place state upgrade
converts the old values automatically:

| Old attribute | New location |
|---|---|
| `company_info_street` | `company_info.street` |
| `company_info_street_number` | `company_info.street_number` |
| `company_info_zip_code` | `company_info.zip_code` |
| `company_info_city` | `company_info.city` |
| `company_info_country` | `company_info.country` |
| `company_info_vat_id` | `company_info.vat_id` |
| `company_info_preferred_billing_method` | `company_info.preferred_billing_method` |
| `company_info_phone` | `company_info.phone_number` |
| `company_info_accepted_tos` | `company_info.accepted_tos` |
| `company_info_company_name` | `company_info.company_name` |

All other attributes (`id`, `name`, `description`, `is_active`, `tags`, `created_at`, `updated_at`) are
unchanged. If the old state contained no company information, `company_info` is upgraded to null and the
(required) `company_info` block must be added to your configuration.

**Migration:**
```diff
-  company_info_street                   = "Examplestreet"
-  company_info_street_number            = "1"
-  company_info_zip_code                 = "80331"
-  company_info_city                     = "Munich"
-  company_info_country                  = "Germany"
-  company_info_vat_id                   = "DE123456789"
-  company_info_preferred_billing_method = "invoice"
-  company_info_phone                    = "+49 89 123456"
-  company_info_accepted_tos             = true
-  company_info_company_name             = "SysEleven GmbH"
+  company_info = {
+    street                   = "Examplestreet"
+    street_number            = "1"
+    zip_code                 = "80331"
+    city                     = "Munich"
+    country                  = "Germany"
+    vat_id                   = "DE123456789"
+    preferred_billing_method = "invoice"
+    phone_number             = "+49 89 123456"
+    accepted_tos             = true
+    company_name             = "SysEleven GmbH"
+  }
```

### `sys11iam_organization_team` — `editable_permissions` → `organization_permissions`

The v1.5.4 attribute `editable_permissions` (the team's organization-level permissions) is now `organization_permissions`. The in-place state upgrade (no `moved` block needed, the type name is unchanged) copies the old value over automatically:

```diff
-  editable_permissions = ["can_become_project_administrator_in_org"]
+  organization_permissions = ["can_become_project_administrator_in_org"]
```

The v1.5.4 `sys11iam_project_team` resources are now represented by the `projects` block of the team resource (see Section 3.1).

### `sys11iam_organization_project_s3_user` — new `keys` attribute

Computed list of key objects with `access_key`, `created_at`, `created_by`, `secret_key`, `updated_at`.

### `sys11iam_organization_team_membership` — new attributes

- `team_name` — computed, name of the team
- `membership_type` — derived from membership data
- Nested `membership` block with `user_team_membership` / `service_account_team_membership`

### `sys11iam_organization_project` — new `is_managed_by_s11` attribute

Computed boolean attribute indicating whether the project is managed by SysEleven. It is populated on read; no
configuration change is required.

### Data sources

The `sys11iam_organization` data source can now look up an organization by its `id` **or** its unique `name`.
Providing both still works but prints a deprecation warning. A new `sys11iam_organization_project` data source
looks up a project by its `id` within an organization.

---

## 5. Migration Procedure

1. **Update the provider** — point `required_providers` to the new version.
2. **Add `moved` blocks** — for Section 1 renamed resources.
3. **Rename resources** — update resource type names.
4. **Rename attributes** — `organization_id` → `org_id`, `s3_access_key` → `access_key`, `editable_permissions` → `organization_permissions` (teams).
5. **Restructure membership and organization** — flat attributes into nested `membership` block
   (`sys11iam_organization_membership`, `sys11iam_organization_project_membership`) and flat `company_info_*`
   attributes into the `company_info` block (`sys11iam_organization`).
6. **Handle removed resources** — `moved` block for a single `sys11iam_project_team` per team, manual otherwise; `sys11iam_project_team_membership` remains manual (Section 3).
7. **`terraform plan`** — verify; expect state migration output and read operations.
8. **`terraform apply`** — apply.

---

## 6. Complete Resource Name Mapping

| v1.5.4 resource | v3 resource | Migration |
|---|---|---|
| `sys11iam_project` | `sys11iam_organization_project` | ✅ `moved` block |
| `sys11iam_project_membership` | `sys11iam_organization_project_membership` | ✅ user (email in state); ❌ service account (manual) |
| `sys11iam_project_s3user` | `sys11iam_organization_project_s3_user` | ✅ `moved` block |
| `sys11iam_project_s3user_key` | `sys11iam_organization_project_s3_user_key` | ✅ `moved` block |
| `sys11iam_project_team` | `sys11iam_organization_team.projects` | ✅ `moved` block (one per team); ❌ manual if the team had multiple project-team resources |
| `sys11iam_project_team_membership` | `sys11iam_organization_team_membership` | ❌ manual |
