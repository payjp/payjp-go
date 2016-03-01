package payjp

import (
	"encoding/json"
	"testing"
)

var tokenResponseJSON = []byte(`
{
  "card": {
    "address_city": null,
    "address_line1": null,
    "address_line2": null,
    "address_state": null,
    "address_zip": null,
    "address_zip_check": "unchecked",
    "brand": "Visa",
    "country": null,
    "created": 1442290383,
    "customer": null,
    "cvc_check": "passed",
    "exp_month": 2,
    "exp_year": 2020,
    "fingerprint": "e1d8225886e3a7211127df751c86787f",
    "id": "car_e3ccd4e0959f45e7c75bacc4be90",
    "last4": "4242",
    "name": null,
    "object": "card"
  },
  "created": 1442290383,
  "id": "tok_5ca06b51685e001723a2c3b4aeb4",
  "livemode": false,
  "object": "token",
  "used": false
}
`)

func TestParseTokenResponseJSON(t *testing.T) {
	token := &TokenResponse{}
	err := json.Unmarshal(tokenResponseJSON, token)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if token.ID != "tok_5ca06b51685e001723a2c3b4aeb4" {
		t.Errorf("token.Id should be 'tok_5ca06b51685e001723a2c3b4aeb4', but '%s'", token.ID)
	}
}

func TestTokenCreate(t *testing.T) {
	mock, transport := NewMockClient(200, tokenResponseJSON)
	service := New("api-key", mock)
	token, err := service.Token.Create(Card{
		Number:   "4242424242424242",
		ExpMonth: 2,
		ExpYear:  2020,
	})
	if transport.URL != "https://api.pay.jp/v1/tokens" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if token == nil {
		t.Error("plan should not be nil")
	} else if token.Card.ExpYear != 2020 {
		t.Errorf("token.Card.ExpYear should be 2020, but %d.", token.Card.ExpYear)
	}
}

func TestTokenGet(t *testing.T) {
	mock, transport := NewMockClient(200, tokenResponseJSON)
	service := New("api-key", mock)
	token, err := service.Token.Get("tok_5ca06b51685e001723a2c3b4aeb4")
	if transport.URL != "https://api.pay.jp/v1/tokens/tok_5ca06b51685e001723a2c3b4aeb4" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if token == nil {
		t.Error("plan should not be nil")
	} else if token.ID != "tok_5ca06b51685e001723a2c3b4aeb4" {
		t.Errorf("parse error: plan.Amount should be tok_5ca06b51685e001723a2c3b4aeb4, but %s.", token.ID)
	}
}
