package payjp

import (
	"encoding/json"
	"testing"
)

var accountResponseJSON = []byte(`
{
  "created": 1439706600,
  "customer": {
    "cards": {
      "count": 1,
      "data": [
        {
          "address_city": "赤坂",
          "address_line1": "7-4",
          "address_line2": "203",
          "address_state": "港区",
          "address_zip": "1070050",
          "address_zip_check": "passed",
          "brand": "Visa",
          "country": "JP",
          "created": 1439706600,
          "cvc_check": "passed",
          "exp_month": 12,
          "exp_year": 2016,
          "fingerprint": "e1d8225886e3a7211127df751c86787f",
          "id": "car_99abf74cb5527ff68233a8b836dd",
          "last4": "4242",
          "livemode": true,
          "name": "Test Hodler",
          "object": "card"
        }
      ],
      "has_more": false,
      "object": "list",
      "url": "/v1/accounts/cards"
    },
    "created": 1439706600,
    "default_card": null,
    "description": "account customer",
    "email": null,
    "id": "acct_cus_7d03658e143dee2ef876b3e",
    "object": "customer"
  },
  "email": "liveaccount@mail.com",
  "id": "acct_8a27db83a7bf11a0c12b0c2833f",
  "merchant": {
    "bank_enabled": false,
    "brands_accepted": [
      "Visa",
      "MasterCard",
      "JCB",
      "American Express",
      "Diners Club",
      "Discover"
    ],
    "business_type": null,
    "charge_type": null,
    "contact_phone": null,
    "country": "JP",
    "created": 1439706600,
    "currencies_supported": [
      "jpy"
    ],
    "default_currency": "jpy",
    "details_submitted": false,
    "id": "acct_mch_21a96cb898ceb6db0932983",
    "livemode_activated_at": 0,
    "livemode_enabled": false,
    "object": "merchant",
    "product_detail": null,
    "product_name": null,
    "product_type": null,
    "site_published": null,
    "url": null
  },
  "object": "account"
}
`)

func TestParseAccountResponseJSON(t *testing.T) {
	account := &AccountResponse{}
	err := json.Unmarshal(accountResponseJSON, account)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if account.ID != "acct_8a27db83a7bf11a0c12b0c2833f" {
		t.Errorf("customer.ID should be 'acct_8a27db83a7bf11a0c12b0c2833f', but '%s'", account.ID)
	}
	if account.Merchant.DefaultCurrency != "jpy" {
		t.Errorf("defaultCurrency should be 'jpy', but %s", account.Merchant.DefaultCurrency)
	}
}

func TestAccountGet(t *testing.T) {
	mock, transport := NewMockClient(200, accountResponseJSON)
	service := New("api-key", mock)
	account, err := service.Account.Get()
	if transport.URL != "https://api.pay.jp/v1/accounts" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if account == nil {
		t.Error("plan should not be nil")
	} else if account.Email != "liveaccount@mail.com" {
		t.Errorf("parse error: account.Email should be 'liveaccount@mail.com', but %s.", account.Email)
	}
}
