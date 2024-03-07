package payjp

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func newMockClient(status int, response []byte) (*http.Client, *mockTransport) {
	transport := &mockTransport{
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

type mockTransport struct {
	responses []*responsePair
	index     int
	URL       string
	Method    string
	Body      *string
	Header    http.Header
}

// Implement http.RoundTripper
func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.URL = req.URL.String()
	t.Method = req.Method
	t.Header = req.Header
	t.Body = nil
	if req.Body != nil {
		b, _ := ioutil.ReadAll(req.Body)
		s := string(b)
		t.Body = &s
	}
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

func (t *mockTransport) AddResponse(status int, body []byte) {
	t.responses = append(t.responses, &responsePair{
		status:   status,
		response: body,
	})
}
