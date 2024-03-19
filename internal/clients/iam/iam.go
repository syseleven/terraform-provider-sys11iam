package iam

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/errors"
)

const IAMOrganizationsEndpoint string = "/v1/orgs"
const IAMOrganizationEndpoint string = "/v1/orgs/%s"

const IAMProjectsEndpoint string = "/v1/orgs/%s/projects"
const IAMProjectEndpoint string = "/v1/orgs/%s/projects/%s"

const IAMOrganizationMembershipsEndpoint string = "/v1/orgs/%s/memberships"
const IAMOrganizationMembershipEndpoint string = "/v1/orgs/%s/memberships/%s"
const IAMOrganizationMembershipPermissionsEndpoint string = "/v1/orgs/%s/memberships/%s/permissions"

const IAMOrganizationInvitationsEndpoint string = "/v1/orgs/%s/invitations"
const IAMOrganizationInvitationEndpoint string = "/v1/orgs/%s/invitations/%s"

const IAMProjectMembershipsEndpoint string = "/v1/orgs/%s/projects/%s/memberships"
const IAMProjectMembershipEndpoint string = "/v1/orgs/%s/projects/%s/memberships/%s"
const IAMProjectMembershipPermissionsEndpoint string = "/v1/orgs/%s/projects/%s/memberships/%s/permissions"

type IAMOrganization struct {
	// org id
	ID string `json:"id"`

	// org name
	Name string `json:"name"`

	// org description
	Description string `json:"description"`

	// org tags
	Tags []string `json:"tags"`

	// org created_at
	CreatedAt string `json:"created_at"`

	// org is_active
	IsActive bool `json:"is_active"`

	// org updated_at
	UpdatedAt string `json:"updated_at"`
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
}

type IAMOrganizationMembership struct {
	// OrganizationMembership id
	ID string `json:"id"`

	// OrganizationMembership email
	Email string `json:"email"`

	// org id
	OrganizationId string `json:"org_id"`

	// org name
	OrganizationName string `json:"org_name,omitempty"`

	// OrganizationMembership permissions
	Permissions []string `json:"editable_permissions,omitempty"`
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
}

type IAMProjectMembership struct {
	// ProjectMembership id
	ID string `json:"id,omitempty"`

	// ProjectMembership email
	Email string `json:"email,omitempty"`

	// project id
	ProjectId string `json:"project_id,omitempty"`

	// project name
	ProjectName string `json:"project_name,omitempty"`

	// ProjectMembership permissions
	Permissions []string `json:"permissions,omitempty"`
}

func (c *Client) GetOrganizations() ([]IAMOrganization, error) {
	response, err := c.client.NewRequest(http.MethodGet, IAMOrganizationsEndpoint).Do()
	if err != nil {
		return []IAMOrganization{}, errors.Trace(fmt.Errorf(GetOrganizationsError, err.Error()))
	}
	err = c.checkResponse(response)
	if err != nil {
		return []IAMOrganization{}, errors.Trace(fmt.Errorf(GetOrganizationsError, err.Error()))
	}

	var iamOrganizations []IAMOrganization
	err = response.JSONUnmarshall(&iamOrganizations)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return []IAMOrganization{}, errors.Trace(fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body))
	}

	return iamOrganizations, nil
}

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

func (c *Client) CreateOrganization(name string, description string, tags []string) (IAMOrganization, error) {
	var iamOrganization IAMOrganization
	path := IAMOrganizationsEndpoint
	payload, err := json.Marshal(map[string]interface{}{
		"name":        name,
		"description": description,
		"tags":        tags,
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

func (c *Client) UpdateOrganization(id string, description string, tags []string) (IAMOrganization, error) {
	var iamOrganization IAMOrganization
	path := fmt.Sprintf(IAMOrganizationEndpoint, id)
	raw_payload := map[string]interface{}{}
	if description != "" {
		raw_payload["description"] = description
	}
	if len(tags) != 0 {
		raw_payload["tags"] = tags
	}

	payload, err := json.Marshal(raw_payload)
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

func (c *Client) GetOrganizationMembershipByEmail(org_id string, email string) (IAMOrganizationMembership, error) {
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
		if iom.Email == email {
			return iom, nil
		}
	}

	return IAMOrganizationMembership{}, fmt.Errorf("membership with that e-mail address was not found: %s", email)
}

func (c *Client) CreateOrganizationMembership(org_id string, user_id string, permissions []string) (IAMOrganizationMembership, error) {
	var iamOrganizationMembership IAMOrganizationMembership
	path := fmt.Sprintf(IAMOrganizationMembershipPermissionsEndpoint, org_id, user_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamOrganizationMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
		UseJSONPayload(payload).
		Do()
	if err != nil {
		return iamOrganizationMembership, err
	}

	err = c.checkResponse(response)
	if err != nil {
		return iamOrganizationMembership, fmt.Errorf(CreateOrganizationMembershipError, err.Error())
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

func (c *Client) UpdateOrganizationMembership(org_id string, user_id string, permissions []string) (IAMOrganizationMembership, error) {
	var iamOrganizationMembership IAMOrganizationMembership
	path := fmt.Sprintf(IAMOrganizationMembershipPermissionsEndpoint, org_id, user_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamOrganizationMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
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

	return IAMOrganizationInvitation{}, fmt.Errorf("organization invitation with that e-mail address was not found: %s", email)
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
		return iamOrganizationInvitation, fmt.Errorf(CreateOrganizationInvitationError, err.Error())
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
		if iom.Email == email {
			return iom, nil
		}
	}

	return IAMProjectMembership{}, fmt.Errorf("membership with that e-mail address was not found: %s", email)
}

func (c *Client) CreateProjectMembership(org_id string, project_id string, user_id string, permissions []string) (IAMProjectMembership, error) {
	var iamProjectMembership IAMProjectMembership
	path := fmt.Sprintf(IAMProjectMembershipPermissionsEndpoint, org_id, project_id, user_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamProjectMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
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

func (c *Client) UpdateProjectMembership(org_id string, project_id string, user_id string, permissions []string) (IAMProjectMembership, error) {
	var iamProjectMembership IAMProjectMembership
	path := fmt.Sprintf(IAMProjectMembershipPermissionsEndpoint, org_id, project_id, user_id)
	payload, err := json.Marshal(permissions)
	if err != nil {
		return iamProjectMembership, err
	}

	response, err := c.client.NewRequest(http.MethodPut, path).
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
