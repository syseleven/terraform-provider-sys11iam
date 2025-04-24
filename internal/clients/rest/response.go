package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	*http.Response
	byteBody    *[]byte
	requestBody []byte
}

type DebugInfo struct {
	Request struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Query  string `json:"query,omitempty"`
		Body   []byte `json:"body,omitempty"`
	} `json:"request"`
	Response struct {
		Code int    `json:"code"`
		Body []byte `json:"body,omitempty"`
	} `json:"response"`
}

func newResponse(response *http.Response, requestBody []byte) *Response {
	return &Response{
		response,
		nil,
		requestBody,
	}
}

func (resp *Response) StringBody() (string, error) {
	byteBody, err := resp.ByteBody()
	return string(byteBody), err
}

func (resp *Response) ByteBody() ([]byte, error) {
	err := resp.readBody()
	if err != nil {
		return []byte{}, err
	}
	return *resp.byteBody, nil
}

func (resp *Response) JSONUnmarshall2(v interface{}) error {
	err := resp.readBody()
	if err != nil {
		return err
	}
	return json.Unmarshal(*resp.byteBody, v)
}

func (resp *Response) JSONUnmarshall(v interface{}) error {
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&v)
	if err != nil {
		return fmt.Errorf("Err: %s ; Body: %s ; From: %s %s", err.Error(), bodyString, resp.Request.Method, resp.Request.URL)
	}
	validate := validator.New()
	err = validate.Var(v, "dive")
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return fmt.Errorf("validation error: %v ; Parsed: %v ; Body: %s", validationErrors, v, bodyString)
	}
	return nil
}

func (resp *Response) readBody() error {
	if resp.byteBody != nil {
		return nil
	}

	byteBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.byteBody = &byteBody
	return resp.Body.Close()
}

func (resp *Response) DebugInfo() *DebugInfo {
	var debugInfo DebugInfo

	debugInfo.Request.Path = resp.Request.URL.Path
	debugInfo.Request.Query = resp.Request.URL.Query().Encode()
	debugInfo.Request.Body = resp.requestBody
	debugInfo.Request.Method = resp.Request.Method
	debugInfo.Response.Code = resp.StatusCode

	responseBody, err := resp.ByteBody()
	if err != nil {
		responseBody = []byte("unable to parse body")
	}
	debugInfo.Response.Body = responseBody

	return &debugInfo
}

func (di *DebugInfo) AsJSON() ([]byte, error) {
	return json.Marshal(di)
}
