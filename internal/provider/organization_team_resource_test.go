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

func TestAccOrganizationTeamResource(t *testing.T) {
	params := &testAccOrganizationTeamResourceParams{
		TeamName:                "test-team",
		OrganizationId:          os.Getenv("SYS11IAM_ORGANIZATION_ID"),
		OrganizationName:        os.Getenv("SYS11IAM_ORGANIZATION_NAME"),
		OrganizationPermissions: []string{"can_become_project_administrator_in_org"},
		ProjectPermissions:      []string{"can_become_administrator_in_project"},
		ServiceAccountSecret:    os.Getenv("SYS11IAM_SERVICEACCOUNT_SECRET"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationTeamResourceConfig(t, params),
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
						"sys11iam_organization_team.test",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
				},
			},
			// Update testing
			{
				Config: testAccOrganizationTeamResourceConfig(t, params),
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
						"sys11iam_organization_team.test",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
				},
			},
			// ImportState testing
			{
				ResourceName: "sys11iam_organization_team.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sys11iam_organization_team.test"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "sys11iam_organization_team.test")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

type testAccOrganizationTeamResourceParams struct {
	TeamName string

	OrganizationId   string
	OrganizationName string

	OrganizationPermissions []string
	ProjectPermissions      []string

	ServiceAccountSecret string
}

func testAccOrganizationTeamResourceConfig(t *testing.T, params *testAccOrganizationTeamResourceParams) string {
	t.Helper()

	var result strings.Builder

	err := mustParseTemplate("organization_team_resource_test", `
provider "sys11iam" {
  serviceaccount_secret = "{{ .ServiceAccountSecret }}"

  iam_url = "https://iam.apis.syseleven.de"
}

data "sys11iam_organization" "test_org" {
  id = "{{ .OrganizationId }}"
  name = "{{ .OrganizationName }}"
}

resource "sys11iam_organization_project" "test_project_1" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-test-project-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

resource "sys11iam_organization_project" "test_project_2" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-test-project-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

resource "sys11iam_organization_project" "test_project_3" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-test-project-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

resource "sys11iam_organization_project" "test_project_4" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-test-project-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

resource "sys11iam_organization_project" "test_project_5" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-test-project-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

resource "sys11iam_organization_project" "test_project_6" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-test-project-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}
  
resource "sys11iam_organization_project" "test_project_7" {
  organization_id = data.sys11iam_organization.test_org.id
  name = "terraform-test-project-`+fmt.Sprintf("%d", time.Now().UnixNano())+`"
}

resource "sys11iam_organization_team" "test" {
  name        = "{{ .TeamName }}"
  organization_id = data.sys11iam_organization.test_org.id
  description = "Test team for acceptance testing"
  organization_permissions = ["{{ range $index, $perm := .OrganizationPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
  tags = ["test", "acceptance-testing"]

  projects = [
    {
      id = sys11iam_organization_project.test_project_1.id
      project_permissions = ["{{ range $index, $perm := .ProjectPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
    },
    {
      id = sys11iam_organization_project.test_project_2.id
      project_permissions = ["{{ range $index, $perm := .ProjectPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
    },
    {
      id = sys11iam_organization_project.test_project_3.id
      project_permissions = ["{{ range $index, $perm := .ProjectPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
    },
    {
      id = sys11iam_organization_project.test_project_4.id
      project_permissions = ["{{ range $index, $perm := .ProjectPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
    },
    {
      id = sys11iam_organization_project.test_project_5.id
      project_permissions = ["{{ range $index, $perm := .ProjectPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
    },
    {
      id = sys11iam_organization_project.test_project_6.id
      project_permissions = ["{{ range $index, $perm := .ProjectPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
    },
    {
      id = sys11iam_organization_project.test_project_7.id
      project_permissions = ["{{ range $index, $perm := .ProjectPermissions }}{{ if $index }}, {{ end }}{{ $perm }}{{ end }}"]
    }
  ]
}
`).Execute(&result, params)

	if err != nil {
		t.Fatalf("Error executing template: %s", err)
	}

	return result.String()
}
