package keycloak

import (
	"fmt"
	"net/http"
	"net/url"
)

type AuthResponse struct {
	// auth token
	AuthToken         string `json:"access_token"`
	RefreshToken      string `json:"refresh_token,omitempty"`
	SessionState      string `json:"session_state,omitempty"`
	ExpiresIn         int    `json:"expires_in,omitempty"`
	RefreshExpires_in int    `json:"refresh_expires_in,omitempty"`
	TokenType         string `json:"token_type,omitempty"`
	NotBeforePolicy   int    `json:"not-before-policy,omitempty"`
	Scope             string `json:"scope,omitempty"`
}

func (c *Client) Login() (string, error) {
	formValues := make(url.Values, 0)
	formValues.Add("grant_type", "password")
	formValues.Add("username", c.auth.username)
	formValues.Add("password", c.auth.password)
	formValues.Add("client_id", c.auth.clientId)
	formValues.Add("client_secret", c.auth.clientSecret)
	formValues.Add("scope", c.auth.clientScope)

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
