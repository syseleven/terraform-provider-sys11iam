package rest

import (
	"crypto/tls"
	"net/http"
	"time"
)

const RequestIDHeader = "X-Request-Id"
const AuthorizationHeader = "Authorization"

type Client struct {
	url  string
	auth struct {
		username            string
		password            string
		bearerToken         string
		xAuthToken          string
		serviceAccountToken string
		clientId            string
		clientSecret        string
		clientScope         string
	}
	requestID      string
	defaultHeaders map[string]string
	client         *http.Client
}

func NewClient(url string) *Client {
	if len(url) > 1 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	return &Client{
		url:            url,
		client:         &http.Client{},
		defaultHeaders: make(map[string]string, 0),
	}
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.client.Timeout = timeout * time.Second
	return c
}

func (c *Client) SkipSSLVerify() *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	c.client.Transport = tr
	return c
}

func (c *Client) WithLogin(username, password string) *Client {
	c.auth.username = username
	c.auth.password = password
	return c
}

func (c *Client) WithXAuthToken(token string) *Client {
	c.auth.xAuthToken = token
	return c
}

func (c *Client) WithBearerToken(token string) *Client {
	c.auth.bearerToken = token
	return c
}

func (c *Client) WithRequestID(requestID string) *Client {
	c.requestID = requestID
	return c
}

// Extracts and persists "X-Request-Id" and "Authorization"
// from *http.Request so all follow-up requests will use them
func (c *Client) WithContext(r *http.Request) *Client {
	context := ContextFromRequest(r)
	c.requestID = context.RequestID
	c.auth.bearerToken = context.BearerToken
	return c
}

func (c *Client) UseContext(ctx *Context) *Client {
	if ctx != nil {
		c.requestID = ctx.RequestID
		c.auth.bearerToken = ctx.BearerToken
	}
	return c
}

// AddDefaultHeader will use the header on every call the client will make
func (c *Client) AddDefaultHeader(key, value string) *Client {
	c.defaultHeaders[key] = value
	return c
}

// Used for testing
func (c *Client) WithHTTPClient(client *http.Client) *Client {
	c.client = client
	return c
}

func (c *Client) NewRequest(method, path string) *Request {
	return &Request{
		client:  c,
		method:  method,
		path:    path,
		headers: c.defaultHeaders,
	}
}
