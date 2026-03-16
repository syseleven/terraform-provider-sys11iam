package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccOrganizationTeamMembershipResource(t *testing.T) {
	if os.Getenv("MEMBER_ID") == "" && os.Getenv("MEMBER_EMAIL") == "" {
		t.Skip("MEMBER_ID or MEMBER_EMAIL must be set for this acceptance test")
	}

	params := &testAccOrganizationTeamMembershipResourceParams{
		TeamPermissions:         []string{"can_manage_team_in_team"},
		OrganizationPermissions: []string{"can_become_project_administrator_in_org"},
		MemberId:                os.Getenv("MEMBER_ID"),
		Email:                   os.Getenv("MEMBER_EMAIL"),
		OrganizationId:          os.Getenv("SYS11IAM_ORGANIZATION_ID"),
		OrganizationName:        os.Getenv("SYS11IAM_ORGANIZATION_NAME"),
		ServiceAccountSecret:    os.Getenv("SYS11IAM_SERVICEACCOUNT_SECRET"),
		IAM_URL:                 os.Getenv("SYS11IAM_IAM_URL"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationTeamMembershipResourceConfig(t, params),
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
						"sys11iam_organization_team_membership.user_membership_test",
						tfjsonpath.New("org_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_team_membership.user_membership_test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(params.MemberId),
					),
				},
			},
			// Update testing
			{
				PreConfig: func() {
					params.TeamPermissions = []string{"can_become_administrator_in_team", "can_become_project_administrator_in_team"}
				},
				Config: testAccOrganizationTeamMembershipResourceConfig(t, params),
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
						"sys11iam_organization_team_membership.user_membership_test",
						tfjsonpath.New("org_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_team_membership.user_membership_test",
						tfjsonpath.New("id"),
						knownvalue.StringExact(params.MemberId),
					),
				},
			},
			// ImportState testing - user membership
			{
				ResourceName:      "sys11iam_organization_team_membership.user_membership_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sys11iam_organization_team_membership.user_membership_test"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "sys11iam_organization_team_membership.user_membership_test")
					}
					return fmt.Sprintf("%s,%s,%s", rs.Primary.Attributes["org_id"], rs.Primary.Attributes["team_id"], rs.Primary.Attributes["id"]), nil
				},
			},
			// ImportState testing - service account membership
			{
				ResourceName:      "sys11iam_organization_team_membership.service_account_membership_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sys11iam_organization_team_membership.service_account_membership_test"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "sys11iam_organization_team_membership.service_account_membership_test")
					}
					return fmt.Sprintf("%s,%s,%s", rs.Primary.Attributes["org_id"], rs.Primary.Attributes["team_id"], rs.Primary.Attributes["id"]), nil
				},
			},
			{
				Config: PreventUserMembershipTeardown(t, params),
			},
		},
	})

}

type testAccOrganizationTeamMembershipResourceParams struct {
	TeamPermissions []string

	OrganizationId          string
	OrganizationName        string
	OrganizationPermissions []string

	MemberId string
	Email    string

	ServiceAccountSecret string
	IAM_URL              string
}

func testAccOrganizationTeamMembershipResourceConfig(t *testing.T, params *testAccOrganizationTeamMembershipResourceParams) string {
	t.Helper()

	var result strings.Builder

	err := mustParseTemplate("organization_team_resource_test", `
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
  org_id = data.sys11iam_organization.test_org.id
}

resource "sys11iam_organization_team" "test_team" {
  name        = "team-team"
  org_id = data.sys11iam_organization.test_org.id
  description = "Test team for acceptance testing"
  organization_permissions = ["{{ range $index, $perm := .OrganizationPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
  tags = ["test", "acceptance-testing"]

  projects = []
}

resource "sys11iam_organization_team_membership" "user_membership_test" {
  team_id = sys11iam_organization_team.test_team.id
  id = "{{ .MemberId }}"
  org_id = data.sys11iam_organization.test_org.id
  membership_type = "user"

  membership = {
	user_team_membership = {
		user = {
			email = "{{ .Email }}"
		}
		team_permissions = ["{{ range $index, $perm := .TeamPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
	}
  }
}

resource "sys11iam_organization_team_membership" "service_account_membership_test" {
  team_id = sys11iam_organization_team.test_team.id
  id = sys11iam_organization_serviceaccount.test_serviceaccount[0].id
  org_id = data.sys11iam_organization.test_org.id
  membership_type = "service_account"

  membership = {
	service_account_team_membership = {
		team_permissions = ["{{ range $index, $perm := .TeamPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
	}
  }
}
`).Execute(&result, params)

	if err != nil {
		t.Fatalf("Error parsing template: %s", err)
	}

	return result.String()
}
