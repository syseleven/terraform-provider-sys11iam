package provider

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/iam"
)

func TestAccOrganizationProjectMembershipResource(t *testing.T) {
	server := orgPrjMembershpTestServer("user")
	if server.Listener != nil {
		server.Listener.Close()
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		panic(fmt.Sprintf("Failed to create listener: %v", err))
	}
	server.Listener = listener
	server.Start()

	defer server.Close()

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
	server := orgPrjMembershpTestServer("service_account")
	if server.Listener != nil {
		server.Listener.Close()
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		panic(fmt.Sprintf("Failed to create listener: %v", err))
	}
	server.Listener = listener
	server.Start()

	defer server.Close()

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
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"sys11iam_organization_project_membership.test",
						tfjsonpath.New("service_account_membership").AtMapKey("service_account").AtMapKey("id"),
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
					return fmt.Sprintf("%s:%s:%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["project_id"], rs.Primary.Attributes["member_id"]), nil
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func orgPrjMembershpTestServer(membershipType string) *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		// Define response structs based on the endpoint

		if r.URL.Path == "/v2/orgs/org-123" {
			// Response for organization data source
			response := `
                    {   
                        "id": "org-123",
                        "name": "terraform-test-org",
                        "description": "A test organization for Terraform acceptance tests",
                        "created_at": "2023-01-01T00:00:00Z",
                        "updated_at": "2023-01-01T00:00:00Z",
                        "tags": ["test", "terraform"],
                        "is_active": true
                    }`
			w.Write([]byte(response))
		} else if r.URL.Path == "/v2/orgs" {
			response := `[
                {
                    "id": "org-123",
                    "name": "terraform-test-org",
                    "description": "A test organization for Terraform acceptance tests",
                    "created_at": "2023-01-01T00:00:00Z",
                    "tags": ["test", "terraform"],
                    "updated_at": "2023-01-01T00:00:00Z",
                    "is_active": true
                }
            ]`
			w.Write([]byte(response))
		} else if r.URL.Path == "/v2/orgs/org-123/memberships" {
			// Response for organization memberships
			response := `{
                    "id": "member-123",
                    "user": {
                        "email": "test@example.com"
                    },
                    "permissions": ["read", "write"],
        }`
			w.Write([]byte(response))
		} else if r.URL.Path == "/v2/orgs/org-123/invitations" {
			// Response for organization invitations
			response := `
                []`
			w.Write([]byte(response))
		} else if r.URL.Path == "/v2/orgs/org-123/memberships/member-123" {
			// Response for specific organization membership
			var response string
			switch membershipType {
			case "user":
				response = `{
                        "id": "member-123",
                        "affiliation": "user",
                        "user": {
                            "email": "test@example.com"
                        },
                        "permissions": ["read", "write"]
                    }`
			case "service_account":
				response = `{
                        "id": "member-123",
                        "affiliation": "service_account",
                        "service_account": {
                            "id": "sa-123",
                            "name": "test-service-account"
                        },
                        "permissions": ["admin"]
                }`
			}
			w.Write([]byte(response))
		} else if strings.HasPrefix(r.URL.Path, "/v2/orgs/org-123/projects/78e5709a12334684ad03f8b1233f9123/memberships/member-123/permissions") {
			// Response for project membership operations
			var response iam.IAMProjectMembership
			switch membershipType {
			case "user":
				// {
                //         "project_id": "78e5709a12334684ad03f8b1233f9123",
                //         "project_name": "Test Project",
                //         "permissions": ["read", "write", "admin"],
                //         "membership_type": "user",
                //         "user": {
                //             "id": "member-123",
                //             "name": "Test User",
                //             "email": "test@example.com",
                //             "description": "A test user for Terraform acceptance tests",
                //             "created_at": "2023-01-01T00:00:00Z",
                //             "updated_at": "2023-01-01T00:00:00Z
                //         },
                //         "project": {
                //             "id": "78e5709a12334684ad03f8b1233f9123",
                //             "name": "Test Project",
                //             "tags": ["test", "terraform"],
                //             "description": "A test project for Terraform acceptance tests",
                //             "created_at": "2023-01-01T00:00:00Z",
                //             "updated_at": "2023-01-01T00:00:00Z"
                //         }
                // }

                response = iam.IAMProjectMembership{
                    ProjectId:             "member-123",
                    ProjectName:    "Test Project",
                    Permissions:    []string{"read", "write", "admin"},
                    MembershipType: membershipType,
                    User: iam.IAMOrganisationUser{
                        ID:          "member-123",
                        Name:        "Test User",
                        Email:      "test@example.com",
                        Description: "A test user for Terraform acceptance tests",
                        CreatedAt:  "2023-01-01T00:00:00Z",
                        UpdatedAt:  "2023-01-01T00:00:00Z",
                    },
                    Project: iam.IAMProject{
                        ID:          "78e5709a12334684ad03f8b1233f9123",
                        Name:        "Test Project",
                        Tags:        []string{"test", "terraform"},
                        Description: "A test project for Terraform acceptance tests",
                        CreatedAt:  "2023-01-01T00:00:00Z",
                        UpdatedAt:  "2023-01-01T00:00:00Z",
                    },
                }
			case "service_account":
				// {
                //         "id": "member-123",
                //         "project_id": "78e5709a12334684ad03f8b1233f9123",
                //         "project_name": "Test Project",
                //         "permissions": ["admin"],
                //         "membership_type": "service_account",
                //         "service_account": {
                //             "id": "sa-123",
                //             "name": "test-service-account",
                //             "description": "A test service account for Terraform acceptance tests",
                //             "created_at": "2023-01-01T00:00:00Z",
                //             "updated_at": "2023-01-01T00:00:00Z"
                //         },
                //         "project": {
                //             "id": "78e5709a12334684ad03f8b1233f9123",
                //             "name": "Test Project",
                //             "tags": ["test", "terraform"],
                //             "description": "A test project for Terraform acceptance tests",
                //             "created_at": "2023-01-01T00:00:00Z",
                //             "updated_at": "2023-01-01T00:00:00Z"
                //         }
                // }

                response = iam.IAMProjectMembership{
                    ProjectId:      "78e5709a12334684ad03f8b1233f9123",
                    ProjectName:    "Test Project",
                    Permissions:    []string{"admin"},
                    MembershipType: membershipType,
                    ServiceAccount: iam.IAMOrganisationServiceAccount{
                        ID:          "sa-123",
                        Name:        "test-service-account",
                        Description: "A test service account for Terraform acceptance tests",
                        CreatedAt:  "2023-01-01T00:00:00Z",
                        UpdatedAt:  "2023-01-01T00:00:00Z",
                    },
                    Project: iam.IAMProject{
                        ID:          "78e5709a12334684ad03f8b1233f9123",
                        Name:        "Test Project",
                        Tags:        []string{"test", "terraform"},
                        Description: "A test project for Terraform acceptance tests",
                        CreatedAt:  "2023-01-01T00:00:00Z",
                        UpdatedAt:  "2023-01-01T00:00:00Z",
                    },
                }
			}

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            if err := json.NewEncoder(w).Encode(response); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
		}
	}))
}

func testAccOrganizationProjectMembershipResourceConfig(memberType, email, permissions string) string {
	return fmt.Sprintf(`
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_b1dfa4c9-5344-1234-1234-d96a71234567"

  iam_url = "http://127.0.0.1:8000"
}

data "sys11iam_organization" "test_org" {
  id = "org-123"
  name = "terraform-test-org"
}

resource "sys11iam_organization_project_membership" "test" {
    organization_id = data.sys11iam_organization.test_org.id
    project_id      = "78e5709a12334684ad03f8b1233f9123"
    member_id       = "member-123"
  
    user_membership = {
        permissions = [%[3]s]
        membership_type = "%[1]s"
        user = {
            email = "%[2]s"
        }
    }
}
`, memberType, email, formatPermissions(permissions))
}

func testAccOrganizationProjectMembershipResourceServiceAccountConfig(serviceAccountId, permissions string) string {
	return fmt.Sprintf(`
provider "sys11iam" {
  serviceaccount_secret = "s11_orgsa_b1dfa4c9-5344-1234-1234-d96a71234567"

  iam_url = "http://127.0.0.1:8000"
}

data "sys11iam_organization" "test_org" {
  id = "org-123"
  name = "terraform-test-org"
}

resource "sys11iam_organization_project_membership" "test" {
  organization_id = data.sys11iam_organization.test_org.id
  project_id      = "78e5709a12334684ad03f8b1233f9123"
  member_id       = "member-123"

    service_account_membership = {
        permissions = [%[2]s]
        membership_type = "service_account"
        service_account = {
            id = "%[1]s"
        }
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
