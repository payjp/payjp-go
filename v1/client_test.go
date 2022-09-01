package payjp

import (
	"net/http"
	"testing"

    "github.com/stretchr/testify/assert"
)

type TestListParams struct {
	service *Service `form:"-"`
	Param *string `form:"param"`
}

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
	mock, transport := NewMockClient(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	service := New("api-key", mock)
	qb := newRequestBuilder()

	body, err := service.request("POST", "/test", qb.Reader())
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/test", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, errorJSONStr, string(body))

	_, err = service.retrieve("/test")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/test", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))

	err = service.delete("/test")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/test", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
}

func TestGetList(t *testing.T) {
	mock, transport := NewMockClient(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	service := New("api-key", mock)

	l := &TestListParams{}
	body, err := service.getList("/test", l)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/test", transport.URL)
	assert.Equal(t, errorJSONStr, string(body))

	str := "str"
	l2 := &TestListParams{
		Param: &str,
	}
	_, err = service.getList("/test", l2)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/test?param=str", transport.URL)
}
