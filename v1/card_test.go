package payjp

import (
	"encoding/json"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

var cardResponseJSON = []byte(`
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
  "three_d_secure_status": null,
  "object": "card"
}
`)

var cardListResponseJSON = []byte(`
{
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
      "three_d_secure_status": "verified",
      "object": "card"
    }
  ],
  "object": "list",
  "has_more": true,
  "url": "/v1/customers/cus_4df4b5ed720933f4fb9e28857517/cards"
}
`)

var cardDeleteResponseJSON = []byte(`
{
  "deleted": true,
  "id": "car_f7d9fa98594dc7c2e42bfcd641ff",
  "livemode": false
}
`)

func TestParseCardResponseJSON(t *testing.T) {
	card := &CardResponse{}
	err := json.Unmarshal(cardResponseJSON, card)

	assert.NoError(t, err)
	assert.Equal(t, "car_f7d9fa98594dc7c2e42bfcd641ff", card.ID)
	assert.Nil(t, card.ThreeDSecureStatus)

	createdAt := card.CreatedAt.UTC().Format("2006-01-02 15:04:05")
	if createdAt != "2015-06-01 03:06:23" {
		t.Errorf("card.CreatedAt() should be '2015-06-01 03:06:23' but '%s'", createdAt)
	}
}

func TestCustomerAddCard(t *testing.T) {
	mock, transport := newMockClient(200, cardResponseJSON)
	service := New("api-key", mock)
	card, err := service.Customer.AddCard("cus_121673955bd7aa144de5a8f6c262", Card{
		Number:   "4242424242424242",
		ExpMonth: 2,
		ExpYear:  2020,
	})
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if card == nil {
		t.Error("card should not be nil")
	} else if card.Last4 != "4242" {
		t.Errorf("card.Last4 should be '4242', but %s.", card.Last4)
	}
}

func TestCustomerAddCard2(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	if customer == nil {
		t.Error("plan should not be nil")
		return
	}
	card, err := customer.AddCard(Card{
		Number:   "4242424242424242",
		ExpMonth: "2",
		ExpYear:  "2020",
		CVC:      "000",
	})
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if card == nil {
		t.Error("card should not be nil")
	} else if card.Last4 != "4242" {
		t.Errorf("card.Last4 should be '4242', but %s.", card.Last4)
	}
}

func TestCustomerGetCard(t *testing.T) {
	mock, transport := newMockClient(200, cardResponseJSON)
	service := New("api-key", mock)
	card, err := service.Customer.GetCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff")
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if card == nil {
		t.Error("card should not be nil")
	} else if card.Last4 != "4242" {
		t.Errorf("parse error: card.Last4 should be 4242, but %s.", card.Last4)
	}
}

func TestCustomerGetCard2(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	if customer == nil {
		t.Error("plan should not be nil")
		return
	}
	card, err := customer.GetCard("car_f7d9fa98594dc7c2e42bfcd641ff")
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if card == nil {
		t.Error("card should not be nil")
	} else if card.Last4 != "4242" {
		t.Errorf("parse error: card.Last4 should be 4242, but %s.", card.Last4)
	}
}

func TestCustomerUpdateCard(t *testing.T) {
	mock, transport := newMockClient(200, cardResponseJSON)
	service := New("api-key", mock)
	card, err := service.Customer.UpdateCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff", Card{
		Number:   "4242424242424242",
		ExpMonth: 2,
		ExpYear:  2020,
	})
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if card == nil {
		t.Error("card should not be nil")
	} else if card.Last4 != "4242" {
		t.Errorf("parse error: card.Last4 should be 4242, but %s.", card.Last4)
	}
}

func TestCustomerUpdateCard2(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
    transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, customer)

	err = customer.Cards[0].Update(Card{
		ExpMonth: "2",
		ExpYear:  "2025",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card[exp_month]=2&card[exp_year]=2025", *transport.Body)

	card, err := service.Customer.UpdateCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff", Card{
		ExpMonth: 2,
		ExpYear:  2020,
	})
	assert.Nil(t, card)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerDeleteCard(t *testing.T) {
	mock, transport := newMockClient(200, cardDeleteResponseJSON)
	service := New("api-key", mock)
	err := service.Customer.DeleteCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff")
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "DELETE" {
		t.Errorf("Method should be DELETE, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
}

func TestCustomerDeleteCard2(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardDeleteResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	if customer == nil {
		t.Error("card should not be nil")
		return
	}
	err = customer.Cards[0].Delete()
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "DELETE" {
		t.Errorf("Method should be DELETE, but %s", transport.Method)
	}
}

func TestCustomerListCard(t *testing.T) {
	mock, transport := newMockClient(200, cardListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	cards, hasMore, err := service.Customer.ListCard("cus_121673955bd7aa144de5a8f6c262").
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	assert.NoError(t, err)
    assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards?limit=10&offset=15&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(cards), 1)
	assert.Equal(t, "verified", *cards[0].ThreeDSecureStatus)

	_, hasMore, err = service.Customer.ListCard("cus_121673955bd7aa144de5a8f6c262").Do()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
