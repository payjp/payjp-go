package payjp

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func NewMockClient(status int, response []byte) (*http.Client, *MockTransport) {
	transport := &MockTransport{
		responses: []*responsePair{&responsePair{status, response}},
	}
	return &http.Client{
		Transport: transport,
	}, transport
}

type responsePair struct {
	status   int
	response []byte
}

type MockTransport struct {
	responses []*responsePair
	index     int
	URL       string
	Method    string
}

// Implement http.RoundTripper
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.URL = req.URL.String()
	t.Method = req.Method
	// Create mocked http.Response
	responseSet := t.responses[t.index]
	t.index++
	if t.index == len(t.responses) {
		t.index--
	}
	response := &http.Response{
		Header:     make(http.Header),
		Request:    req,
		StatusCode: responseSet.status,
	}
	response.Body = ioutil.NopCloser(bytes.NewReader(responseSet.response))
	return response, nil
}

func (t *MockTransport) AddResponse(status int, body []byte) {
	t.responses = append(t.responses, &responsePair{
		status:   status,
		response: body,
	})
}
