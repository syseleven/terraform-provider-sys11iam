package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccOrganizationProjectMembershipResourceUser(t *testing.T) {
	params := &testAccCheckOrganizationProjectMembershipResourceParams{
		MemberType:           "user",
		MemberId:             os.Getenv("MEMBER_ID"),
		ProjectPermissions:   []string{"can_crud_permissions_in_project"},
		OrgPermissions:       []string{"can_crud_permissions_in_org"},
		Email:                os.Getenv("MEMBER_EMAIL"),
		OrganizationId:       os.Getenv("SYS11IAM_ORGANIZATION_ID"),
		OrganizationName:     os.Getenv("SYS11IAM_ORGANIZATION_NAME"),
		ServiceAccountSecret: os.Getenv("SYS11IAM_SERVICEACCOUNT_SECRET"),
		IAM_URL:              os.Getenv("SYS11IAM_IAM_URL"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - User membership
			{
				Config: testAccOrganizationProjectMembershipResourceConfig(t, params),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("project_name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("membership").AtMapKey("user_membership").AtMapKey("user").AtMapKey("email"),
						knownvalue.StringExact(params.Email),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "sys11iam_organization_project_membership.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sys11iam_organization_project_membership.test"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "sys11iam_organization_project_membership.test")
					}
					return fmt.Sprintf("%s,%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["project_id"], rs.Primary.Attributes["id"]), nil
				},
			},
			// Update testing
			{
				PreConfig: func() {
					params.ProjectPermissions = []string{"can_crud_permissions_in_project", "can_become_administrator_in_project"}
				},
				Config: testAccOrganizationProjectMembershipResourceConfig(t, params),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("membership").AtMapKey("user_membership").AtMapKey("user").AtMapKey("email"),
						knownvalue.StringExact(params.Email),
					),
				},
			},
			{
				Config: PreventUserMembershipTeardown(t, params),
			},
		},
	})
}

func TestAccOrganizationProjectMembershipResourceServiceAccount(t *testing.T) {
	params := &testAccCheckOrganizationProjectMembershipResourceParams{
		MemberType:           "service_account",
		ProjectPermissions:   []string{"can_crud_permissions_in_project"},
		OrganizationId:       os.Getenv("SYS11IAM_ORGANIZATION_ID"),
		OrganizationName:     os.Getenv("SYS11IAM_ORGANIZATION_NAME"),
		ServiceAccountSecret: os.Getenv("SYS11IAM_SERVICEACCOUNT_SECRET"),
		IAM_URL:              os.Getenv("SYS11IAM_IAM_URL"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing - Service Account membership
			{
				Config: testAccOrganizationProjectMembershipResourceServiceAccountConfig(t, params),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("membership").AtMapKey("service_account_membership").AtMapKey("service_account").AtMapKey("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "sys11iam_organization_project_membership.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sys11iam_organization_project_membership.test"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "sys11iam_organization_project_membership.test")
					}
					return fmt.Sprintf("%s,%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["project_id"], rs.Primary.Attributes["id"]), nil
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

type testAccCheckOrganizationProjectMembershipResourceParams struct {
	MemberType         string
	MemberId           string
	ProjectPermissions []string
	OrgPermissions     []string
	Email              string

	OrganizationId   string
	OrganizationName string

	ServiceAccountSecret string
	IAM_URL              string
}

func testAccOrganizationProjectMembershipResourceConfig(t *testing.T, params *testAccCheckOrganizationProjectMembershipResourceParams) string {
	t.Helper()

	var result strings.Builder

	err := mustParseTemplate("organization project user membership test template", `
provider "sys11iam" {
  serviceaccount_secret = "{{ .ServiceAccountSecret }}"

  iam_url = "{{ .IAM_URL }}"
}

data "sys11iam_organization" "test_org" {
  id = "{{ .OrganizationId }}"
  name = "{{ .OrganizationName }}"
}

resource "sys11iam_organization_project" "test_project" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-user-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

import {
  to = sys11iam_organization_membership.test_user_membership[0]
  id = "{{ .OrganizationId }},{{ .MemberId }}"
}
resource "sys11iam_organization_membership" "test_user_membership" {
  count = data.sys11iam_organization.test_org.is_active ? 1 : 0
  id = "{{ .MemberId }}"
  organization_id = data.sys11iam_organization.test_org.id

  membership = {
		user_membership = {
			permissions = [{{ range $i, $perm := .OrgPermissions }}{{ if $i }}, {{ end }}"{{ $perm }}"{{ end }}]
			affiliation = "member"
			membership_type = "{{ .MemberType }}"
			email = "{{ .Email }}"
		}
	}
}

resource "sys11iam_organization_project_membership" "test" {
    organization_id = data.sys11iam_organization.test_org.id
    project_id      = sys11iam_organization_project.test_project.id
	project_name      = sys11iam_organization_project.test_project.name
    id              = sys11iam_organization_membership.test_user_membership[0].id #member id

	membership = {
		user_membership = {
			permissions = [{{ range $i, $perm := .ProjectPermissions }}{{ if $i }}, {{ end }}"{{ $perm }}"{{ end }}]
			membership_type = "{{ .MemberType }}"
			user = {
				email = "{{ .Email }}"
			}
		}
	}
}
`).Execute(&result, params)

	if err != nil {
		t.Fatalf("Error executing template: %s", err)
	}
	return result.String()
}

func testAccOrganizationProjectMembershipResourceServiceAccountConfig(t *testing.T, params *testAccCheckOrganizationProjectMembershipResourceParams) string {
	t.Helper()

	var result strings.Builder

	err := mustParseTemplate("organization project service account membership test template", `
provider "sys11iam" {
  serviceaccount_secret = "{{ .ServiceAccountSecret }}"

  iam_url = "{{ .IAM_URL }}"
}

data "sys11iam_organization" "test_org" {
  id = "{{ .OrganizationId }}"
  name = "{{ .OrganizationName }}"
}

resource "sys11iam_organization_serviceaccount" "test_serviceaccount" {
  count = data.sys11iam_organization.test_org.is_active ? 1 : 0
  name = "terraform-acceptance-test-serviceaccount"
  description = "test service account"
  organization_id = data.sys11iam_organization.test_org.id
}

resource "sys11iam_organization_project" "test_project" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-sa-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

resource "sys11iam_organization_membership" "test_service_account_membership" {
  count = data.sys11iam_organization.test_org.is_active ? 1 : 0
  id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
  organization_id = data.sys11iam_organization.test_org.id
  membership = {
		service_account_membership = {
			permissions = [{{ range $i, $perm := .OrgPermissions }}{{ if $i }}, {{ end }}"{{ $perm }}"{{ end }}]
			affiliation = "member"
			membership_type = "{{ .MemberType }}"
			id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
			name = sys11iam_organization_serviceaccount.test_serviceaccount[0].name
		}
	}
}

resource "sys11iam_organization_project_membership" "test" {
  organization_id = data.sys11iam_organization.test_org.id
  project_name      = sys11iam_organization_project.test_project.name
  project_id      = sys11iam_organization_project.test_project.id
  id              = sys11iam_organization_membership.test_service_account_membership[0].id #member id

	membership = {
		service_account_membership = {
			permissions = [{{ range $i, $perm := .ProjectPermissions }}{{ if $i }}, {{ end }}"{{ $perm }}"{{ end }}]
			membership_type = "{{ .MemberType }}"
			service_account = {
				id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
			}
		}
	}
}
`).Execute(&result, params)

	if err != nil {
		t.Fatalf("Error executing template: %s", err)
	}
	return result.String()
}
