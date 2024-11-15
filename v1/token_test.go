package payjp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
    "name": "PAY TARO",
    "object": "card",
    "email": "liveaccount@example.com",
    "phone": "+81301234567"
  },
  "created": 1442290383,
  "id": "tok_5ca06b51685e001723a2c3b4aeb4",
  "livemode": false,
  "object": "token",
  "used": false
}
`)

// changes: three_d_secure_status
var tokenNewResponseJSON = []byte(`
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
    "name": "PAY TARO",
    "object": "card",
    "email": "liveaccount@example.com",
    "phone": "+81301234567",
    "three_d_secure_status": "verified"
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

	assert.NoError(t, err)
	assert.Equal(t, "tok_5ca06b51685e001723a2c3b4aeb4", token.ID)
	assert.Equal(t, "liveaccount@example.com", *token.Card.Email)
	assert.Equal(t, "+81301234567", *token.Card.Phone)
}

func TestTokenCreate(t *testing.T) {
	mock, transport := newMockClient(200, tokenResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	p := Token{
		Number:   "4242424242424242",
		ExpMonth: 2,
		ExpYear:  "2020",
	}
	p.Name = "pay"
	token, err := service.Token.Create(p)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/tokens", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "true", transport.Header.Get("X-Payjp-Direct-Token-Generate"))
	assert.Equal(t, "card[number]=4242424242424242&card[exp_month]=2&card[exp_year]=2020&card[name]=pay", *transport.Body)
	assert.NotNil(t, token)
	assert.Equal(t, "4242", token.Card.Last4)

	token, err = service.Token.Create(Token{})
	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestTokenRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, tokenResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	token, err := service.Token.Retrieve("tok_5ca06b51685e001723a2c3b4aeb4")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/tokens/tok_5ca06b51685e001723a2c3b4aeb4", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "tok_5ca06b51685e001723a2c3b4aeb4", token.ID)
	assert.PanicsWithValue(t, "token's card doens't support Delete()", func() { token.Card.Delete() })

	token, err = service.Token.Retrieve("hoge")
	assert.Nil(t, token)
	assert.Error(t, err)
}

func TestTokenTdsFinish(t *testing.T) {
	mock, transport := newMockClient(200, tokenResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, tokenNewResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	token, err := service.Token.TdsFinish("tok_req")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/tokens/tok_req/tds_finish", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Nil(t, transport.Body)

	tokenErr, err := service.Token.TdsFinish("tok_req")
	assert.Error(t, err)
	payErr := err.(*Error)
	assert.Equal(t, 400, payErr.Status)
	assert.Nil(t, tokenErr)

	err = token.TdsFinish()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/tokens/tok_5ca06b51685e001723a2c3b4aeb4/tds_finish", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Nil(t, transport.Body)
	assert.Equal(t, "verified", *token.Card.ThreeDSecureStatus)

	err = token.TdsFinish()
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.NotNil(t, token)
}
