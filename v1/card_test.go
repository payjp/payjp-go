package payjp

import (
	"encoding/json"
	"testing"
	"time"
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

var cardErrorResponseJSON = []byte(`
{
  "error": {
    "message": "There is no card with ID: dummy",
    "param": "id",
    "status": 404,
    "type": "client_error"
  }
}
`)

func TestParseCardResponseJSON(t *testing.T) {
	card := &CardResponse{}
	json.Unmarshal(cardResponseJSON, card)

	if card.ID != "car_f7d9fa98594dc7c2e42bfcd641ff" {
		t.Errorf("card.Id should be 'car_f7d9fa98594dc7c2e42bfcd641ff', but '%s'", card.ID)
	}
	createdAt := card.CreatedAt.UTC().Format("2006-01-02 15:04:05")
	if createdAt != "2015-06-01 03:06:23" {
		t.Errorf("card.CreatedAt() should be '2015-06-01 03:06:23' but '%s'", createdAt)
	}
}

func TestCustomerAddCard(t *testing.T) {
	mock, transport := NewMockClient(200, cardResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
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
	mock, transport := NewMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Get("cus_121673955bd7aa144de5a8f6c262")
	if customer == nil {
		t.Error("plan should not be nil")
		return
	}
	card, err := customer.AddCard(Card{
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

func TestCustomerGetCard(t *testing.T) {
	mock, transport := NewMockClient(200, cardResponseJSON)
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
	mock, transport := NewMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Get("cus_121673955bd7aa144de5a8f6c262")
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
	mock, transport := NewMockClient(200, cardResponseJSON)
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
	mock, transport := NewMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Get("cus_121673955bd7aa144de5a8f6c262")
	if customer == nil {
		t.Error("plan should not be nil")
		return
	}
	err = customer.Cards[0].Update(Card{
		Number:   "4242424242424242",
		ExpMonth: 2,
		ExpYear:  2020,
	})
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
}

func TestCustomerUpdateCardError(t *testing.T) {
	mock, _ := NewMockClient(200, cardErrorResponseJSON)
	service := New("api-key", mock)
	card, err := service.Customer.UpdateCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff", Card{
		Number:   "4242424242424242",
		ExpMonth: 2,
		ExpYear:  2020,
	})
	if err == nil {
		t.Error("err should not be nil")
		return
	}
	if card != nil {
		t.Errorf("card should be nil, but %v", card)
	}
}

func TestCustomerDeleteCard(t *testing.T) {
	mock, transport := NewMockClient(200, cardResponseJSON)
	transport.AddResponse(200, cardDeleteResponseJSON)
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
	mock, transport := NewMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardDeleteResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Get("cus_121673955bd7aa144de5a8f6c262")
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
	mock, transport := NewMockClient(200, cardListResponseJSON)
	service := New("api-key", mock)
	cards, hasMore, err := service.Customer.ListCard("cus_121673955bd7aa144de5a8f6c262").
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards?limit=10&offset=15&since=1455328095&until=1455500895" {
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
	if len(cards) != 1 {
		t.Error("parse error: plans")
	}
}
