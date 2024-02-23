package responses

import (
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MockSuite struct {
	suite  *suite.Suite
	errors []string
}

func NewMockSuite(suite *suite.Suite) *MockSuite {
	return &MockSuite{
		suite:  suite,
		errors: make([]string, 0),
	}
}

func (ms *MockSuite) Equal(expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert.Equal(ms, expected, actual, msgAndArgs...)
}

func (ms *MockSuite) Failf(failureMessage string, msg string, args ...interface{}) bool {
	return assert.Failf(ms, failureMessage, msg, args...)
}

func (ms *MockSuite) NoError(err error, msgAndArgs ...interface{}) bool {
	return assert.NoError(ms, err, msgAndArgs...)
}

func (ms *MockSuite) JSONEq(expected string, actual string, msgAndArgs ...interface{}) bool {
	return assert.JSONEq(ms, expected, actual, msgAndArgs...)
}

func (ms *MockSuite) Errorf(format string, args ...interface{}) {
	ms.errors = append(ms.errors, fmt.Sprintf(format, args...))
}

func (ms *MockSuite) HasErrors(num int) {
	ms.suite.Len(ms.errors, num)
}

func (ms *MockSuite) ErrorContains(index int, contains interface{}) {
	ms.suite.Contains(ms.errors[index], contains)
}
