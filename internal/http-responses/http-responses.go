package responses

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type Expectation struct {
	request struct {
		method          string
		path            string
		queryParameters map[string]string
		headers         map[string]string
		jsonParameters  interface{}
		body            *[]byte
	}
	response struct {
		code    int
		body    []byte
		headers map[string]string
	}
}

func Expect(method, path string) *Expectation {
	expectation := &Expectation{}
	expectation.request.method = method
	expectation.request.path = path
	expectation.response.headers = make(map[string]string)
	return expectation
}

func (ex *Expectation) WithQueryParameters(queryParameters map[string]string) *Expectation {
	ex.request.queryParameters = queryParameters
	return ex
}

func (ex *Expectation) WithJSONParameters(jsonParameters interface{}) *Expectation {
	ex.request.jsonParameters = jsonParameters
	return ex
}

func (ex *Expectation) WithHeaders(headers map[string]string) *Expectation {
	ex.request.headers = headers
	return ex
}

func (ex *Expectation) WithBody(body []byte) *Expectation {
	ex.request.body = &body
	return ex
}

func (ex *Expectation) ReturnWithCode(code int) *Expectation {
	ex.response.code = code
	return ex
}

func (ex *Expectation) ReturnWithHeaders(headers map[string]string) *Expectation {
	for k, v := range headers {
		ex.response.headers[k] = v
	}
	return ex
}

func (ex *Expectation) ReturnWithBody(body []byte) *Expectation {
	ex.response.body = body
	return ex
}

func (ex *Expectation) ReturnWithJSONFile(path string) *Expectation {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		ex.response.body = []byte(fmt.Sprintf("unable to read json file: %s", err.Error()))
	}

	ex.response.body = buf
	ex.response.headers["Content-Type"] = "application/json"
	return ex
}

type MockServer struct {
	URL            string
	httpMockServer *httptest.Server
	expectations   []*Expectation
	counter        int
	suite          SuiteLike
}

type SuiteLike interface {
	Equal(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool
	Failf(failureMessage string, msg string, args ...interface{}) bool
	NoError(err error, msgAndArgs ...interface{}) bool
	JSONEq(expected string, actual string, msgAndArgs ...interface{}) bool
}

func NewMockServer(suite SuiteLike, expectations ...*Expectation) *MockServer {
	mockServer := &MockServer{
		suite:        suite,
		expectations: expectations,
		counter:      0,
	}
	httpMockServer := httptest.NewServer(mockServer.handlerFunc())
	mockServer.httpMockServer = httpMockServer
	mockServer.URL = httpMockServer.URL
	return mockServer
}

func (ms *MockServer) Close() {
	ms.httpMockServer.Close()
}

func (ms *MockServer) HasExpectedRequests() {
	ms.suite.Equal(len(ms.expectations), ms.counter, "invalid amount of requests have been made")
}

func (ms *MockServer) handlerFunc() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ms.counter >= len(ms.expectations) {
			ms.suite.Failf("invalid amount of requests", "unknown request received: %s %s (%d)", r.Method, r.RequestURI, ms.counter)
			return
		}
		expectation := ms.expectations[ms.counter]
		requestNumber := fmt.Sprintf("request #%d", ms.counter+1)
		ms.suite.Equal(expectation.request.method, r.Method, "invalid request method", requestNumber)
		ms.suite.Equal(expectation.request.path, r.URL.Path, "invalid request path", requestNumber)

		queryValues := r.URL.Query()
		for key, value := range expectation.request.queryParameters {
			ms.suite.Equal(value, queryValues.Get(key), "invalid query value", requestNumber)
		}

		for key, value := range expectation.request.headers {
			ms.suite.Equal(value, r.Header.Get(key), "invalid header value", requestNumber)
		}

		if expectation.request.jsonParameters != nil {
			expectedJSONPayload, err := json.Marshal(expectation.request.jsonParameters)
			ms.suite.NoError(err, requestNumber)

			bodyBytes, err := ioutil.ReadAll(r.Body)
			ms.suite.NoError(err, requestNumber)

			ms.suite.JSONEq(string(expectedJSONPayload), string(bodyBytes), requestNumber)
		}

		if expectation.request.body != nil {
			bodyBytes, err := ioutil.ReadAll(r.Body)
			ms.suite.NoError(err, requestNumber)

			ms.suite.Equal(*expectation.request.body, bodyBytes, requestNumber)
		}

		if expectation.response.code > 0 {
			w.WriteHeader(expectation.response.code)
		}

		for key, value := range expectation.response.headers {
			w.Header().Add(key, value)
		}

		if len(expectation.response.body) > 0 {
			w.Write(expectation.response.body)
		}

		ms.counter = ms.counter + 1
	})
}
