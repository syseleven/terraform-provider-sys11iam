package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/syseleven/terraform-provider-sys11iam/internal/errors"
)

// Endpoint constants for the IAM API v3.
// Format specifiers (%s) are positional: the parameter order is documented inline.

const IAMOrganizationsEndpoint string = "/v3/orgs"
const IAMOrganizationEndpoint string = "/v3/orgs/%s" // org_id

const IAMProjectsEndpoint string = "/v3/orgs/%s/projects"   // org_id
const IAMProjectEndpoint string = "/v3/orgs/%s/projects/%s" // org_id, project_id

const IAMOrganizationMembershipsEndpoint string = "/v3/orgs/%s/memberships"                          // org_id
const IAMOrganizationMembershipEndpoint string = "/v3/orgs/%s/memberships/%s"                        // org_id, membership_id
const IAMOrganizationMembershipPermissionsEndpoint string = "/v3/orgs/%s/memberships/%s/permissions" // org_id, membership_id

const IAMOrganizationInvitationsEndpoint string = "/v3/orgs/%s/invitations"   // org_id
const IAMOrganizationInvitationEndpoint string = "/v3/orgs/%s/invitations/%s" // org_id, invitation_id

const IAMProjectMembershipsEndpoint string = "/v3/orgs/%s/projects/%s/memberships"                          // org_id, project_id
const IAMProjectMembershipEndpoint string = "/v3/orgs/%s/projects/%s/memberships/%s"                        // org_id, project_id, membership_id
const IAMProjectMembershipPermissionsEndpoint string = "/v3/orgs/%s/projects/%s/memberships/%s/permissions" // org_id, project_id, membership_id

const IAMOrganizationServiceaccountsEndpoint string = "/v3/orgs/%s/service_accounts"                          // org_id
const IAMOrganizationServiceaccountEndpoint string = "/v3/orgs/%s/service_accounts/%s"                        // org_id, service_account_id
const IAMOrganizationServiceaccountPermissionsEndpoint string = "/v3/orgs/%s/service_accounts/%s/permissions" // org_id, service_account_id

const IAMOrganizationTeamsEndpoint string = "/v3/orgs/%s/teams"                          // org_id
const IAMOrganizationTeamEndpoint string = "/v3/orgs/%s/teams/%s"                        // org_id, team_id
const IAMOrganizationTeamPermissionsEndpoint string = "/v3/orgs/%s/teams/%s/permissions" // org_id, team_id

const IAMOrganizationContactsEndpoint string = "/v3/orgs/%s/contacts"   // org_id
const IAMOrganizationContactEndpoint string = "/v3/orgs/%s/contacts/%s" // org_id, contact_id

const IAMProjectTeamPermissionsEndpoint string = "/v3/orgs/%s/projects/%s/teams/%s/permissions" // org_id, project_id, team_id

const IAMTeamProjectsEndpoint string = "/v3/orgs/%s/teams/%s/projects" // org_id, team_id

const IAMOrganizationTeamMembershipsEndpoint string = "/v3/orgs/%s/teams/%s/memberships"   // org_id, team_id
const IAMOrganizationTeamMembershipEndpoint string = "/v3/orgs/%s/teams/%s/memberships/%s" // org_id, team_id, membership_id

const IAMProjectTeamMembershipsEndpoint string = "/v3/orgs/%s/projects/%s/teams/%s/memberships"                          // org_id, project_id, team_id
const IAMProjectTeamMembershipEndpoint string = "/v3/orgs/%s/projects/%s/teams/%s/memberships/%s"                        // org_id, project_id, team_id, membership_id
const IAMProjectTeamMembershipPermissionsEndpoint string = "/v3/orgs/%s/projects/%s/teams/%s/memberships/%s/permissions" // org_id, project_id, team_id, membership_id

const IAMProjectS3UsersEndpoint string = "/v3/orgs/%s/projects/%s/s3-users"                         // org_id, project_id
const IAMProjectS3UserEndpoint string = "/v3/orgs/%s/projects/%s/s3-users/%s"                       // org_id, project_id, s3_user_id
const IAMProjectS3UserKeysEndpoint string = "/v3/orgs/%s/projects/%s/s3-users/%s/ec2-credentials"   // org_id, project_id, s3_user_id
const IAMProjectS3UserKeyEndpoint string = "/v3/orgs/%s/projects/%s/s3-users/%s/ec2-credentials/%s" // org_id, project_id, s3_user_id, access_key

const IAMOrganizationTeamMembershipPermissionsEndpoint string = "/v3/orgs/%s/teams/%s/permissions/%s" // org_id, team_id, permission_name

// Glue v3 permission endpoints — per-principal permission management
const IAMUserOrgPermissionsEndpoint string = "/v3/orgs/%s/users/%s/permissions"                                      // org_id, user_id
const IAMServiceAccountOrgPermissionsEndpoint string = "/v3/orgs/%s/service_accounts/%s/permissions"                 // org_id, service_account_id
const IAMUserOrgAffiliationEndpoint string = "/v3/orgs/%s/users/%s/affiliation"                                      // org_id, user_id
const IAMServiceAccountOrgAffiliationEndpoint string = "/v3/orgs/%s/service_accounts/%s/affiliation"                 // org_id, service_account_id
const IAMUserProjectPermissionsEndpoint string = "/v3/orgs/%s/projects/%s/users/%s/permissions"                      // org_id, project_id, user_id
const IAMServiceAccountProjectPermissionsEndpoint string = "/v3/orgs/%s/projects/%s/service_accounts/%s/permissions" // org_id, project_id, service_account_id
const IAMUserTeamPermissionsEndpoint string = "/v3/orgs/%s/teams/%s/users/%s/permissions"                            // org_id, team_id, user_id
const IAMServiceAccountTeamPermissionsEndpoint string = "/v3/orgs/%s/teams/%s/service_accounts/%s/permissions"       // org_id, team_id, service_account_id
const IAMUserTeamMemberEndpoint string = "/v3/orgs/%s/teams/%s/users/%s"                                            // org_id, team_id, user_id
const IAMServiceAccountTeamMemberEndpoint string = "/v3/orgs/%s/teams/%s/service_accounts/%s"                       // org_id, team_id, service_account_id

type IAMOrganization struct {
	// org id
	ID string `json:"id" validate:"required"`

	// org name
	Name string `json:"name" validate:"required"`

	// org description
	Description string `json:"description"`

	// org tags
	Tags []string `json:"tags" validate:"required"`

	// org created_at
	CreatedAt string `json:"created_at" validate:"required"`

	// org is_active
	IsActive bool `json:"is_active"`

	// org updated_at
	UpdatedAt string `json:"updated_at" validate:"required"`

	// org company_info
	CompanyInfo IAMOrganizationCompanyInfo `json:"company_info"`
}

type IAMOrganizationCompanyInfo struct {
	// company street
	Street string `json:"street"`

	// company street number
	StreetNumber string `json:"street_number"`

	// company zip code
	ZipCode string `json:"zip_code"`

	// company city
	City string `json:"city"`

	// company country
	Country string `json:"country"`

	// company vat id
	VatID string `json:"vat_id"`

	// company preferred billing method
	PreferredBillingMethod string `json:"preferred_billing_method"`

	// company phone
	Phone string `json:"phone"`

	// company accepted tos
	AcceptedTos bool `json:"accepted_tos"`

	// company name
	CompanyName string `json:"company_name"`
}

type IAMProject struct {
	// project id
	ID string `json:"id"`

	// project name
	Name string `json:"name"`

	// project description
	Description string `json:"description"`

	// project tags
	Tags []string `json:"tags"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Status    string `json:"status"`
}

type IAMTeamProjectWithPermissions struct {
	// project id
	ID string `json:"id"`

	// project name
	Name string `json:"name"`

	// project permissions
	Permissions []string `json:"permissions"`

	// project description
	Description string `json:"description"`

	// project tags
	Tags []string `json:"tags"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Status    string `json:"status"`
}

type IAMOrganizationMembership struct {
	ID             string `json:"id"`
	Affiliation    string `json:"affiliation"`
	MembershipType string `json:"membership_type"`
	// OrganizationMembership permissions
	Permissions          []string                      `json:"editable_permissions,omitempty"`
	ImmutablePermissions []string                      `json:"non_editable_permissions,omitempty"`
	ServiceAccount       IAMOrganisationServiceAccount `json:"service_account"`
	Organisation         IAMOrganization               `json:"organization"`
	User                 IAMOrganisationUser           `json:"user"`
}

type IAMOrganizationTeam struct {
	// team id
	ID string `json:"id"`

	// team name
	Name string `json:"name"`

	// team description
	Description string `json:"description"`

	// team tags
	Tags []string `json:"tags"`
}

type IAMOrganizationTeamPermissions struct {
	// team id
	TeamPermissions []string `json:"."`
}

type IAMProjectTeamPermissions struct {
	TeamPermissions []string `json:"team_permissions,omitempty"`

	// org id
	OrganizationId string `json:"org_id,omitempty"`
	// project id
	ProjectId string `json:"project_id,omitempty"`
	// team id
	TeamId string `json:"team_id,omitempty"`

	AddedPermissions   []string `json:"added_permissions,omitempty"`
	RemovedPermissions []string `json:"removed_permissions,omitempty"`
	UpdatedPermissions []string `json:"updated_permissions,omitempty"`
}

type IAMOrganizationTeamMembership struct {
	MembershipType          string                        `json:"membership_type,omitempty"`
	OrganizationPermissions []string                      `json:"org_permissions,omitempty"`
	TeamPermissions         []string                      `json:"team_permissions,omitempty"`
	Projects                map[string]interface{}        `json:"projects"`
	ServiceAccount          IAMOrganisationServiceAccount `json:"service_account"`
	Organisation            IAMOrganization               `json:"organization"`
	User                    IAMOrganisationUser           `json:"user"`
	Team                    IAMOrganizationTeam           `json:"team"`
}

type IAMProjectTeamMembership struct {
	// project id
	ProjectId string `json:"project_id,omitempty"`

	// project name
	ProjectName string `json:"project_name,omitempty"`

	// ProjectMembership permissions
	Permissions []string `json:"permissions,omitempty"`

	MembershipType string                        `json:"membership_type,omitempty"`
	ServiceAccount IAMOrganisationServiceAccount `json:"service_account"`
	User           IAMOrganisationUser           `json:"user"`
	Project        IAMProject                    `json:"project"`
}

type IAMProjectS3User struct {
	// s3user id
	ID string `json:"id"`
	// s3user name
	Name string `json:"name"`
	// s3user description
	Description string `json:"description"`
	// s3user keys
	Keys []IAMProjectS3UserKey `json:"keys"`
}

type IAMProjectS3UserKey struct {
	//s3user access key
	AccessKey string `json:"access_key"`
	//s3user secret key
	SecretKey string `json:"secret_key"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type IAMOrganizationContact struct {
	// contact id
	ID string `json:"id"`

	// contact first name
	FirstName string `json:"first_name"`

	// contact last name
	LastName string `json:"last_name"`

	// contact email
	Email string `json:"email"`

	// contact phone
	Phone string `json:"phone"`

	// contact notes
	Notes string `json:"notes"`

	// contact roles
	Roles []string `json:"roles"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type IAMOrganisationServiceAccount struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type IAMOrganisationUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Email       string `json:"email"`
}

type IAMOrganizationInvitation struct {
	// OrganizationInvitation id
	ID string `json:"id,omitempty"`

	// OrganizationInvitation email
	Email string `json:"email,omitempty"`

	// OrganizationInvitation email
	ExpirationDate string `json:"expiration_date,omitempty"`

	// org id
	OrganizationId string `json:"org_id,omitempty"`

	// org name
	OrganizationName string `json:"org_name,omitempty"`

	// OrganizationInvitation permissions
	Permissions []string `json:"permissions,omitempty"`

	CreatedAt string `json:"created_at"`
}

type IAMProjectMembership struct {
	// project id
	ProjectId string `json:"project_id,omitempty"`

	// project name
	ProjectName string `json:"project_name,omitempty"`

	// ProjectMembership permissions
	Permissions []string `json:"permissions,omitempty"`

	MembershipType string                        `json:"membership_type,omitempty"`
	ServiceAccount IAMOrganisationServiceAccount `json:"service_account,omitempty"`
	User           IAMOrganisationUser           `json:"user,omitempty"`
	Project        IAMProject                    `json:"project,omitempty"`
}

type IAMOrganizationServiceaccount struct {
	// OrganizationServiceaccount id
	ID string `json:"id"`

	// OrganizationServiceaccount name
	Name string `json:"name"`

	// OrganizationServiceaccount description
	Description string `json:"description"`

	// org id
	OrganizationId string `json:"org_id"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type IAMOrganizationMembershipPermission struct {
	MemberId       string                                             `json:"member_id"`
	OrganizationId string                                             `json:"org_id"`
	ServiceAccount IAMOrganizationMembershipPermissionsServiceAccount `json:"service_account,omitempty"`
	User           IAMOrganizationMembershipPermissionsUser           `json:"user,omitempty"`
}

type IAMOrganizationMembershipPermissionsUser struct {
	Affiliation    string              `json:"affiliation"`
	Permissions    []string            `json:"permissions,omitempty"`
	Id             string              `json:"id"`
	MembershipType string              `json:"membership_type"`
	OrganizationId string              `json:"org_id"`
	User           IAMOrganisationUser `json:"user"`
}

type IAMOrganizationMembershipPermissionsServiceAccount struct {
	Affiliation    string                        `json:"affiliation"`
	Permissions    []string                      `json:"permissions,omitempty"`
	Id             string                        `json:"id"`
	MembershipType string                        `json:"membership_type"`
	OrganizationId string                        `json:"org_id"`
	ServiceAccount IAMOrganisationServiceAccount `json:"service_account"`
}

type IAMOrganizationTeamPermission struct {
	OrganizationId     string   `json:"org_id"`
	TeamId             string   `json:"team_id"`
	AddedPermissions   []string `json:"added_permissions,omitempty"`
	RemovedPermissions []string `json:"removed_permissions,omitempty"`
	UpdatedPermissions []string `json:"updated_permissions,omitempty"`
}

type IAMOrganizationTeamMembershipPermission struct {
	MemberId           string   `json:"member_id"`
	OrganizationId     string   `json:"org_id"`
	TeamId             string   `json:"team_id"`
	MemberType         string   `json:"member_type"`
	AddedPermissions   []string `json:"added_permissions,omitempty"`
	RemovedPermissions []string `json:"removed_permissions,omitempty"`
	UpdatedPermissions []string `json:"updated_permissions,omitempty"`
}

// Glue v3 permission response models

type IAMPermissionEntry struct {
	Permission           string `json:"permission"`
	Active               bool   `json:"active"`
	Direct               bool   `json:"direct"`
	IsDirectlyAssignable bool   `json:"is_directly_assignable"`
}

type IAMAffiliation struct {
	Affiliation string `json:"affiliation"`
}

// FilterActiveDirectPermissions extracts permission names where both active and direct are true.
func FilterActiveDirectPermissions(entries []IAMPermissionEntry) []string {
	var result []string
	for _, e := range entries {
		if e.Active && e.Direct {
			result = append(result, e.Permission)
		}
	}
	return result
}

// IAMOrgMembershipV3 represents a membership entry from the v3 membership list endpoint.
type IAMOrgMembershipV3 struct {
	ID             string               `json:"id"`
	Type           string               `json:"type"` // "user", "service_account", "team"
	DisplayName    string               `json:"display_name"`
	OrgAffiliation *string              `json:"org_affiliation"`
	Permissions    []IAMPermissionEntry `json:"permissions"`
}

// IAMProjectMembershipV3 represents a membership entry from the v3 project membership list endpoint.
type IAMProjectMembershipV3 struct {
	ID          string               `json:"id"`
	Type        string               `json:"type"`
	DisplayName string               `json:"display_name"`
	Permissions []IAMPermissionEntry `json:"permissions"`
}

// IAMTeamMembershipV3 represents a membership entry from the v3 team membership list endpoint.
type IAMTeamMembershipV3 struct {
	ID          string               `json:"id"`
	Type        string               `json:"type"`
	DisplayName string               `json:"display_name"`
	Permissions []IAMPermissionEntry `json:"permissions"`
}

type IAMClient interface {
	GetOrganization(id string) (IAMOrganization, error)
	GetOrganizationByName(name string) (IAMOrganization, error)
	CreateOrganization(org IAMOrganization) (IAMOrganization, error)
	UpdateOrganization(id string, org IAMOrganization) (IAMOrganization, error)
	DeleteOrganization(id string) error

	GetProject(org_id string, project_id string) (IAMProject, error)
	CreateProject(org_id string, name string, description string, tags []string) (IAMProject, error)
	UpdateProject(org_id string, id string, name string, description string, tags []string) (IAMProject, error)
	DeleteProject(org_id string, id string) error

	GetOrganizationMembershipByEmail(ctx context.Context, org_id string, email string) (IAMOrganizationMembership, error)
	GetOrganizationMembership(org_id string, id string) (IAMOrganizationMembership, error)
	CreateOrUpdateOrganizationMembership(org_id string, user_id string, affiliation string, membershipType string, permissions []string) (IAMOrganizationMembership, error)
	DeleteOrganizationMembership(org_id string, id string) error

	// V3 membership list lookups
	GetOrgMembershipV3(org_id string, member_id string) (IAMOrgMembershipV3, error)
	GetOrgMembershipV3ByEmail(org_id string, email string) (IAMOrgMembershipV3, error)
	GetProjectMembershipV3(org_id string, project_id string, member_id string) (IAMProjectMembershipV3, error)
	GetTeamMembershipV3(org_id string, team_id string, member_id string) (IAMTeamMembershipV3, error)

	DeleteOrganizationMembershipPermission(org_id string, member_id string) error
	CreateOrUpdateOrganizationMembershipPermission(member_id, org_id string, permissions []string) (IAMOrganizationMembershipPermission, error)

	GetOrganizationInvitationByEmail(org_id string, email string) (IAMOrganizationInvitation, error)
	CreateOrganizationInvitation(org_id string, email string, permissions []string) (IAMOrganizationInvitation, error)
	DeleteOrganizationInvitation(org_id string, email string) error

	GetProjectMembership(org_id string, project_id string, id string) (IAMProjectMembership, error)
	GetProjectMembershipByEmail(org_id string, project_id string, email string) (IAMProjectMembership, error)
	CreateProjectMembership(org_id string, project_id string, user_id string, membershipType string, permissions []string) (IAMProjectMembership, error)
	UpdateProjectMembership(org_id string, project_id string, user_id string, membershipType string, permissions []string) (IAMProjectMembership, error)
	DeleteProjectMembership(org_id string, project_id string, id string) error

	GetOrganizationServiceaccount(org_id string, id string) (IAMOrganizationServiceaccount, error)
	CreateOrganizationServiceaccount(org_id string, name string, description string) (IAMOrganizationServiceaccount, error)
	UpdateOrganizationServiceaccount(org_id string, serviceaccount_id string, name string, description string) (IAMOrganizationServiceaccount, error)
	DeleteOrganizationServiceaccount(org_id string, id string) error

	GetOrganizationContact(org_id string, id string) (IAMOrganizationContact, error)
	CreateOrganizationContact(org_id string, first_name string, last_name string, notes string, email string, phone string, roles []string) (IAMOrganizationContact, error)
	UpdateOrganizationContact(org_id string, team_id string, first_name string, last_name string, notes string, email string, phone string, roles []string) (IAMOrganizationContact, error)
	DeleteOrganizationContact(org_id string, id string) error

	GetOrganizationTeam(org_id string, id string) (IAMOrganizationTeam, error)
	CreateOrganizationTeam(org_id string, name string, description string, tags []string) (IAMOrganizationTeam, error)
	UpdateOrganizationTeam(org_id string, team_id string, name string, description string, tags []string) (IAMOrganizationTeam, error)
	DeleteOrganizationTeam(org_id string, id string) error

	GetOrganizationTeamPermissions(org_id string, id string) ([]IAMPermissionEntry, error)
	CreateOrganizationTeamPermission(org_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error)
	UpdateOrganizationTeamPermission(org_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error)
	DeleteOrganizationTeamPermission(org_id string, team_id string) error

	GetOrganizationTeamMembership(org_id string, team_id string, id string) (IAMOrganizationTeamMembership, error)
	CreateOrganizationTeamMembership(org_id string, team_id string, member_id string) (IAMOrganizationTeamMembership, error)
	UpdateOrganizationTeamMembership(org_id string, team_id string, member_id string) (IAMOrganizationTeamMembership, error)
	DeleteOrganizationTeamMembership(org_id string, team_id string, id string) error

	GrantOrganizationTeamMemberPermissions(org_id string, team_id string, member_id string, permissions []string) (IAMOrganizationTeamMembershipPermission, error)
	RevokeOrganizationTeamMemberPermissions(org_id string, team_id string, member_id string, permissions []string) (IAMOrganizationTeamMembershipPermission, error)
	UpdateOrganizationTeamMemberPermissions(org_id string, team_id string, member_id string, permissions []string) (IAMOrganizationTeamMembershipPermission, error)

	GetTeamProjects(org_id string, team_id string) ([]IAMTeamProjectWithPermissions, error)

	GetProjectTeamMembership(org_id string, project_id string, team_id string, id string) (IAMProjectTeamMembership, error)
	CreateProjectTeamMembership(org_id string, project_id string, team_id string, member_id string, permissions []string) (IAMProjectTeamMembership, error)
	UpdateProjectTeamMembership(org_id string, project_id string, team_id string, member_id string, permissions []string) (IAMProjectTeamMembership, error)
	DeleteProjectTeamMembership(org_id string, project_id string, team_id string, member_id string) error

	GetProjectTeamPermissions(org_id string, project_id string, team_id string) ([]IAMPermissionEntry, error)
	CreateProjectTeamPermissions(org_id string, project_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error)
	UpdateProjectTeamPermissions(org_id string, project_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error)
	DeleteProjectTeamPermissions(org_id string, project_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error)

	// Glue v3 per-principal permission endpoints
	GetUserOrgPermissions(org_id string, user_id string) ([]IAMPermissionEntry, error)
	PutUserOrgPermissions(org_id string, user_id string, permissions []string) ([]IAMPermissionEntry, error)
	GetServiceAccountOrgPermissions(org_id string, sa_id string) ([]IAMPermissionEntry, error)
	PutServiceAccountOrgPermissions(org_id string, sa_id string, permissions []string) ([]IAMPermissionEntry, error)
	GetUserOrgAffiliation(org_id string, user_id string) (IAMAffiliation, error)
	PutUserOrgAffiliation(org_id string, user_id string, affiliation string) (IAMAffiliation, error)
	GetServiceAccountOrgAffiliation(org_id string, sa_id string) (IAMAffiliation, error)
	PutServiceAccountOrgAffiliation(org_id string, sa_id string, affiliation string) (IAMAffiliation, error)

	GetUserProjectPermissions(org_id string, project_id string, user_id string) ([]IAMPermissionEntry, error)
	PutUserProjectPermissions(org_id string, project_id string, user_id string, permissions []string) ([]IAMPermissionEntry, error)
	GetServiceAccountProjectPermissions(org_id string, project_id string, sa_id string) ([]IAMPermissionEntry, error)
	PutServiceAccountProjectPermissions(org_id string, project_id string, sa_id string, permissions []string) ([]IAMPermissionEntry, error)

	GetUserTeamPermissions(org_id string, team_id string, user_id string) ([]IAMPermissionEntry, error)
	PutUserTeamPermissions(org_id string, team_id string, user_id string, permissions []string) ([]IAMPermissionEntry, error)
	GetServiceAccountTeamPermissions(org_id string, team_id string, sa_id string) ([]IAMPermissionEntry, error)
	PutServiceAccountTeamPermissions(org_id string, team_id string, sa_id string, permissions []string) ([]IAMPermissionEntry, error)

	AddUserToTeam(org_id string, team_id string, user_id string) error
	AddServiceAccountToTeam(org_id string, team_id string, sa_id string) error
	RemoveUserFromTeam(org_id string, team_id string, user_id string) error
	RemoveServiceAccountFromTeam(org_id string, team_id string, sa_id string) error

	GetProjectS3User(org_id string, project_id string, id string) (IAMProjectS3User, error)
	CreateProjectS3User(org_id string, project_id, name string, description string) (IAMProjectS3User, error)
	UpdateProjectS3User(org_id string, project_id string, s3user_id string, name string, description string) (IAMProjectS3User, error)
	DeleteProjectS3User(org_id string, project_id string, id string) error

	CreateProjectS3UserKey(org_id string, project_id string, s3user_id string) (IAMProjectS3UserKey, error)
	DeleteProjectS3UserKey(org_id string, project_id string, s3user_id string, key_id string) error
	GetProjectS3UserKey(org_id string, project_id string, s3user_id string, key_id string) (IAMProjectS3UserKey, error)
}

var _ IAMClient = (*Client)(nil)

func (c *Client) GetOrganization(id string) (IAMOrganization, error) {
	path := fmt.Sprintf(IAMOrganizationEndpoint, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganization{}, errors.Trace(fmt.Errorf(GetOrganizationsError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganization{}, errors.Trace(fmt.Errorf(GetOrganizationsError, err.Error()))
	}

	var iamOrganization IAMOrganization
	err = response.JSONUnmarshall(&iamOrganization)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganization{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamOrganization, nil
}

func (c *Client) GetOrganizationByName(name string) (IAMOrganization, error) {
	path := IAMOrganizationsEndpoint
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganization{}, errors.Trace(fmt.Errorf(GetOrganizationsError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganization{}, errors.Trace(fmt.Errorf(GetOrganizationsError, err.Error()))
	}

	var iamOrganizations []IAMOrganization
	err = response.JSONUnmarshall(&iamOrganizations)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganization{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	for _, io := range iamOrganizations {
		if io.Name == name {
			return io, nil
		}
	}

	return IAMOrganization{}, nil
}

func (c *Client) CreateOrganization(org IAMOrganization) (IAMOrganization, error) {
	var iamOrganization IAMOrganization
	path := IAMOrganizationsEndpoint
	payload, err := json.Marshal(map[string]interface{}{
		"name":        org.Name,
		"description": org.Description,
		"tags":        org.Tags,
		"company_info": map[string]interface{}{
			"street":                   org.CompanyInfo.Street,
			"street_number":            org.CompanyInfo.StreetNumber,
			"zip_code":                 org.CompanyInfo.ZipCode,
			"city":                     org.CompanyInfo.City,
			"country":                  org.CompanyInfo.Country,
			"vat_id":                   org.CompanyInfo.VatID,
			"preferred_billing_method": org.CompanyInfo.PreferredBillingMethod,
			"phone":                    org.CompanyInfo.Phone,
			"accepted_tos":             org.CompanyInfo.AcceptedTos,
			"company_name":             org.CompanyInfo.CompanyName,
		},
	})
	if err != nil {
		return iamOrganization, err
	}

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganization, err
	}

	err = c.checkResponse(response)
	if err != nil {
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, response.Body)
		return iamOrganization, fmt.Errorf(CreateOrganizationError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganization)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganization, err
	}

	return iamOrganization, nil
}

func (c *Client) UpdateOrganization(id string, org IAMOrganization) (IAMOrganization, error) {
	var iamOrganization IAMOrganization
	path := fmt.Sprintf(IAMOrganizationEndpoint, id)
	payload, err := json.Marshal(map[string]interface{}{
		"description": org.Description,
		"tags":        org.Tags,
	})
	if err != nil {
		return iamOrganization, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganization, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganization, fmt.Errorf(UpdateOrganizationError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganization)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganization, err
	}
	return iamOrganization, nil
}

func (c *Client) DeleteOrganization(id string) error {
	path := fmt.Sprintf(IAMOrganizationEndpoint, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationError, err.Error())
	}
	return nil
}

func (c *Client) GetProject(org_id string, id string) (IAMProject, error) {
	path := fmt.Sprintf(IAMProjectEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMProject{}, errors.Trace(fmt.Errorf(GetProjectError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMProject{}, errors.Trace(fmt.Errorf(GetProjectError, err.Error()))
	}

	var iamProject IAMProject
	err = response.JSONUnmarshall(&iamProject)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMProject{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamProject, nil
}

func (c *Client) CreateProject(org_id string, name string, description string, tags []string) (IAMProject, error) {
	var iamProject IAMProject
	path := fmt.Sprintf(IAMProjectsEndpoint, org_id)
	payload, err := json.Marshal(map[string]interface{}{
		"name":        name,
		"description": description,
		"tags":        tags,
	})
	if err != nil {
		return iamProject, err
	}

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProject, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProject, fmt.Errorf(CreateProjectError, err.Error())
	}

	err = response.JSONUnmarshall(&iamProject)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProject, err
	}
	return iamProject, nil
}

func (c *Client) UpdateProject(org_id string, id string, name string, description string, tags []string) (IAMProject, error) {
	var iamProject IAMProject
	path := fmt.Sprintf(IAMProjectEndpoint, org_id, id)
	payload, err := json.Marshal(map[string]interface{}{
		"name":        name,
		"description": description,
		"tags":        tags,
	})
	if err != nil {
		return iamProject, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProject, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProject, fmt.Errorf(UpdateProjectError, err.Error())
	}

	err = response.JSONUnmarshall(&iamProject)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProject, err
	}
	return iamProject, nil
}

func (c *Client) DeleteProject(org_id string, id string) error {
	path := fmt.Sprintf(IAMProjectEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteProjectError, err.Error())
	}
	return nil
}

func (c *Client) GetOrganizationMembership(org_id string, id string) (IAMOrganizationMembership, error) {
	path := fmt.Sprintf(IAMOrganizationMembershipEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganizationMembership{}, errors.Trace(fmt.Errorf(GetOrganizationMembershipError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganizationMembership{}, errors.Trace(fmt.Errorf(GetOrganizationMembershipError, err.Error()))
	}

	var iamOrganizationMembership IAMOrganizationMembership
	err = response.JSONUnmarshall(&iamOrganizationMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganizationMembership{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamOrganizationMembership, nil
}

func (c *Client) GetOrganizationMembershipByEmail(ctx context.Context, org_id string, email string) (IAMOrganizationMembership, error) {
	path := fmt.Sprintf(IAMOrganizationMembershipsEndpoint, org_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganizationMembership{}, errors.Trace(fmt.Errorf(GetOrganizationMembershipError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganizationMembership{}, errors.Trace(fmt.Errorf(GetOrganizationMembershipError, err.Error()))
	}

	var iamOrganizationMemberships []IAMOrganizationMembership
	err = response.JSONUnmarshall(&iamOrganizationMemberships)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganizationMembership{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	for _, iom := range iamOrganizationMemberships {
		if iom.User.Email == email {
			return iom, nil
		}
	}

	return IAMOrganizationMembership{}, fmt.Errorf("membership with that e-mail address was not found: %s", email)
}

// V3 membership list lookups

func (c *Client) GetOrgMembershipV3(org_id string, member_id string) (IAMOrgMembershipV3, error) {
	path := fmt.Sprintf(IAMOrganizationMembershipsEndpoint, org_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrgMembershipV3{}, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrgMembershipV3{}, errors.Trace(err)
	}
	var members []IAMOrgMembershipV3
	err = response.JSONUnmarshall(&members)
	if err != nil {
		body, _ := response.StringBody()
		return IAMOrgMembershipV3{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	for _, m := range members {
		if m.ID == member_id {
			return m, nil
		}
	}
	return IAMOrgMembershipV3{}, fmt.Errorf("membership with id %s not found in organization %s", member_id, org_id)
}

func (c *Client) GetOrgMembershipV3ByEmail(org_id string, email string) (IAMOrgMembershipV3, error) {
	path := fmt.Sprintf(IAMOrganizationMembershipsEndpoint, org_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrgMembershipV3{}, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrgMembershipV3{}, errors.Trace(err)
	}
	var members []IAMOrgMembershipV3
	err = response.JSONUnmarshall(&members)
	if err != nil {
		body, _ := response.StringBody()
		return IAMOrgMembershipV3{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	for _, m := range members {
		if m.Type == "user" && m.DisplayName == email {
			return m, nil
		}
	}
	return IAMOrgMembershipV3{}, fmt.Errorf("membership with email %s not found in organization %s", email, org_id)
}

func (c *Client) GetProjectMembershipV3(org_id string, project_id string, member_id string) (IAMProjectMembershipV3, error) {
	path := fmt.Sprintf(IAMProjectMembershipsEndpoint, org_id, project_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMProjectMembershipV3{}, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMProjectMembershipV3{}, errors.Trace(err)
	}
	var members []IAMProjectMembershipV3
	err = response.JSONUnmarshall(&members)
	if err != nil {
		body, _ := response.StringBody()
		return IAMProjectMembershipV3{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	for _, m := range members {
		if m.ID == member_id {
			return m, nil
		}
	}
	return IAMProjectMembershipV3{}, fmt.Errorf("membership with id %s not found in project %s", member_id, project_id)
}

func (c *Client) GetTeamMembershipV3(org_id string, team_id string, member_id string) (IAMTeamMembershipV3, error) {
	path := fmt.Sprintf(IAMOrganizationTeamMembershipsEndpoint, org_id, team_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMTeamMembershipV3{}, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMTeamMembershipV3{}, errors.Trace(err)
	}
	var members []IAMTeamMembershipV3
	err = response.JSONUnmarshall(&members)
	if err != nil {
		body, _ := response.StringBody()
		return IAMTeamMembershipV3{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	for _, m := range members {
		if m.ID == member_id {
			return m, nil
		}
	}
	return IAMTeamMembershipV3{}, fmt.Errorf("membership with id %s not found in team %s", member_id, team_id)
}

func (c *Client) CreateOrUpdateOrganizationMembership(org_id string, user_id string, affiliation string, membershipType string, permissions []string) (IAMOrganizationMembership, error) {
	iamOrganizationMembership, err := c.GetOrganizationMembership(org_id, user_id)
	if err != nil {
		return iamOrganizationMembership, err
	}

	path := fmt.Sprintf(IAMOrganizationMembershipEndpoint, org_id, user_id)
	type membership = map[string]interface{}
	payload, err := json.Marshal(membership{
		"affiliation":          affiliation,
		"membership_type":      membershipType,
		"editable_permissions": permissions,
	})
	if err != nil {
		return iamOrganizationMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPatch, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationMembership, fmt.Errorf(CreateOrganizationMembershipError, err.Error(), path)
	}

	err = response.JSONUnmarshall(&iamOrganizationMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationMembership, err
	}
	return iamOrganizationMembership, nil
}

func (c *Client) UpdateOrganizationMembership(org_id string, user_id string, affiliation string, membershipType string, permissions []string) (IAMOrganizationMembership, error) {
	iamOrganizationMembership, err := c.GetOrganizationMembership(org_id, user_id)
	if err != nil {
		return iamOrganizationMembership, err
	}

	path := fmt.Sprintf(IAMOrganizationMembershipEndpoint, org_id, user_id)
	type membership = map[string]interface{}
	payload, err := json.Marshal(membership{
		"affiliation":          affiliation,
		"membership_type":      membershipType,
		"editable_permissions": permissions,
	})
	if err != nil {
		return iamOrganizationMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPatch, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationMembership, fmt.Errorf(UpdateOrganizationMembershipError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationMembership, err
	}
	return iamOrganizationMembership, nil
}

func (c *Client) DeleteOrganizationMembership(org_id string, id string) error {
	err := c.DeleteOrganizationServiceaccount(org_id, id)
	if err == nil {
		return nil
	}
	path := fmt.Sprintf(IAMOrganizationMembershipEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationMembershipError, err.Error())
	}
	return nil
}

func (c *Client) GetOrganizationInvitationByEmail(org_id string, email string) (IAMOrganizationInvitation, error) {
	path := fmt.Sprintf(IAMOrganizationInvitationsEndpoint, org_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganizationInvitation{}, errors.Trace(fmt.Errorf(GetOrganizationInvitationError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganizationInvitation{}, errors.Trace(fmt.Errorf(GetOrganizationInvitationError, err.Error()))
	}

	var iamOrganizationInvitations []IAMOrganizationInvitation
	err = response.JSONUnmarshall(&iamOrganizationInvitations)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganizationInvitation{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	for _, ioi := range iamOrganizationInvitations {
		if ioi.Email == email {
			return ioi, nil
		}
	}

	// no invitation. user has: 1. accepted the invitation, or 2. needs to be invited
	return IAMOrganizationInvitation{}, nil
}

func (c *Client) CreateOrganizationInvitation(org_id string, email string, permissions []string) (IAMOrganizationInvitation, error) {
	var iamOrganizationInvitation IAMOrganizationInvitation
	path := fmt.Sprintf(IAMOrganizationInvitationsEndpoint, org_id)
	type invitation = map[string]interface{}
	payload, err := json.Marshal([]invitation{{
		"email":       email,
		"permissions": permissions,
	}})
	if err != nil {
		return iamOrganizationInvitation, err
	}

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationInvitation, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationInvitation, fmt.Errorf(CreateOrganizationInvitationError, err.Error(), path)
	}

	var iamOrganizationInvitations []IAMOrganizationInvitation
	err = response.JSONUnmarshall(&iamOrganizationInvitations)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationInvitation, err
	}
	return iamOrganizationInvitations[0], nil
}

func (c *Client) DeleteOrganizationInvitation(org_id string, email string) error {
	invitation, err := c.GetOrganizationInvitationByEmail(org_id, email)
	if err != nil {
		return err
	}
	path := fmt.Sprintf(IAMOrganizationInvitationEndpoint, org_id, invitation.ID)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteProjectError, err.Error())
	}
	return nil
}

func (c *Client) GetProjectMembership(org_id string, project_id string, id string) (IAMProjectMembership, error) {
	path := fmt.Sprintf(IAMProjectMembershipEndpoint, org_id, project_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMProjectMembership{}, errors.Trace(fmt.Errorf(GetProjectMembershipError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMProjectMembership{}, errors.Trace(fmt.Errorf(GetProjectMembershipError, err.Error()))
	}

	var iamProjectMembership IAMProjectMembership
	err = response.JSONUnmarshall(&iamProjectMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMProjectMembership{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamProjectMembership, nil
}

func (c *Client) GetProjectMembershipByEmail(org_id string, project_id string, email string) (IAMProjectMembership, error) {
	path := fmt.Sprintf(IAMProjectMembershipsEndpoint, org_id, project_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMProjectMembership{}, errors.Trace(fmt.Errorf(GetProjectMembershipError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMProjectMembership{}, errors.Trace(fmt.Errorf(GetProjectMembershipError, err.Error()))
	}

	var iamProjectMemberships []IAMProjectMembership
	err = response.JSONUnmarshall(&iamProjectMemberships)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMProjectMembership{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	for _, iom := range iamProjectMemberships {
		if iom.User.Email == email {
			return iom, nil
		}
	}

	return IAMProjectMembership{}, fmt.Errorf("membership with that e-mail address was not found: %s", email)
}

func (c *Client) CreateProjectMembership(org_id string, project_id string, user_id string, membershipType string, permissions []string) (IAMProjectMembership, error) {
	var iamProjectMembership IAMProjectMembership
	path := fmt.Sprintf(IAMProjectMembershipEndpoint, org_id, project_id, user_id)

	body := map[string]any{
		"membership_type": membershipType,
		"permissions":     permissions,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return iamProjectMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPatch, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProjectMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProjectMembership, fmt.Errorf(CreateProjectMembershipError, err.Error())
	}

	err = response.JSONUnmarshall(&iamProjectMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProjectMembership, err
	}
	return iamProjectMembership, nil
}

func (c *Client) UpdateProjectMembership(org_id string, project_id string, user_id string, membershipType string, permissions []string) (IAMProjectMembership, error) {
	var iamProjectMembership IAMProjectMembership
	path := fmt.Sprintf(IAMProjectMembershipEndpoint, org_id, project_id, user_id)

	payload, err := json.Marshal(map[string]interface{}{
		"membership_type": membershipType,
		"permissions":     permissions,
	})
	if err != nil {
		return iamProjectMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPatch, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProjectMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProjectMembership, fmt.Errorf(UpdateProjectMembershipError, err.Error())
	}

	err = response.JSONUnmarshall(&iamProjectMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProjectMembership, err
	}
	return iamProjectMembership, nil
}

func (c *Client) DeleteProjectMembership(org_id string, project_id string, id string) error {
	path := fmt.Sprintf(IAMProjectMembershipEndpoint, org_id, project_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteProjectMembershipError, err.Error())
	}
	return nil
}

func (c *Client) GetOrganizationServiceaccount(org_id string, id string) (IAMOrganizationServiceaccount, error) {
	path := fmt.Sprintf(IAMOrganizationServiceaccountEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganizationServiceaccount{}, errors.Trace(fmt.Errorf(GetOrganizationServiceaccountError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganizationServiceaccount{}, errors.Trace(fmt.Errorf(GetOrganizationServiceaccountError, err.Error()))
	}

	var iamOrganizationServiceaccount IAMOrganizationServiceaccount
	err = response.JSONUnmarshall(&iamOrganizationServiceaccount)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganizationServiceaccount{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamOrganizationServiceaccount, nil
}

func (c *Client) CreateOrganizationServiceaccount(org_id string, name string, description string) (IAMOrganizationServiceaccount, error) {
	var iamOrganizationServiceaccount IAMOrganizationServiceaccount

	// create service account
	path := fmt.Sprintf(IAMOrganizationServiceaccountsEndpoint, org_id)
	type serviceaccount = map[string]interface{}
	payload, err := json.Marshal(serviceaccount{
		"name":        name,
		"description": description,
	})
	if err != nil {
		return iamOrganizationServiceaccount, err
	}

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationServiceaccount, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationServiceaccount, fmt.Errorf(CreateOrganizationServiceaccountError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationServiceaccount)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationServiceaccount, err
	}
	iamOrganizationServiceaccount.OrganizationId = org_id
	return iamOrganizationServiceaccount, nil
}

func (c *Client) UpdateOrganizationServiceaccount(org_id string, serviceaccount_id string, name string, description string) (IAMOrganizationServiceaccount, error) {
	var iamOrganizationServiceaccount IAMOrganizationServiceaccount
	// update service account
	path := fmt.Sprintf(IAMOrganizationServiceaccountEndpoint, org_id, serviceaccount_id)
	type serviceaccount = map[string]interface{}
	payload, err := json.Marshal(serviceaccount{
		"name":        name,
		"description": description,
	})

	response, err := c.client.NewRequest(http.MethodPut, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationServiceaccount, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationServiceaccount, fmt.Errorf(UpdateOrganizationServiceaccountPermissionError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationServiceaccount)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationServiceaccount, err
	}
	iamOrganizationServiceaccount.OrganizationId = org_id
	return iamOrganizationServiceaccount, nil
}

func (c *Client) DeleteOrganizationServiceaccount(org_id string, id string) error {
	path := fmt.Sprintf(IAMOrganizationServiceaccountEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationServiceaccountError, err.Error())
	}
	return nil
}

// organization teams

func (c *Client) GetOrganizationTeam(org_id string, id string) (IAMOrganizationTeam, error) {
	path := fmt.Sprintf(IAMOrganizationTeamEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganizationTeam{}, errors.Trace(fmt.Errorf(GetOrganizationTeamError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganizationTeam{}, errors.Trace(fmt.Errorf(GetOrganizationTeamError, err.Error()))
	}

	var iamOrganizationTeam IAMOrganizationTeam
	err = response.JSONUnmarshall(&iamOrganizationTeam)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganizationTeam{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamOrganizationTeam, nil
}

func (c *Client) GetOrganizationTeamPermissions(org_id string, id string) ([]IAMPermissionEntry, error) {
	path := fmt.Sprintf(IAMOrganizationTeamPermissionsEndpoint, org_id, id)
	return c.getPermissions(path)
}

func (c *Client) CreateOrganizationTeam(org_id string, name string, description string, tags []string) (IAMOrganizationTeam, error) {
	var iamOrganizationTeam IAMOrganizationTeam
	path := fmt.Sprintf(IAMOrganizationTeamsEndpoint, org_id)
	type serviceaccount = map[string]interface{}
	payload, err := json.Marshal(serviceaccount{
		"name":        name,
		"description": description,
		"tags":        tags,
	})

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationTeam, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationTeam, fmt.Errorf(CreateOrganizationTeamError, err.Error(), path)
	}

	err = response.JSONUnmarshall(&iamOrganizationTeam)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationTeam, err
	}
	return iamOrganizationTeam, nil
}

func (c *Client) UpdateOrganizationTeam(org_id string, team_id string, name string, description string, tags []string) (IAMOrganizationTeam, error) {
	var iamOrganizationTeam IAMOrganizationTeam
	path := fmt.Sprintf(IAMOrganizationTeamEndpoint, org_id, team_id)
	type serviceaccount = map[string]interface{}
	payload, err := json.Marshal(serviceaccount{
		"name":        name,
		"description": description,
		"tags":        tags,
	})
	if err != nil {
		return iamOrganizationTeam, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationTeam, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationTeam, fmt.Errorf(UpdateOrganizationTeamError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationTeam)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationTeam, err
	}
	return iamOrganizationTeam, nil
}

func (c *Client) DeleteOrganizationTeam(org_id string, id string) error {
	path := fmt.Sprintf(IAMOrganizationTeamEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationTeamError, err.Error())
	}
	return nil
}

// organization contacts

func (c *Client) GetOrganizationContact(org_id string, id string) (IAMOrganizationContact, error) {
	path := fmt.Sprintf(IAMOrganizationContactEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganizationContact{}, errors.Trace(fmt.Errorf(GetOrganizationContactError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganizationContact{}, errors.Trace(fmt.Errorf(GetOrganizationContactError, err.Error()))
	}

	var iamOrganizationContact IAMOrganizationContact
	err = response.JSONUnmarshall(&iamOrganizationContact)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganizationContact{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamOrganizationContact, nil
}

func (c *Client) CreateOrganizationContact(org_id string, first_name string, last_name string, notes string, email string, phone string, roles []string) (IAMOrganizationContact, error) {
	var iamOrganizationContact IAMOrganizationContact
	path := fmt.Sprintf(IAMOrganizationContactsEndpoint, org_id)
	type serviceaccount = map[string]interface{}
	payload, err := json.Marshal(serviceaccount{
		"first_name": first_name,
		"last_name":  last_name,
		"phone":      phone,
		"email":      email,
		"notes":      notes,
		"roles":      roles,
	})

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationContact, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationContact, fmt.Errorf(CreateOrganizationContactError, err.Error(), path)
	}

	err = response.JSONUnmarshall(&iamOrganizationContact)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationContact, err
	}
	return iamOrganizationContact, nil
}

func (c *Client) UpdateOrganizationContact(org_id string, team_id string, first_name string, last_name string, notes string, email string, phone string, roles []string) (IAMOrganizationContact, error) {
	var iamOrganizationContact IAMOrganizationContact
	path := fmt.Sprintf(IAMOrganizationContactEndpoint, org_id, team_id)
	type serviceaccount = map[string]interface{}
	payload, err := json.Marshal(serviceaccount{
		"first_name": first_name,
		"last_name":  last_name,
		"phone":      phone,
		"email":      email,
		"notes":      notes,
		"roles":      roles,
	})
	if err != nil {
		return iamOrganizationContact, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationContact, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationContact, fmt.Errorf(UpdateOrganizationContactError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationContact)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationContact, err
	}
	return iamOrganizationContact, nil
}

func (c *Client) DeleteOrganizationContact(org_id string, id string) error {
	path := fmt.Sprintf(IAMOrganizationContactEndpoint, org_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationContactError, err.Error())
	}
	return nil
}

// project team permissions

func (c *Client) GetProjectTeamPermissions(org_id string, project_id string, team_id string) ([]IAMPermissionEntry, error) {
	path := fmt.Sprintf(IAMProjectTeamPermissionsEndpoint, org_id, project_id, team_id)
	return c.getPermissions(path)
}

func (c *Client) CreateProjectTeamPermissions(org_id string, project_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error) {
	path := fmt.Sprintf(IAMProjectTeamPermissionsEndpoint, org_id, project_id, team_id)
	return c.putPermissions(path, permissions)
}

func (c *Client) UpdateProjectTeamPermissions(org_id string, project_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error) {
	path := fmt.Sprintf(IAMProjectTeamPermissionsEndpoint, org_id, project_id, team_id)
	return c.putPermissions(path, permissions)
}

func (c *Client) DeleteProjectTeamPermissions(org_id string, project_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error) {
	path := fmt.Sprintf(IAMProjectTeamPermissionsEndpoint, org_id, project_id, team_id)
	return c.putPermissions(path, []string{})
}

// organization team memberships

func (c *Client) GetOrganizationTeamMembership(org_id string, team_id string, id string) (IAMOrganizationTeamMembership, error) {
	path := fmt.Sprintf(IAMOrganizationTeamMembershipEndpoint, org_id, team_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMOrganizationTeamMembership{}, errors.Trace(fmt.Errorf(GetOrganizationTeamMembershipError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMOrganizationTeamMembership{}, errors.Trace(fmt.Errorf(GetOrganizationTeamMembershipError, err.Error()))
	}

	var iamOrganizationTeamMembership IAMOrganizationTeamMembership
	err = response.JSONUnmarshall(&iamOrganizationTeamMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMOrganizationTeamMembership{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamOrganizationTeamMembership, nil
}

func (c *Client) CreateOrganizationTeamMembership(org_id string, team_id string, member_id string) (IAMOrganizationTeamMembership, error) {
	var iamOrganizationTeamMembership IAMOrganizationTeamMembership
	path := fmt.Sprintf(IAMOrganizationTeamMembershipEndpoint, org_id, team_id, member_id)

	response, err := c.client.NewRequest(http.MethodPost, path).
		Do()
	if err != nil {
		return iamOrganizationTeamMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationTeamMembership, fmt.Errorf(CreateOrganizationTeamMembershipError, err.Error(), path)
	}

	err = response.JSONUnmarshall(&iamOrganizationTeamMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationTeamMembership, err
	}
	return iamOrganizationTeamMembership, nil
}

func (c *Client) UpdateOrganizationTeamMembership(org_id string, team_id string, member_id string) (IAMOrganizationTeamMembership, error) {
	var iamOrganizationTeamMembership IAMOrganizationTeamMembership
	path := fmt.Sprintf(IAMOrganizationTeamMembershipEndpoint, org_id, team_id, member_id)

	response, err := c.client.NewRequest(http.MethodPost, path).
		Do()
	if err != nil {
		return iamOrganizationTeamMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationTeamMembership, fmt.Errorf(UpdateOrganizationTeamMembershipError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationTeamMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationTeamMembership, err
	}
	return iamOrganizationTeamMembership, nil
}

func (c *Client) DeleteOrganizationTeamMembership(org_id string, team_id string, id string) error {
	path := fmt.Sprintf(IAMOrganizationTeamMembershipEndpoint, org_id, team_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationTeamMembershipError, err.Error())
	}
	return nil
}

// project team memberships

func (c *Client) GetProjectTeamMembership(org_id string, project_id string, team_id string, id string) (IAMProjectTeamMembership, error) {
	path := fmt.Sprintf(IAMProjectTeamMembershipEndpoint, org_id, project_id, team_id, id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMProjectTeamMembership{}, errors.Trace(fmt.Errorf(GetProjectTeamMembershipError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMProjectTeamMembership{}, errors.Trace(fmt.Errorf(GetProjectTeamMembershipError, err.Error()))
	}

	var iamProjectTeamMembership IAMProjectTeamMembership
	err = response.JSONUnmarshall(&iamProjectTeamMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMProjectTeamMembership{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamProjectTeamMembership, nil
}

func (c *Client) CreateProjectTeamMembership(org_id string, project_id string, team_id string, member_id string, permissions []string) (IAMProjectTeamMembership, error) {
	var iamProjectTeamMembership IAMProjectTeamMembership
	path := fmt.Sprintf(IAMProjectTeamMembershipPermissionsEndpoint, org_id, project_id, team_id, member_id)
	type permissions_payload = map[string]interface{}
	payload, err := json.Marshal(permissions_payload{
		"permissions_to_grant": permissions,
	})
	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProjectTeamMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProjectTeamMembership, fmt.Errorf(CreateProjectTeamMembershipError, err.Error(), path)
	}

	err = response.JSONUnmarshall(&iamProjectTeamMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProjectTeamMembership, err
	}
	return iamProjectTeamMembership, nil
}

func (c *Client) UpdateProjectTeamMembership(org_id string, project_id string, team_id string, member_id string, permissions []string) (IAMProjectTeamMembership, error) {
	var iamProjectTeamMembership IAMProjectTeamMembership
	path := fmt.Sprintf(IAMProjectTeamMembershipPermissionsEndpoint, org_id, project_id, team_id, member_id)
	type permissions_payload = map[string]interface{}
	payload, err := json.Marshal(permissions_payload{
		"new_permissions": permissions,
	})
	if err != nil {
		return iamProjectTeamMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPatch, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProjectTeamMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProjectTeamMembership, fmt.Errorf(UpdateProjectTeamMembershipError, err.Error())
	}

	err = response.JSONUnmarshall(&iamProjectTeamMembership)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProjectTeamMembership, err
	}
	return iamProjectTeamMembership, nil
}

func (c *Client) DeleteProjectTeamMembership(org_id string, project_id string, team_id string, member_id string) error {
	path := fmt.Sprintf(IAMProjectTeamMembershipPermissionsEndpoint, org_id, project_id, team_id, member_id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteProjectTeamMembershipError, err.Error())
	}
	return nil
}

// project s3user memberships

func (c *Client) GetProjectS3User(org_id string, project_id string, id string) (IAMProjectS3User, error) {
	path := fmt.Sprintf(IAMProjectS3UsersEndpoint, org_id, project_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMProjectS3User{}, errors.Trace(fmt.Errorf(GetProjectS3UserError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMProjectS3User{}, errors.Trace(fmt.Errorf(GetProjectS3UserError, err.Error()))
	}

	var iamProjectS3Users []IAMProjectS3User
	err = response.JSONUnmarshall(&iamProjectS3Users)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMProjectS3User{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	for _, ips := range iamProjectS3Users {
		if ips.ID == id {
			return ips, nil
		}
	}

	return IAMProjectS3User{}, nil
}

func (c *Client) CreateProjectS3User(org_id string, project_id, name string, description string) (IAMProjectS3User, error) {
	var iamProjectS3User IAMProjectS3User
	path := fmt.Sprintf(IAMProjectS3UsersEndpoint, org_id, project_id)
	type s3user = map[string]interface{}
	payload, err := json.Marshal(s3user{
		"name":        name,
		"description": description,
	})

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProjectS3User, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProjectS3User, fmt.Errorf(CreateProjectS3UserError, err.Error(), path)
	}

	err = response.JSONUnmarshall(&iamProjectS3User)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProjectS3User, err
	}
	return iamProjectS3User, nil
}

func (c *Client) UpdateProjectS3User(org_id string, project_id string, s3user_id string, name string, description string) (IAMProjectS3User, error) {
	var iamProjectS3User IAMProjectS3User
	path := fmt.Sprintf(IAMProjectS3UserEndpoint, org_id, project_id, s3user_id)
	type s3user = map[string]interface{}
	payload, err := json.Marshal(s3user{
		"name":        name,
		"description": description,
	})
	if err != nil {
		return iamProjectS3User, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProjectS3User, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProjectS3User, fmt.Errorf(UpdateProjectS3UserError, err.Error())
	}

	err = response.JSONUnmarshall(&iamProjectS3User)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProjectS3User, err
	}
	return iamProjectS3User, nil
}

func (c *Client) DeleteProjectS3User(org_id string, project_id string, id string) error {
	path := fmt.Sprintf(IAMProjectS3UserEndpoint, org_id, project_id, id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteProjectS3UserError, err.Error())
	}
	return nil
}

func (c *Client) CreateProjectS3UserKey(org_id string, project_id string, s3user_id string) (IAMProjectS3UserKey, error) {
	var iamProjectS3UserKey IAMProjectS3UserKey
	path := fmt.Sprintf(IAMProjectS3UserKeysEndpoint, org_id, project_id, s3user_id)
	type s3user_key = map[string]interface{}
	payload, err := json.Marshal(s3user_key{
		"key_type": "access",
	})

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamProjectS3UserKey, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamProjectS3UserKey, fmt.Errorf(CreateProjectS3UserKeyError, err.Error(), path)
	}

	err = response.JSONUnmarshall(&iamProjectS3UserKey)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamProjectS3UserKey, err
	}
	return iamProjectS3UserKey, nil
}

func (c *Client) DeleteProjectS3UserKey(org_id string, project_id string, s3user_id string, key_id string) error {
	path := fmt.Sprintf(IAMProjectS3UserKeyEndpoint, org_id, project_id, s3user_id, key_id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteProjectS3UserKeyError, err.Error())
	}
	return nil
}

func (c *Client) GetProjectS3UserKey(org_id string, project_id string, s3user_id string, key_id string) (IAMProjectS3UserKey, error) {
	path := fmt.Sprintf(IAMProjectS3UserKeyEndpoint, org_id, project_id, s3user_id, key_id)
	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return IAMProjectS3UserKey{}, errors.Trace(fmt.Errorf(GetProjectS3UserKeyError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMProjectS3UserKey{}, errors.Trace(fmt.Errorf(GetProjectS3UserKeyError, err.Error()))
	}

	var iamProjectS3UserKey IAMProjectS3UserKey
	err = response.JSONUnmarshall(&iamProjectS3UserKey)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return IAMProjectS3UserKey{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamProjectS3UserKey, nil
}

func (c *Client) CreateOrUpdateOrganizationMembershipPermission(member_id, org_id string, permissions []string) (IAMOrganizationMembershipPermission, error) {
	var iamOrganizationMembershipPermission IAMOrganizationMembershipPermission
	path := fmt.Sprintf(IAMOrganizationMembershipPermissionsEndpoint, org_id, member_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamOrganizationMembershipPermission, err
	}

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationMembershipPermission, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationMembershipPermission, fmt.Errorf(CreateOrganizationMembershipPermissionError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationMembershipPermission)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationMembershipPermission, err
	}
	return iamOrganizationMembershipPermission, nil
}

func (c *Client) DeleteOrganizationMembershipPermission(org_id string, member_id string) error {
	path := fmt.Sprintf(IAMOrganizationMembershipPermissionsEndpoint, org_id, member_id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationMembershipPermissionError, err.Error())
	}
	return nil
}

func (c *Client) CreateOrganizationTeamPermission(org_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error) {
	path := fmt.Sprintf(IAMOrganizationTeamPermissionsEndpoint, org_id, team_id)
	return c.putPermissions(path, permissions)
}

func (c *Client) UpdateOrganizationTeamPermission(org_id string, team_id string, permissions []string) ([]IAMPermissionEntry, error) {
	path := fmt.Sprintf(IAMOrganizationTeamPermissionsEndpoint, org_id, team_id)
	return c.putPermissions(path, permissions)
}

func (c *Client) DeleteOrganizationTeamPermission(org_id string, team_id string) error {
	path := fmt.Sprintf(IAMOrganizationTeamPermissionsEndpoint, org_id, team_id)
	response, err := c.client.NewRequest(http.MethodDelete, path).
		Do()
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	err = c.checkResponse(response)
	if err != nil {
		return fmt.Errorf(DeleteOrganizationTeamPermissionError, err.Error())
	}
	return nil
}

// Organization Team Membership Permissions
func (c *Client) GrantOrganizationTeamMemberPermissions(org_id string, team_id string, member_id string, permissions []string) (IAMOrganizationTeamMembershipPermission, error) {
	var iamOrganizationTeamMembershipPermission IAMOrganizationTeamMembershipPermission
	path := fmt.Sprintf(IAMOrganizationTeamMembershipPermissionsEndpoint, org_id, team_id, member_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamOrganizationTeamMembershipPermission, err
	}

	response, err := c.client.NewRequest(http.MethodPost, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationTeamMembershipPermission, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationTeamMembershipPermission, fmt.Errorf(GrantOrganizationTeamMemberPermissionsError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationTeamMembershipPermission)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationTeamMembershipPermission, err
	}
	return iamOrganizationTeamMembershipPermission, nil
}

func (c *Client) UpdateOrganizationTeamMemberPermissions(org_id string, team_id string, member_id string, permissions []string) (IAMOrganizationTeamMembershipPermission, error) {
	var iamOrganizationTeamMembershipPermission IAMOrganizationTeamMembershipPermission
	path := fmt.Sprintf(IAMOrganizationTeamMembershipPermissionsEndpoint, org_id, team_id, member_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamOrganizationTeamMembershipPermission, err
	}

	response, err := c.client.NewRequest(http.MethodPatch, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationTeamMembershipPermission, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationTeamMembershipPermission, fmt.Errorf(UpdateOrganizationTeamMemberPermissionsError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationTeamMembershipPermission)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationTeamMembershipPermission, err
	}
	return iamOrganizationTeamMembershipPermission, nil
}

func (c *Client) RevokeOrganizationTeamMemberPermissions(org_id string, team_id string, member_id string, permissions []string) (IAMOrganizationTeamMembershipPermission, error) {
	var iamOrganizationTeamMembershipPermission IAMOrganizationTeamMembershipPermission
	path := fmt.Sprintf(IAMOrganizationTeamMembershipPermissionsEndpoint, org_id, team_id, member_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamOrganizationTeamMembershipPermission, err
	}

	response, err := c.client.NewRequest(http.MethodDelete, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationTeamMembershipPermission, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationTeamMembershipPermission, fmt.Errorf(RevokeOrganizationTeamMemberPermissionsError, err.Error())
	}

	err = response.JSONUnmarshall(&iamOrganizationTeamMembershipPermission)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		err = fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
		return iamOrganizationTeamMembershipPermission, err
	}
	return iamOrganizationTeamMembershipPermission, nil
}

func (c *Client) GetTeamProjects(org_id string, team_id string) ([]IAMTeamProjectWithPermissions, error) {
	var iamTeamProjects []IAMTeamProjectWithPermissions
	path := fmt.Sprintf(IAMTeamProjectsEndpoint, org_id, team_id)

	response, err := c.client.NewRequest(http.MethodGet, path).Do()
	if err != nil {
		return iamTeamProjects, errors.Trace(fmt.Errorf(GetTeamProjectsError, err.Error()))
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamTeamProjects, errors.Trace(fmt.Errorf(GetTeamProjectsError, err.Error()))
	}

	err = response.JSONUnmarshall(&iamTeamProjects)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return iamTeamProjects, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamTeamProjects, nil
}

// --- Glue v3 per-principal permission methods ---

func (c *Client) getPermissions(endpoint string) ([]IAMPermissionEntry, error) {
	response, err := c.client.NewRequest(http.MethodGet, endpoint).Do()
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var entries []IAMPermissionEntry
	err = response.JSONUnmarshall(&entries)
	if err != nil {
		body, _ := response.StringBody()
		return nil, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	return entries, nil
}

func (c *Client) putPermissions(endpoint string, permissions []string) ([]IAMPermissionEntry, error) {
	payload, err := json.Marshal(permissions)
	if err != nil {
		return nil, err
	}
	response, err := c.client.NewRequest(http.MethodPut, endpoint).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var entries []IAMPermissionEntry
	err = response.JSONUnmarshall(&entries)
	if err != nil {
		body, _ := response.StringBody()
		return nil, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	return entries, nil
}

func (c *Client) getAffiliation(endpoint string) (IAMAffiliation, error) {
	response, err := c.client.NewRequest(http.MethodGet, endpoint).Do()
	if err != nil {
		return IAMAffiliation{}, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMAffiliation{}, errors.Trace(err)
	}
	var aff string
	err = response.JSONUnmarshall2(&aff)
	if err != nil {
		body, _ := response.StringBody()
		return IAMAffiliation{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	return IAMAffiliation{Affiliation: aff}, nil
}

func (c *Client) putAffiliation(endpoint string, affiliation string) (IAMAffiliation, error) {
	payload, err := json.Marshal(affiliation)
	if err != nil {
		return IAMAffiliation{}, err
	}
	response, err := c.client.NewRequest(http.MethodPut, endpoint).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return IAMAffiliation{}, errors.Trace(err)
	}
	err = c.checkResponse(response)
	if err != nil {
		return IAMAffiliation{}, errors.Trace(err)
	}
	var aff string
	err = response.JSONUnmarshall2(&aff)
	if err != nil {
		body, _ := response.StringBody()
		return IAMAffiliation{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}
	return IAMAffiliation{Affiliation: aff}, nil
}

// Organization-level user permissions

func (c *Client) GetUserOrgPermissions(org_id string, user_id string) ([]IAMPermissionEntry, error) {
	return c.getPermissions(fmt.Sprintf(IAMUserOrgPermissionsEndpoint, org_id, user_id))
}

func (c *Client) PutUserOrgPermissions(org_id string, user_id string, permissions []string) ([]IAMPermissionEntry, error) {
	return c.putPermissions(fmt.Sprintf(IAMUserOrgPermissionsEndpoint, org_id, user_id), permissions)
}

// Organization-level service account permissions

func (c *Client) GetServiceAccountOrgPermissions(org_id string, sa_id string) ([]IAMPermissionEntry, error) {
	return c.getPermissions(fmt.Sprintf(IAMServiceAccountOrgPermissionsEndpoint, org_id, sa_id))
}

func (c *Client) PutServiceAccountOrgPermissions(org_id string, sa_id string, permissions []string) ([]IAMPermissionEntry, error) {
	return c.putPermissions(fmt.Sprintf(IAMServiceAccountOrgPermissionsEndpoint, org_id, sa_id), permissions)
}

// Organization-level affiliation

func (c *Client) GetUserOrgAffiliation(org_id string, user_id string) (IAMAffiliation, error) {
	return c.getAffiliation(fmt.Sprintf(IAMUserOrgAffiliationEndpoint, org_id, user_id))
}

func (c *Client) PutUserOrgAffiliation(org_id string, user_id string, affiliation string) (IAMAffiliation, error) {
	return c.putAffiliation(fmt.Sprintf(IAMUserOrgAffiliationEndpoint, org_id, user_id), affiliation)
}

func (c *Client) GetServiceAccountOrgAffiliation(org_id string, sa_id string) (IAMAffiliation, error) {
	return c.getAffiliation(fmt.Sprintf(IAMServiceAccountOrgAffiliationEndpoint, org_id, sa_id))
}

func (c *Client) PutServiceAccountOrgAffiliation(org_id string, sa_id string, affiliation string) (IAMAffiliation, error) {
	return c.putAffiliation(fmt.Sprintf(IAMServiceAccountOrgAffiliationEndpoint, org_id, sa_id), affiliation)
}

// Project-level user permissions

func (c *Client) GetUserProjectPermissions(org_id string, project_id string, user_id string) ([]IAMPermissionEntry, error) {
	return c.getPermissions(fmt.Sprintf(IAMUserProjectPermissionsEndpoint, org_id, project_id, user_id))
}

func (c *Client) PutUserProjectPermissions(org_id string, project_id string, user_id string, permissions []string) ([]IAMPermissionEntry, error) {
	return c.putPermissions(fmt.Sprintf(IAMUserProjectPermissionsEndpoint, org_id, project_id, user_id), permissions)
}

// Project-level service account permissions

func (c *Client) GetServiceAccountProjectPermissions(org_id string, project_id string, sa_id string) ([]IAMPermissionEntry, error) {
	return c.getPermissions(fmt.Sprintf(IAMServiceAccountProjectPermissionsEndpoint, org_id, project_id, sa_id))
}

func (c *Client) PutServiceAccountProjectPermissions(org_id string, project_id string, sa_id string, permissions []string) ([]IAMPermissionEntry, error) {
	return c.putPermissions(fmt.Sprintf(IAMServiceAccountProjectPermissionsEndpoint, org_id, project_id, sa_id), permissions)
}

// Team-level user permissions

func (c *Client) GetUserTeamPermissions(org_id string, team_id string, user_id string) ([]IAMPermissionEntry, error) {
	return c.getPermissions(fmt.Sprintf(IAMUserTeamPermissionsEndpoint, org_id, team_id, user_id))
}

func (c *Client) PutUserTeamPermissions(org_id string, team_id string, user_id string, permissions []string) ([]IAMPermissionEntry, error) {
	return c.putPermissions(fmt.Sprintf(IAMUserTeamPermissionsEndpoint, org_id, team_id, user_id), permissions)
}

// Team-level service account permissions

func (c *Client) GetServiceAccountTeamPermissions(org_id string, team_id string, sa_id string) ([]IAMPermissionEntry, error) {
	return c.getPermissions(fmt.Sprintf(IAMServiceAccountTeamPermissionsEndpoint, org_id, team_id, sa_id))
}

func (c *Client) PutServiceAccountTeamPermissions(org_id string, team_id string, sa_id string, permissions []string) ([]IAMPermissionEntry, error) {
	return c.putPermissions(fmt.Sprintf(IAMServiceAccountTeamPermissionsEndpoint, org_id, team_id, sa_id), permissions)
}

// Team member add/remove

func (c *Client) AddUserToTeam(org_id string, team_id string, user_id string) error {
	path := fmt.Sprintf(IAMUserTeamMemberEndpoint, org_id, team_id, user_id)
	response, err := c.client.NewRequest(http.MethodPut, path).Do()
	if err != nil {
		return errors.Trace(err)
	}
	return c.checkResponse(response)
}

func (c *Client) AddServiceAccountToTeam(org_id string, team_id string, sa_id string) error {
	path := fmt.Sprintf(IAMServiceAccountTeamMemberEndpoint, org_id, team_id, sa_id)
	response, err := c.client.NewRequest(http.MethodPut, path).Do()
	if err != nil {
		return errors.Trace(err)
	}
	return c.checkResponse(response)
}

func (c *Client) RemoveUserFromTeam(org_id string, team_id string, user_id string) error {
	path := fmt.Sprintf(IAMUserTeamMemberEndpoint, org_id, team_id, user_id)
	response, err := c.client.NewRequest(http.MethodDelete, path).Do()
	if err != nil {
		return errors.Trace(err)
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	return c.checkResponse(response)
}

func (c *Client) RemoveServiceAccountFromTeam(org_id string, team_id string, sa_id string) error {
	path := fmt.Sprintf(IAMServiceAccountTeamMemberEndpoint, org_id, team_id, sa_id)
	response, err := c.client.NewRequest(http.MethodDelete, path).Do()
	if err != nil {
		return errors.Trace(err)
	}
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	return c.checkResponse(response)
}
