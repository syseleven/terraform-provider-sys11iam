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

func TestAccOrganizationProjectResource(t *testing.T) {
	projectName := fmt.Sprintf("terraform-test-project-%d", time.Now().UnixNano())

	params := &testAccOrganizationProjectResourceParams{
		ProjectName:          projectName,
		ProjectDescription:   "Test project for acceptance testing",
		Tags:                 []string{"test", "acceptance-testing"},
		OrganizationId:       os.Getenv("SYS11IAM_ORGANIZATION_ID"),
		OrganizationName:     os.Getenv("SYS11IAM_ORGANIZATION_NAME"),
		ServiceAccountSecret: os.Getenv("SYS11IAM_SERVICEACCOUNT_SECRET"),
		IAM_URL:              os.Getenv("SYS11IAM_IAM_URL"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationProjectResourceConfig(t, params),
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
						"sys11iam_organization_project.test",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(params.ProjectName),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(params.ProjectDescription),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test",
						tfjsonpath.New("created_at"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test",
						tfjsonpath.New("updated_at"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test",
						tfjsonpath.New("tags"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("test"),
							knownvalue.StringExact("acceptance-testing"),
						}),
					),
				},
			},
			// Update testing
			{
				PreConfig: func() {
					params.ProjectName = fmt.Sprintf("terraform-updated-project-%d", time.Now().UnixNano())
					params.ProjectDescription = "Updated test project description"
					params.Tags = []string{"test", "acceptance-testing", "updated"}
				},
				Config: testAccOrganizationProjectResourceConfig(t, params),
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
						"sys11iam_organization_project.test",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
				},
			},
			// ImportState testing
			{
				ResourceName: "sys11iam_organization_project.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sys11iam_organization_project.test"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "sys11iam_organization_project.test")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func TestAccOrganizationProjectResourceMinimal(t *testing.T) {
	projectName := fmt.Sprintf("terraform-minimal-project-%d", time.Now().UnixNano())

	params := &testAccOrganizationProjectResourceMinimalParams{
		ProjectName:          projectName,
		OrganizationId:       os.Getenv("SYS11IAM_ORGANIZATION_ID"),
		OrganizationName:     os.Getenv("SYS11IAM_ORGANIZATION_NAME"),
		ServiceAccountSecret: os.Getenv("SYS11IAM_SERVICEACCOUNT_SECRET"),
		IAM_URL:              os.Getenv("SYS11IAM_IAM_URL"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with minimal configuration
			{
				Config: testAccOrganizationProjectResourceMinimalConfig(t, params),
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
						"sys11iam_organization_project.test_minimal",
						tfjsonpath.New("organization_id"),
						knownvalue.StringExact(params.OrganizationId),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test_minimal",
						tfjsonpath.New("name"),
						knownvalue.StringExact(params.ProjectName),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test_minimal",
						tfjsonpath.New("description"),
						knownvalue.StringExact(""), // Default empty description
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test_minimal",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project.test_minimal",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName: "sys11iam_organization_project.test_minimal",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["sys11iam_organization_project.test_minimal"]
					if !ok {
						return "", fmt.Errorf("not found: %s", "sys11iam_organization_project.test_minimal")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

type testAccOrganizationProjectResourceParams struct {
	ProjectName        string
	ProjectDescription string
	Tags               []string

	OrganizationId   string
	OrganizationName string

	ServiceAccountSecret string
	IAM_URL              string
}

type testAccOrganizationProjectResourceMinimalParams struct {
	ProjectName string

	OrganizationId   string
	OrganizationName string

	ServiceAccountSecret string
	IAM_URL              string
}

func testAccOrganizationProjectResourceConfig(t *testing.T, params *testAccOrganizationProjectResourceParams) string {
	t.Helper()

	var result strings.Builder

	err := mustParseTemplate("organization_project_resource_test", `
provider "sys11iam" {
  serviceaccount_secret = "{{ .ServiceAccountSecret }}"
  iam_url = "{{ .IAM_URL }}"
}

data "sys11iam_organization" "test_org" {
  id = "{{ .OrganizationId }}"
  name = "{{ .OrganizationName }}"
}

resource "sys11iam_organization_project" "test" {
  organization_id = data.sys11iam_organization.test_org.id
  name            = "{{ .ProjectName }}"
  description     = "{{ .ProjectDescription }}"
  tags            = [{{ range $index, $tag := .Tags }}{{ if $index }}, {{ end }}"{{ $tag }}"{{ end }}]
}
`).Execute(&result, params)

	if err != nil {
		t.Fatalf("Error executing template: %s", err)
	}

	return result.String()
}

func testAccOrganizationProjectResourceMinimalConfig(t *testing.T, params *testAccOrganizationProjectResourceMinimalParams) string {
	t.Helper()

	var result strings.Builder

	err := mustParseTemplate("organization_project_resource_minimal_test", `
provider "sys11iam" {
  serviceaccount_secret = "{{ .ServiceAccountSecret }}"
  iam_url = "{{ .IAM_URL }}"
}

data "sys11iam_organization" "test_org" {
  id = "{{ .OrganizationId }}"
  name = "{{ .OrganizationName }}"
}

resource "sys11iam_organization_project" "test_minimal" {
  organization_id = data.sys11iam_organization.test_org.id
  name            = "{{ .ProjectName }}"
}
`).Execute(&result, params)

	if err != nil {
		t.Fatalf("Error executing template: %s", err)
	}

	return result.String()
}
