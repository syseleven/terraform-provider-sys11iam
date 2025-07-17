package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccOrganizationProjectMembershipResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing - User membership
            {
                Config: testAccOrganizationProjectMembershipResourceConfig("user", "test@example.com", "read,write"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("organization_id"),
                        knownvalue.StringExact("org-123"),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("project_id"),
                        knownvalue.StringExact("78e5709a12334684ad03f8b1233f9123"),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("member_id"),
                        knownvalue.NotNull(),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("user_membership").AtMapKey("user").AtMapKey("email"),
                        knownvalue.StringExact("test@example.com"),
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
                    return fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["project_id"], rs.Primary.Attributes["member_id"]), nil
                },
            },
            // Update and Read testing - Change permissions
            {
                Config: testAccOrganizationProjectMembershipResourceConfig("user", "test@example.com", "read,write,admin"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("organization_id"),
                        knownvalue.StringExact("org-123"),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("project_id"),
                        knownvalue.StringExact("78e5709a12334684ad03f8b1233f9123"),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("user_membership").AtMapKey("permissions"),
                        knownvalue.ListExact([]knownvalue.Check{
                            knownvalue.StringExact("read"),
                            knownvalue.StringExact("write"),
                            knownvalue.StringExact("admin"),
                        }),
                    ),
                },
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}

func TestAccOrganizationProjectMembershipResourceServiceAccount(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing - Service Account membership
            {
                Config: testAccOrganizationProjectMembershipResourceServiceAccountConfig("sa-123", "admin"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("organization_id"),
                        knownvalue.StringExact("org-123"),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("project_id"),
                        knownvalue.StringExact("78e5709a12334684ad03f8b1233f9123"),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("member_id"),
                        knownvalue.StringExact("sa-123"),
                    ),
                    statecheck.ExpectKnownValue(
                        "sys11iam_organization_project_membership.test",
                        tfjsonpath.New("service_account_membership").AtMapKey("service_account").AtMapKey("id"),
                        knownvalue.StringExact("sa-123"),
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
                    return fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["project_id"], rs.Primary.Attributes["member_id"]), nil
                },
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}


// TODO: This requires setting up a docker container with wiremock or testcontainers to mock the IAM API

func testAccOrganizationProjectMembershipResourceConfig(memberType, email, permissions string) string {
    return fmt.Sprintf(`
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_b1dfa4c9-5344-1234-1234-d96a71234567"

  iam_url = "https://127.0.1:8000"
}

data "sys11iam_organization" "test_org" {
  id = "org-123"
  name = "terraform-test-org"
}

resource "sys11iam_organization_project_membership" "test" {
  organization_id = data.sys11iam_organization.test_org.id
  project_id      = "78e5709a12334684ad03f8b1233f9123"
  
  user_membership = {
    permissions = [%[3]s]
  }
}
`, memberType, email, formatPermissions(permissions))
}

func testAccOrganizationProjectMembershipResourceServiceAccountConfig(serviceAccountId, permissions string) string {
    return fmt.Sprintf(`
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_b1dfa4c9-5344-1234-1234-d96a71234567"

  iam_url = "https://127.0.1:8000"
}

data "sys11iam_organization" "test_org" {
  id = "org-123"
  name = "terraform-test-org"
}

resource "sys11iam_organization_project_membership" "test" {
  organization_id = data.sys11iam_organization.test_org.id
  project_id      = "78e5709a12334684ad03f8b1233f9123"
  
  service_account_membership = {
    permissions = [%[2]s]
  }
}
`, serviceAccountId, formatPermissions(permissions))
}

// Helper function to format permissions string for HCL
func formatPermissions(permissions string) string {
    if permissions == "" {
        return ""
    }
    
    var formatted []string
    for _, perm := range strings.Split(permissions, ",") {
        formatted = append(formatted, fmt.Sprintf(`"%s"`, strings.TrimSpace(perm)))
    }
    return strings.Join(formatted, ", ")
}
