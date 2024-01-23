package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var customerResponseJSONStr = `
{
  "cards": {
    "count": 1,
    "data": [` + cardResponseJSONStr +
	`],
    "has_more": false,
    "object": "list",
    "url": "/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards"
  },
  "created": 1433127983,
  "default_card": null,
  "description": "test",
  "email": null,
  "id": "cus_121673955bd7aa144de5a8f6c262",
  "livemode": false,
  "object": "customer",
  "subscriptions": {
    "count": 1,
    "data": [` + subscriptionResponseJSONStr +
	`],
    "has_more": false,
    "object": "list",
    "url": "/v1/customers/cus_121673955bd7aa144de5a8f6c262/subscriptions"
  }
}`
var customerResponseJSON = []byte(customerResponseJSONStr)

var customerListResponseJSON = []byte(`
{
  "count": 1,
  "data": [` + customerResponseJSONStr +
	`],
  "has_more": true,
  "object": "list",
  "url": "/v1/customers"
}
`)
var customerDeleteResponseJSON = []byte(deleteResponseJSONStr)

func TestParseCustomerResponseJson(t *testing.T) {
	service := &Service{}
	customer := &CustomerResponse{
		service: service,
	}
	err := json.Unmarshal(customerResponseJSON, customer)

	assert.NoError(t, err)
	assert.Equal(t, "cus_121673955bd7aa144de5a8f6c262", customer.ID)
	assert.Equal(t, 1, len(customer.Cards))
	assert.Equal(t, "car_f7d9fa98594dc7c2e42bfcd641ff", customer.Cards[0].ID)
	assert.Equal(t, service, customer.Cards[0].service)
	assert.Equal(t, 1, len(customer.Subscriptions))
	assert.Equal(t, "sub_response1", customer.Subscriptions[0].ID)
	assert.Equal(t, service, customer.Subscriptions[0].service)
	assert.False(t, customer.LiveMode)
	assert.Equal(t, 1433127983, *customer.Created)
	assert.IsType(t, time.Unix(0, 0), customer.CreatedAt)
	assert.Equal(t, "", customer.Email)
	assert.Equal(t, map[string]string(nil), customer.Metadata)
	assert.Equal(t, "cus_121673955bd7aa144de5a8f6c262", customer.Cards[0].customerID)
	assert.Equal(t, service, customer.service)
}

func TestCustomerCreate(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, customerResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	customer, err := service.Customer.Create(Customer{
		Email:       "example@example.com",
		Description: "test",
		ID:          "test",
		CardToken:   "tok_xxx",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "email=example%40example.com&description=test&id=test&card=tok_xxx", *transport.Body)
	assert.NotNil(t, customer)
	assert.Equal(t, "4242", customer.Cards[0].Last4)

	customer, err = service.Customer.Create(Customer{
		Email:       String("example@example.com"),
		Description: String("test"),
		ID:          String("test"),
		CardToken:   String("tok_xxx"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "email=example%40example.com&description=test&id=test&card=tok_xxx", *transport.Body)
	assert.NotNil(t, customer)

	_, err = service.Customer.Create(Customer{
		Email:       String("error"),
		Description: nil,
	})
	assert.Equal(t, "email=error", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "test", customer.Description)

	_, err = service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerUpdate(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	customer, err := service.Customer.Update("cus_121673955bd7aa144de5a8f6c262", Customer{
		Email: String("test@mail.com"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "email=test%40mail.com", *transport.Body)
	assert.NotNil(t, customer)
	assert.Equal(t, "4242", customer.Cards[0].Last4)
}

func TestCustomerResponseUpdate(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.NotNil(t, customer)
	err = customer.Update(Customer{
		Email: "test@mail.com",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "email=test%40mail.com", *transport.Body)
	assert.NotNil(t, customer)
	assert.Equal(t, "4242", customer.Cards[0].Last4)
}

func TestCustomerDelete(t *testing.T) {
	mock, transport := newMockClient(200, customerDeleteResponseJSON)
	transport.AddResponse(200, customerResponseJSON)
	transport.AddResponse(200, customerDeleteResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	err := service.Customer.Delete("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))

	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	err = customer.Delete()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))

	err = service.Customer.Delete("cus_121673955bd7aa144de5a8f6c262")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerList(t *testing.T) {
	mock, transport := newMockClient(200, customerListResponseJSON)
	transport.AddResponse(200, customerListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	c, hasMore, err := service.Customer.List().
		Limit(10).
		Offset(0).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers?limit=10&offset=0&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(c), 1)
	assert.Equal(t, service, c[0].service)

	params := &CustomerListParams{
		ListParams: ListParams{
			Limit: Int(10),
			Since: Int(1455328095),
			Until: Int(1455500895),
		},
	}
	params.Offset = Int(0)
	c, hasMore, err = service.Customer.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers?limit=10&offset=0&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(c), 1)
	assert.Equal(t, service, c[0].service)

	_, hasMore, err = service.Customer.All()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
