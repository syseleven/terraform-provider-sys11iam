package rest

import (
	"net/http"
	"strings"
)

type Context struct {
	RequestID   string
	BearerToken string
}

func ContextFromRequest(r *http.Request) *Context {
	c := Context{}
	c.RequestID = r.Header.Get(RequestIDHeader)
	tokenParts := strings.Split(r.Header.Get(AuthorizationHeader), " ")
	c.BearerToken = tokenParts[len(tokenParts)-1]
	return &c
}
