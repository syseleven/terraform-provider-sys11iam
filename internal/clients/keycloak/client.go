package keycloak

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/syseleven/terraform-provider-sys11iam/internal/clients/rest"
	"github.com/syseleven/terraform-provider-sys11iam/internal/logging"
)

type Client struct {
	client *rest.Client
	auth   struct {
		username     string
		password     string
		clientId     string
		clientSecret string
		clientScope  string
	}
}

func NewClient(url string, timeout time.Duration) *Client {
	// remove trailing slash
	if len(url) > 1 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	// remove api suffix
	if strings.HasSuffix(url, "/api") {
		url = url[:len(url)-4]
	}
	return &Client{
		client: rest.NewClient(url).WithTimeout(timeout),
	}
}

func (c *Client) WithContext(ctx *rest.Context) *Client {
	c.client.UseContext(ctx)
	return c
}

func (c *Client) WithContextFromRequest(r *http.Request) *Client {
	c.client.WithContext(r)
	return c
}

func (c *Client) WithClientConfig(clientId string, clientSecret string, clientScope string, clientUserName string, clientPassword string) *Client {
	c.auth.clientId = clientId
	c.auth.clientSecret = clientSecret
	c.auth.clientScope = clientScope
	c.auth.username = clientUserName
	c.auth.password = clientPassword
	return c
}

func (c Client) Health() error {
	// check for availability and auth by using
	resp, err := c.client.NewRequest(http.MethodGet, "/").Do()
	if err != nil {
		return err
	}

	statusIsOK := (resp.StatusCode >= http.StatusOK) && (resp.StatusCode < 300)

	if !statusIsOK {
		return fmt.Errorf("invalid status code received: %d", resp.StatusCode)
	}

	return nil
}

var errorResponse struct {
	Errors []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) checkResponse(response *rest.Response) error {
	genericError := func(message string) error {
		errorMsg := fmt.Sprintf(
			"unexpected response from keycloak service: HTTP %d - %s",
			response.StatusCode,
			message,
		)

		logging.Error(errorMsg)
		return errors.New(errorMsg)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		buf, err := response.StringBody()
		if err != nil {
			return genericError(err.Error())
		}

		return genericError(buf)
	}
	return nil
}
