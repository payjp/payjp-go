package payjp

import (
	"encoding/json"
	"testing"
)

var accountResponseJSON = []byte(`
{
  "created": 1439706600,
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
  "object": "account",
  "team_id": "example-team-id"
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
	if account.TeamID != "example-team-id" {
		t.Errorf("account.TeamID should be 'example-team-id', but %s", account.TeamID)
	}
}

func TestAccountRetrieve(t *testing.T) {
	mock, transport := NewMockClient(200, accountResponseJSON)
	service := New("api-key", mock)
	account, err := service.Account.Retrieve()
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
	} else if account.TeamID != "example-team-id" {
		t.Errorf("parse error: account.TeamID should be 'example-team-id', but %s.", account.TeamID)
	}
}
