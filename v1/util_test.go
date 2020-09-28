package payjp

import (
	"bytes"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"testing"
	"testing/quick"
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


// test for RandUniform
type randUniformArgSource struct {
  valueX float64
  valueY float64
}

func(s randUniformArgSource) maxValue() float64 {
  return math.Max(s.valueX, s.valueY)
}

func(s randUniformArgSource) minValue() float64 {
  return math.Min(s.valueX, s.valueY)
}

func (randUniformArgSource) Generate(r *rand.Rand, size int) reflect.Value {
  s := randUniformArgSource {
    float64(r.Int63()),
    float64(r.Int63()),
  }
  return reflect.ValueOf(s)
}


func TestRandUniform(t *testing.T) {
  testFunc := func(arg randUniformArgSource) bool {
    min := arg.minValue()
    max := arg.maxValue()
    result:= RandUniform(min, max)
    return result >= min || result < max
  }
  if err := quick.Check(testFunc, nil); err != nil {
    t.Error(err)
  }
}
