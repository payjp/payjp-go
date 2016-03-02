package payjp

import (
	"encoding/json"
	"testing"
	"time"
)

var customerResponseJSON = []byte(`
{
  "cards": {
    "count": 1,
    "data": [
      {
        "address_city": null,
        "address_line1": null,
        "address_line2": null,
        "address_state": null,
        "address_zip": null,
        "address_zip_check": "unchecked",
        "brand": "Visa",
        "country": null,
        "created": 1433127983,
        "cvc_check": "unchecked",
        "exp_month": 2,
        "exp_year": 2020,
        "fingerprint": "e1d8225886e3a7211127df751c86787f",
        "id": "car_f7d9fa98594dc7c2e42bfcd641ff",
        "last4": "4242",
        "livemode": false,
        "name": null,
        "object": "card"
      }
    ],
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
    "data": [
      {
        "canceled_at": null,
        "created": 1433127983,
        "current_period_end": 1435732422,
        "current_period_start": 1433140422,
        "customer": "cus_4df4b5ed720933f4fb9e28857517",
        "id": "sub_567a1e44562932ec1a7682d746e0",
        "livemode": false,
        "object": "subscription",
        "paused_at": null,
        "plan": {
          "amount": 1000,
          "billing_day": null,
          "created": 1432965397,
          "currency": "jpy",
          "id": "pln_9589006d14aad86aafeceac06b60",
          "interval": "month",
          "name": "test plan",
          "object": "plan",
          "trial_days": 0
        },
        "resumed_at": null,
        "start": 1433140422,
        "status": "active",
        "trial_end": null,
        "trial_start": null,
        "prorate": false
      }
    ],
    "has_more": false,
    "object": "list",
    "url": "/v1/customers/cus_121673955bd7aa144de5a8f6c262/subscriptions"
  }
}
`)

var customerErrorResponseJSON = []byte(`
{
  "error": {
    "code": "invalid_param_key",
    "message": "Invalid param key to customer.",
    "param": "dummy",
    "status": 400,
    "type": "client_error"
  }
}
`)

var customerListResponseJSON = []byte(`
{
  "count": 1,
  "data": [
    {
      "cards": {
        "count": 0,
        "data": [],
        "has_more": false,
        "object": "list",
        "url": "/v1/customers/cus_842e21be700d1c8156d9dac025f6/cards"
      },
      "created": 1433059905,
      "default_card": null,
      "description": "test",
      "email": null,
      "id": "cus_842e21be700d1c8156d9dac025f6",
      "livemode": false,
      "object": "customer",
      "subscriptions": {
        "count": 0,
        "data": [],
        "has_more": false,
        "object": "list",
        "url": "/v1/customers/cus_842e21be700d1c8156d9dac025f6/subscriptions"
      }
    }
  ],
  "has_more": true,
  "object": "list",
  "url": "/v1/customers"
}
`)

func TestParseCustomerResponseJson(t *testing.T) {
	customer := &CustomerResponse{}
	err := json.Unmarshal(customerResponseJSON, customer)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if customer.ID != "cus_121673955bd7aa144de5a8f6c262" {
		t.Errorf("customer.ID should be 'cus_121673955bd7aa144de5a8f6c262', but '%s'", customer.ID)
	}
}

func TestCustomerCreate(t *testing.T) {
	mock, transport := NewMockClient(200, customerResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Create(Customer{
		Description: "test",
	})
	if transport.URL != "https://api.pay.jp/v1/customers" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if customer == nil {
		t.Error("customer should not be nil")
	} else if customer.Description != "test" {
		t.Errorf("customer.Description should be 'test', but %s.", customer.Description)
	}
}

func TestCustomerCreateError(t *testing.T) {
	mock, _ := NewMockClient(400, customerErrorResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Create(Customer{
		Description: "test",
	})
	if err == nil {
		t.Error("err should not be nil")
	}
	if customer != nil {
		t.Errorf("customer should be nil, but %v", customer)
	}
}

func TestCustomerRetrieve(t *testing.T) {
	mock, transport := NewMockClient(200, customerResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if customer == nil {
		t.Error("customer should not be nil")
	} else if customer.Description != "test" {
		t.Errorf("parse error: customer.Description should be 500, but %s.", customer.Description)
	} else if len(customer.Cards) != 1 {
		t.Errorf("parse error: customer.Cards should have 1 card, but %d cards.", len(customer.Cards))
	}
}

func TestCustomerUpdate(t *testing.T) {
	mock, transport := NewMockClient(200, customerResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Update("cus_121673955bd7aa144de5a8f6c262", Customer{
		Email: "test@mail.com",
	})
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if customer == nil {
		t.Error("plan should not be nil")
	} else if customer.Description != "test" {
		t.Errorf("parse error: customer.Description should be 500, but %s.", customer.Description)
	}
}

func TestCustomerUpdate2(t *testing.T) {
	mock, transport := NewMockClient(200, customerResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	if plan == nil {
		t.Error("plan should not be nil")
		return
	}
	err = plan.Update(Customer{
		Email: "test@mail.com",
	})
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
}

func TestCustomerDelete(t *testing.T) {
	mock, transport := NewMockClient(200, customerResponseJSON)
	service := New("api-key", mock)
	err := service.Customer.Delete("cus_121673955bd7aa144de5a8f6c262")
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "DELETE" {
		t.Errorf("Method should be DELETE, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
}

func TestCustomerDelete2(t *testing.T) {
	mock, transport := NewMockClient(200, customerResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	if customer == nil {
		t.Error("plan should not be nil")
		return
	}
	err = customer.Delete()
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "DELETE" {
		t.Errorf("Method should be DELETE, but %s", transport.Method)
	}
}

func TestCustomerList(t *testing.T) {
	mock, transport := NewMockClient(200, customerListResponseJSON)
	service := New("api-key", mock)
	plans, hasMore, err := service.Customer.List().
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/customers?limit=10&offset=15&since=1455328095&until=1455500895" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if !hasMore {
		t.Error("parse error: hasMore")
	}
	if len(plans) != 1 {
		t.Error("parse error: plans")
	}
}
