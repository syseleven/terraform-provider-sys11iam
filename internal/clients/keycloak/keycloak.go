package keycloak

import (
	"net/url"

	"fmt"
	"net/http"
)

type AuthResponse struct {
	// auth token
	AuthToken string `json:"access_token,omitempty"`
}

func (c *Client) Login() (string, error) {
	formValues := make(url.Values, 0)
	formValues.Add("grant_type", "client_credentials")
	formValues.Add("scope", c.auth.clientScope)
	formValues.Add("client_id", c.auth.clientId)
	formValues.Add("client_secret", c.auth.clientSecret)

	response, err := c.client.NewRequest(http.MethodPost, "").
		UseFormData(formValues).
		Do()

	if err != nil {
		return "", err
	}
	err = c.checkResponse(response)
	if err != nil {
		return "", err
	}
	var authResponse AuthResponse
	err = response.JSONUnmarshall(&authResponse)
	if err != nil {
		body, respErr := response.StringBody()
		if respErr != nil {
			body = "unable to parse body"
		}
		return "", fmt.Errorf("%s (code: %d, body: %s)", err.Error(), response.StatusCode, body)
	}

	return authResponse.AuthToken, nil
}
