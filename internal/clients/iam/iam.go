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

type IAMOrganization struct {
	// org id
	ID string `json:"id,omitempty"`

	// org id
	OrganizationId string `json:"organization_id,omitempty"`

	// org name
	Name string `json:"name,omitempty"`

	// org description
	Description string `json:"description,omitempty"`

	// org tags
	Tags []string `json:"tags,omitempty"`

	// org created_at
	CreatedAt string `json:"created_at,omitempty"`

	// org is_active
	IsActive bool `json:"is_active,omitempty"`

	// org updated_at
	UpdatedAt string `json:"updated_at,omitempty"`
}

type IAMProject struct {
	// org id
	ID string `json:"id,omitempty"`

	// org id
	OrganizationId string `json:"organization_id,omitempty"`

	// project id
	ProjectId string `json:"project_id,omitempty"`

	// project name
	Name string `json:"name,omitempty"`

	// project description
	Description string `json:"description,omitempty"`

	// project tags
	Tags []string `json:"tags,omitempty"`
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

func (c *Client) UpdateProject(org_id string, id string, description string, tags []string) (IAMProject, error) {
	var iamProject IAMProject
	path := fmt.Sprintf(IAMProjectEndpoint, org_id, id)
	raw_payload := map[string]interface{}{}
	if description != "" {
		raw_payload["description"] = description
	}
	if len(tags) != 0 {
		raw_payload["tags"] = tags
	}

	payload, err := json.Marshal(raw_payload)
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
