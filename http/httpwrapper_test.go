package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math"
	. "net/http"
	"testing"
	"time"
)

type sampleData struct {
	SampleString string `json:"sampleString"`
}

func TestNoProxy_NewHttpWrapper(t *testing.T) {
	httpWrapper, err := NewHttpWrapper(120, "", "username", "password")
	assert.NoError(t, err, "No error should be returned")
	assert.NotNil(t, httpWrapper, "HttpWrapper should be returned")
	assert.Nil(t, httpWrapper.httpClient.Transport, "HTTP transport should be nil")
	assert.NotNil(t, httpWrapper.httpClient, "IHttpClient should be set")
	assert.Equal(t, time.Second*time.Duration(120), httpWrapper.httpClient.Timeout)
}

func TestProxy_NewHttpWrapper(t *testing.T) {
	httpWrapper, err := NewHttpWrapper(120, "http://proxy:8080", "username", "password")
	assert.NoError(t, err, "No error should be returned")
	assert.NotNil(t, httpWrapper, "HttpWrapper should be returned")
	assert.NotNil(t, httpWrapper.httpClient.Transport, "HTTP transport should be set")
	assert.NotNil(t, httpWrapper.httpClient, "IHttpClient should be set")
	assert.Equal(t, time.Second*time.Duration(120), httpWrapper.httpClient.Timeout)
}

func TestInvalidProxy_NewHttpWrapper(t *testing.T) {
	httpWrapper, err := NewHttpWrapper(120, "://proxy:8080", "username", "password")
	assert.Error(t, err, "No error should be returned")
	assert.Nil(t, httpWrapper, "HttpWrapper should be nil")
}

func TestMarshallError_ExecuteRequest(t *testing.T) {
	// This test doesn't need the http client
	httpWrapper := HttpWrapper{}

	// Send in unmarshable body data
	statusCode, resultStr, err := httpWrapper.ExecuteRequest("POST", "https://www.google.com", math.Inf(1), nil)
	assert.Equal(t, 0, statusCode)
	assert.Equal(t, "", resultStr)
	assert.Contains(t, err.Error(), "marshalling")
}

func TestInvalidMethod_ExecuteRequest(t *testing.T) {
	// This test doesn't need the http client
	httpWrapper := HttpWrapper{}

	// Send in invalid HTTP method
	statusCode, resultStr, err := httpWrapper.ExecuteRequest(":GOOBLEGOOK", "https://www.google.com", nil, nil)
	assert.Equal(t, 0, statusCode)
	assert.Equal(t, "", resultStr)
	assert.Contains(t, err.Error(), "building request")
}

func TestNoResult_ExecuteRequest(t *testing.T) {
	client := NewTestClient(func(request *Request) *Response {
		// Test request parameters
		assert.Equal(t, "PUT", request.Method)
		assert.Equal(t, "https://www.google.com", request.URL.String())
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))

		return &Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(Header),
		}
	})

	// This test doesn't need the http client
	httpWrapper := HttpWrapper{
		client,
		"user",
		"pwd",
	}

	// Send in invalid HTTP method
	statusCode, resultStr, err := httpWrapper.ExecuteRequest("PUT", "https://www.google.com", sampleData{"data"}, nil)
	assert.Equal(t, 200, statusCode, "Invalid status code returned")
	assert.Equal(t, "OK", resultStr, "Invalid result string")
	assert.NoError(t, err, "There should be no error")
}

func TestReaderError_ExecuteRequest(t *testing.T) {
	client := NewTestClient(func(request *Request) *Response {
		// Test request parameters
		assert.Equal(t, "PUT", request.Method)
		assert.Equal(t, "https://www.google.com", request.URL.String())
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))

		response := &Response{

			StatusCode: 200,
			// Send response to be tested
			Body: errReaderCloser{},
			// Must be set to non-nil value or it panics
			Header: make(Header),
		}
		_ = response.Body.Close()
		return response
	})

	// This test doesn't need the http client
	httpWrapper := HttpWrapper{
		client,
		"user",
		"pwd",
	}

	// Send in invalid HTTP method
	statusCode, resultStr, err := httpWrapper.ExecuteRequest("PUT", "https://www.google.com", sampleData{"data"}, nil)
	assert.Equal(t, 0, statusCode, "Invalid status code returned")
	assert.Equal(t, "", resultStr, "Invalid result string")
	assert.Error(t, err, "There should be no error")
	assert.Contains(t, err.Error(), "test read error", "Some error")
}

func TestWithResult_ExecuteRequest(t *testing.T) {
	expectedOutput := &sampleData{"output"}

	client := NewTestClient(func(request *Request) *Response {
		// Test request parameters
		assert.Equal(t, "PUT", request.Method)
		assert.Equal(t, "https://www.google.com", request.URL.String())
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		sampleDataBuf, _ := json.Marshal(expectedOutput)

		return &Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(string(sampleDataBuf))),
			// Must be set to non-nil value or it panics
			Header: make(Header),
		}
	})

	// This test doesn't need the http client
	httpWrapper := HttpWrapper{
		client,
		"user",
		"pwd",
	}

	sampleInput := &sampleData{"input"}
	expectedOutputBytes, _ := json.Marshal(expectedOutput)
	expectedOutputStr := string(expectedOutputBytes)
	var resultData sampleData

	// Send in invalid HTTP method
	statusCode, resultStr, err := httpWrapper.ExecuteRequest("PUT", "https://www.google.com", sampleInput, &resultData)
	assert.Equal(t, 200, statusCode, "Invalid status code returned")
	assert.Equal(t, expectedOutputStr, resultStr, "Invalid result string")
	assert.NoError(t, err, "There should be no error")
}

// test: error unmarshalling result
func TestWithUnmarshallingResultError_ExecuteRequest(t *testing.T) {
	expectedOutput := &sampleData{"output"}

	client := NewTestClient(func(request *Request) *Response {
		// Test request parameters
		assert.Equal(t, "PUT", request.Method)
		assert.Equal(t, "https://www.google.com", request.URL.String())
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		sampleDataBuf, _ := json.Marshal(expectedOutput)

		return &Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(string(sampleDataBuf))),
			// Must be set to non-nil value or it panics
			Header: make(Header),
		}
	})

	// This test doesn't need the http client
	httpWrapper := HttpWrapper{
		client,
		"user",
		"pwd",
	}

	type wrongType struct {
		SampleString int `json:"sampleString"`
	}

	sampleInput := &sampleData{"input"}
	expectedOutputBytes, _ := json.Marshal(expectedOutput)
	expectedOutputStr := string(expectedOutputBytes)
	var resultData wrongType

	// Send in invalid HTTP method
	statusCode, resultStr, err := httpWrapper.ExecuteRequest("PUT", "https://www.google.com", sampleInput, &resultData)
	assert.Equal(t, 200, statusCode, "Invalid status code returned")
	assert.Equal(t, expectedOutputStr, resultStr, "Invalid result string")
	assert.Error(t, err, "There should be an error unmarshalling the result")
}

// RoundTripFunc .
type RoundTripFunc func(req *Request) *Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *Request) (*Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *Client {
	return &Client{
		Transport: RoundTripFunc(fn),
	}
}

type errReaderCloser struct{}

func (errReaderCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("test read error")
}

func (errReaderCloser) Close() error {
	return fmt.Errorf("test close error")
}
