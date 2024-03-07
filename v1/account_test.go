package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
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
	a := &AccountResponse{}
	err := json.Unmarshal(accountResponseJSON, a)

	assert.NoError(t, err)
	assert.Equal(t, "acct_8a27db83a7bf11a0c12b0c2833f", a.ID)
	assert.False(t, a.Merchant.LiveModeEnabled)
	assert.Equal(t, "example-team-id", a.TeamID)
}

func TestAccountRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, accountResponseJSON)
	service := New("api-key", mock)
	a, err := service.Account.Retrieve()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/accounts", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Equal(t, "liveaccount@mail.com", a.Email)
}
