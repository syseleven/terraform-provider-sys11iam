package responses

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HTTPResponsesTestSuite struct {
	suite.Suite
}

func (suite *HTTPResponsesTestSuite) TestRequestAmount() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/ips"),
	)
	defer mockServer.Close()

	for i := 1; i < 3; i++ {
		_, err := http.Get(mockServer.URL + "/ips")
		suite.NoError(err)
	}
	mockSuite.HasErrors(1)
	mockSuite.ErrorContains(0, `invalid amount of requests`)
	mockSuite.ErrorContains(0, `unknown request received: GET /ips (1)`)
}

func (suite *HTTPResponsesTestSuite) TestRequestMethod() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/ips"),
	)
	defer mockServer.Close()

	_, err := http.Post(mockServer.URL+"/ips", "", nil)
	suite.NoError(err)
	mockSuite.HasErrors(1)
	mockSuite.ErrorContains(0, `expected: "GET"`)
	mockSuite.ErrorContains(0, `actual  : "POST"`)
	mockSuite.ErrorContains(0, `request #1`)
}

func (suite *HTTPResponsesTestSuite) TestRequestPath() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/ips"),
	)
	defer mockServer.Close()

	_, err := http.Get(mockServer.URL + "/ip")
	suite.NoError(err)
	mockSuite.HasErrors(1)
	mockSuite.ErrorContains(0, `expected: "/ips"`)
	mockSuite.ErrorContains(0, `actual  : "/ip"`)
	mockSuite.ErrorContains(0, `request #1`)
}

func (suite *HTTPResponsesTestSuite) TestRequestQueryParameters() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/ips").
			WithQueryParameters(map[string]string{
				"foo": "bar",
			}),
	)
	defer mockServer.Close()

	_, err := http.Get(mockServer.URL + "/ips?foo=nobar")
	suite.NoError(err)
	mockSuite.HasErrors(1)
	mockSuite.ErrorContains(0, `expected: "bar"`)
	mockSuite.ErrorContains(0, `actual  : "nobar"`)
	mockSuite.ErrorContains(0, `request #1`)
}

func (suite *HTTPResponsesTestSuite) TestRequestHeaders() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/").
			WithHeaders(map[string]string{
				"foo": "bar",
			}),
		Expect(http.MethodGet, "/").
			WithHeaders(map[string]string{
				"foo": "bar",
			}),
	)
	defer mockServer.Close()

	// first request with wrong header value
	req, err := http.NewRequest(http.MethodGet, mockServer.URL, nil)
	suite.NoError(err)
	req.Header.Add("foo", "baz")
	_, err = http.DefaultClient.Do(req)
	suite.NoError(err)

	// should error
	mockSuite.HasErrors(1)
	mockSuite.ErrorContains(0, `expected: "bar"`)
	mockSuite.ErrorContains(0, `actual  : "baz"`)
	mockSuite.ErrorContains(0, `request #1`)

	// second request with right header value
	req.Header.Set("foo", "bar")
	_, err = http.DefaultClient.Do(req)
	suite.NoError(err)

	// no second error
	mockSuite.HasErrors(1)
}

func (suite *HTTPResponsesTestSuite) TestRequestJSONParameters() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodPost, "/ips").
			WithJSONParameters(struct {
				Foo string `json:"foo"`
			}{
				"bar",
			}),
	)
	defer mockServer.Close()

	jsonPayload := []byte(`{"foo":"nobar"}`)
	_, err := http.Post(mockServer.URL+"/ips", "application/json", bytes.NewReader(jsonPayload))
	suite.NoError(err)
	mockSuite.HasErrors(1)
	mockSuite.ErrorContains(0, `expected: map[string]interface {}{"foo":"bar"}`)
	mockSuite.ErrorContains(0, `actual  : map[string]interface {}{"foo":"nobar"}`)
	mockSuite.ErrorContains(0, `request #1`)
}

func (suite *HTTPResponsesTestSuite) TestRequestBody() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodPost, "/ips").
			WithBody([]byte("foo")),
	)
	defer mockServer.Close()

	body := []byte(`bar`)
	_, err := http.Post(mockServer.URL+"/ips", "application/json", bytes.NewReader(body))
	suite.NoError(err)
	mockSuite.HasErrors(1)
	mockSuite.ErrorContains(0, `expected: []byte{0x66, 0x6f, 0x6f}`)
	mockSuite.ErrorContains(0, `actual  : []byte{0x62, 0x61, 0x72}`)
	mockSuite.ErrorContains(0, `request #1`)
}

func (suite *HTTPResponsesTestSuite) TestResponseCode() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/ips").
			ReturnWithCode(http.StatusAccepted),
	)
	defer mockServer.Close()

	resp, err := http.Get(mockServer.URL + "/ips")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
	mockSuite.HasErrors(0)
	suite.Equal(http.StatusAccepted, resp.StatusCode)
}

func (suite *HTTPResponsesTestSuite) TestResponseHeader() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		// setting headers alone
		Expect(http.MethodGet, "/").
			ReturnWithHeaders(map[string]string{
				"foo": "bar",
				"baz": "123",
			}),
		// setting headers after JSON response
		Expect(http.MethodGet, "/").
			ReturnWithJSONFile("testfile.json").
			ReturnWithHeaders(map[string]string{
				"foo": "bar",
				"baz": "123",
			}),
		// setting headers before JSON response
		Expect(http.MethodGet, "/").
			ReturnWithHeaders(map[string]string{
				"foo": "bar",
				"baz": "123",
			}).
			ReturnWithJSONFile("testfile.json"),
	)
	defer mockServer.Close()

	for range []int{1, 2, 3} {
		resp, err := http.Get(mockServer.URL + "/")
		suite.NoError(err)
		mockSuite.HasErrors(0)
		suite.Equal("bar", resp.Header.Get("foo"))
		suite.Equal("123", resp.Header.Get("baz"))
	}
	mockServer.HasExpectedRequests()
}

func (suite *HTTPResponsesTestSuite) TestResponseBody() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/ips").
			ReturnWithBody([]byte("foobar")),
	)
	defer mockServer.Close()

	resp, err := http.Get(mockServer.URL + "/ips")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
	mockSuite.HasErrors(0)
	body, err := ioutil.ReadAll(resp.Body)
	suite.NoError(err)
	suite.Equal([]byte("foobar"), body)
}

func (suite *HTTPResponsesTestSuite) TestResponseJSONFile() {
	mockSuite := NewMockSuite(&suite.Suite)
	mockServer := NewMockServer(
		mockSuite,
		Expect(http.MethodGet, "/ips").
			ReturnWithJSONFile("testfile.json"),
	)
	defer mockServer.Close()

	resp, err := http.Get(mockServer.URL + "/ips")
	suite.NoError(err)
	mockServer.HasExpectedRequests()
	mockSuite.HasErrors(0)
	suite.Equal("application/json", resp.Header.Get("Content-Type"))
	body, err := ioutil.ReadAll(resp.Body)
	suite.NoError(err)
	suite.JSONEq(`{"foo": "bar"}`, string(body))
}

func TestHTTPResponsesTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPResponsesTestSuite))
}
