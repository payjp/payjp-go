package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
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

	assert.NoError(t, err)
	assert.Equal(t, "acct_8a27db83a7bf11a0c12b0c2833f", account.ID)
	assert.True(t, strings.Contains(account.CreatedAt.Format(time.RFC1123Z), "Sun, 16 Aug 2015 "))
	assert.Equal(t, "liveaccount@mail.com", account.Email)
	assert.Equal(t, "acct_mch_21a96cb898ceb6db0932983", account.Merchant.ID)
	assert.False(t, account.Merchant.BankEnabled)
	assert.Equal(t, "Visa, MasterCard, JCB, American Express, Diners Club, Discover", strings.Join(account.Merchant.BrandsAccepted, ", "))
	assert.Equal(t, "", account.Merchant.ContactPhone)
	assert.Equal(t, "JP", account.Merchant.Country)
	assert.Equal(t, 1, len(account.Merchant.CurrenciesSupported))
	assert.Equal(t, "jpy", account.Merchant.CurrenciesSupported[0])
	assert.Equal(t, "jpy", account.Merchant.DefaultCurrency)
	assert.False(t, account.Merchant.DetailsSubmitted)
	assert.True(t, strings.Contains(account.Merchant.LiveModeActivatedAt.Format(time.RFC1123Z), "Thu, 01 Jan 1970 "))
	assert.False(t, account.Merchant.LiveModeEnabled)
	assert.Equal(t, "", account.Merchant.ProductDetail)
	assert.Equal(t, "", account.Merchant.ProductName)
	assert.False(t, account.Merchant.SitePublished)
	assert.Equal(t, "", account.Merchant.URL)
	assert.Equal(t, "example-team-id", account.TeamID)
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
