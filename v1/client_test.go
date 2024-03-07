package payjp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type TestListParams struct {
	Param *string `form:"param"`
}

var errorJSONStr = `
{
  "code": "code",
  "message": "message",
  "param": "param",
  "status": 400,
  "type": "type"
}
`

var errorJSON = []byte(errorJSONStr)

func TestNew(t *testing.T) {
	// New(): default constructor
	service := New("api-key", nil)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Client)

	assert.NotNil(t, service.Charge)
	assert.NotNil(t, service.Customer)
	assert.NotNil(t, service.Plan)
	assert.NotNil(t, service.Subscription)
	assert.NotNil(t, service.Account)
	assert.NotNil(t, service.Token)
	assert.NotNil(t, service.Transfer)
	assert.NotNil(t, service.Event)
	assert.Equal(t, "https://api.pay.jp/v1", service.apiBase)
	assert.Regexp(t, "^Basic .*", service.apiKey)

	client := &http.Client{}
	assert.NotSame(t, client, service.Client)

	service = New("api-key", client)
	assert.Same(t, client, service.Client)
}

func TestAPIBase(t *testing.T) {
	service := New("api-key", nil, Config{
		APIBase: "https://api.pay.jp/v2",
	})
	assert.Equal(t, "https://api.pay.jp/v2", service.APIBase())
}

func TestRequests(t *testing.T) {
	mock, transport := newMockClient(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	service := New("api-key", mock)
	qb := newRequestBuilder()

	body, err := service.request("POST", "/test", qb.Reader())
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/test", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Regexp(t, "^Go-http-client/payjp-v[0-9.]+$", transport.Header.Get("User-Agent"))
	assert.Regexp(t, "^payjp-go/v[0-9.]+\\(go[0-9.]+,os\\:.+,arch\\:.+\\)$", transport.Header.Get("X-Payjp-Client-User-Agent"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, errorJSONStr, string(body))
}

func TestGetList(t *testing.T) {
	mock, transport := newMockClient(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	service := New("api-key", mock)

	l := &TestListParams{}
	q := service.getQuery(l)
	assert.Equal(t, "", q)

	str := "str"
	l2 := &TestListParams{
		Param: &str,
	}
	q = service.getQuery(l2)
	assert.Equal(t, "?param=str", q)
}
