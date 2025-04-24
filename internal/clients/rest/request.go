package rest

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	client      *Client
	method      string
	path        string
	jsonPayload []byte
	formData    url.Values
	headers     map[string]string
}

func (req *Request) UseJSONPayload(payload []byte) *Request {
	req.jsonPayload = payload
	return req
}

func (req *Request) UseFormData(formData url.Values) *Request {
	req.formData = formData
	return req
}

func (req *Request) AddHeader(key, value string) *Request {
	req.headers[key] = value
	return req
}

func (req *Request) Do() (resp *Response, err error) {
	request, err := http.NewRequest(req.method, req.client.url+req.path, nil)
	if err != nil {
		return
	}

	if len(req.jsonPayload) > 0 && len(req.formData) > 0 {
		return nil, errors.New("cannot use JSON payload and form data at the same time")
	}

	var requestBody []byte
	if len(req.jsonPayload) > 0 {
		request.Header.Add("Content-Type", "application/json")
		request.ContentLength = int64(len(req.jsonPayload))
		request.Body = ioutil.NopCloser(bytes.NewBuffer(req.jsonPayload))
		request.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(req.jsonPayload)
			return ioutil.NopCloser(r), nil
		}
		requestBody = req.jsonPayload
	}

	if len(req.formData) > 0 {
		body := req.formData.Encode()

		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		request.ContentLength = int64(len(body))
		request.Body = ioutil.NopCloser(strings.NewReader(body))
		request.GetBody = func() (io.ReadCloser, error) {
			r := strings.NewReader(body)
			return ioutil.NopCloser(r), nil
		}

		requestBody = []byte(body)
	}

	if len(req.client.requestID) > 0 {
		request.Header.Add("X-Request-Id", req.client.requestID)
	}

	if len(req.client.auth.bearerToken) > 0 {
		request.Header.Add("Authorization", "Bearer "+req.client.auth.bearerToken)
	}
	if len(req.client.auth.xAuthToken) > 0 {
		request.Header.Add("X-Auth-Token", req.client.auth.xAuthToken)
	}
	if len(req.client.auth.username) > 0 {
		request.SetBasicAuth(req.client.auth.username, req.client.auth.password)
	}

	for k, v := range req.headers {
		request.Header.Add(k, v)
	}

	response, err := req.client.client.Do(request)
	if err != nil {
		return
	}

	return newResponse(response, requestBody), nil
}
