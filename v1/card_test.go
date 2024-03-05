package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var cardResponseJSONStr = `
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
  "metadata": {},
  "object": "card"
}`
var cardResponseJSON = []byte(cardResponseJSONStr)

var cardListResponseJSON = []byte(`
{
  "count": 1,
  "data": [` + cardResponseJSONStr +
	`],
  "object": "list",
  "has_more": true,
  "url": "/v1/customers/cus_4df4b5ed720933f4fb9e28857517/cards"
}
`)

var cardUpdateResponseJSON = []byte(`
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
  "name": "pay",
  "three_d_secure_status": "verified",
  "metadata": {},
  "object": "card"
}`)

var deleteResponseJSONStr = `
{
  "deleted": true,
  "id": "xxx",
  "livemode": false
}`
var cardDeleteResponseJSON = []byte(deleteResponseJSONStr)

func TestParseCardResponseJSON(t *testing.T) {
	customerID := "cus_xxx"
	service := &Service{}
	card := &CardResponse{
		service:    service,
		customerID: customerID,
	}
	err := json.Unmarshal(cardUpdateResponseJSON, card)

	assert.NoError(t, err)
	assert.Equal(t, "car_f7d9fa98594dc7c2e42bfcd641ff", card.ID)
	assert.False(t, card.LiveMode)
	assert.Equal(t, 1433127983, *card.Created)
	assert.IsType(t, time.Unix(0, 0), card.CreatedAt)
	assert.Equal(t, "verified", *card.ThreeDSecureStatus)
	assert.Equal(t, map[string]string{}, card.Metadata)
	assert.Equal(t, customerID, card.customerID)
	assert.Equal(t, service, card.service)
}

func TestCustomerAddCardToken(t *testing.T) {
	mock, transport := newMockClient(200, cardResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	card, err := service.Customer.AddCardToken("cus_121673955bd7aa144de5a8f6c262", "tok_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card=tok_xxx", *transport.Body)
	assert.NotNil(t, card)
	assert.Equal(t, "4242", card.Last4)

	card, err = service.Customer.AddCardToken("cus_121673955bd7aa144de5a8f6c262", "tok_xxx", Customer{
		DefaultCard: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card=tok_xxx&default=true", *transport.Body)
	assert.NotNil(t, card)
	assert.Equal(t, "4242", card.Last4)

	_, err = service.Customer.AddCardToken("cus_121673955bd7aa144de5a8f6c262", "tok_xxx", Customer{
		DefaultCard: 1,
	})
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Equal(t, "card=tok_xxx&default=1", *transport.Body)
}

func TestCustomerResponseAddCardToken(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.NotNil(t, customer)

	card, err := customer.AddCardToken("tok_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card=tok_xxx", *transport.Body)
	assert.NotNil(t, card)
	assert.Equal(t, "4242", card.Last4)

	card, err = customer.AddCardToken("tok_xxx", Customer{
		DefaultCard: "true",
		Metadata: map[string]string{
			"hoge": "fuga",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card=tok_xxx&default=true&metadata[hoge]=fuga", *transport.Body)
	assert.NotNil(t, card)
	assert.Equal(t, "4242", card.Last4)

	_, err = customer.AddCardToken("tok_xxx", Customer{
		DefaultCard: 1,
	})
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Equal(t, "card=tok_xxx&default=1", *transport.Body)
}

func TestCustomerGetCard(t *testing.T) {
	mock, transport := newMockClient(200, cardResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	card, err := service.Customer.GetCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.NotNil(t, card)
	assert.Equal(t, "4242", card.Last4)

	_, err = service.Customer.GetCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerResponseGetCard(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.NotNil(t, customer)

	card, err := customer.GetCard("car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.NotNil(t, card)
	assert.Equal(t, "4242", card.Last4)

	_, err = customer.GetCard("car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerUpdateCard(t *testing.T) {
	mock, transport := newMockClient(200, cardResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	card, err := service.Customer.UpdateCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff", Card{
		Name: "pay",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card[name]=pay", *transport.Body)
	assert.NotNil(t, card)
	assert.Equal(t, "4242", card.Last4)

	_, err = service.Customer.UpdateCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff", Card{})
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerResponseUpdateCard(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardUpdateResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, cardResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.NotNil(t, customer)

	card := customer.Cards[0]
	assert.Equal(t, "", card.Name)
	err = card.Update(Card{
		Name: "pay",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card[name]=pay", *transport.Body)
	assert.Equal(t, "pay", card.Name)

	err = card.Update(Card{
		Name: "pay",
	})
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Equal(t, "pay", card.Name)

	card, err = customer.UpdateCard("car_f7d9fa98594dc7c2e42bfcd641ff", Card{
		Name: "",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "card[name]=", *transport.Body)
	assert.Equal(t, "", card.Name)

	res, err := customer.UpdateCard("car_f7d9fa98594dc7c2e42bfcd641ff", Card{})
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Equal(t, "", card.Name)
	assert.Equal(t, "", res.Name)
	// assert.Nil(t, res)
}

func TestCustomerDeleteCard(t *testing.T) {
	mock, transport := newMockClient(200, cardDeleteResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	err := service.Customer.DeleteCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.NoError(t, err)

	err = service.Customer.DeleteCard("cus_121673955bd7aa144de5a8f6c262", "car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerResponseDeleteCard(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardDeleteResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, cardDeleteResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	assert.NotNil(t, customer)

	card := customer.Cards[0]
	err = card.Delete()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.NotNil(t, card)

	err = card.Delete()
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())

	err = customer.DeleteCard("car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards/car_f7d9fa98594dc7c2e42bfcd641ff", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.NotNil(t, card)

	err = customer.DeleteCard("car_f7d9fa98594dc7c2e42bfcd641ff")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestCustomerListCard(t *testing.T) {
	mock, transport := newMockClient(200, customerResponseJSON)
	transport.AddResponse(200, cardListResponseJSON)
	transport.AddResponse(200, customerResponseJSON)
	transport.AddResponse(200, cardListResponseJSON)
	transport.AddResponse(200, customerResponseJSON)
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
	assert.Nil(t, cards[0].ThreeDSecureStatus)

	customer, err := service.Customer.Retrieve("cus_121673955bd7aa144de5a8f6c262")
	assert.NoError(t, err)
	params := &CardListParams{
		ListParams: ListParams{
			Limit:  Int(10),
			Offset: Int(15),
		},
	}
	cards, hasMore, err = customer.AllCard(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/cards?limit=10&offset=15", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(cards), 1)
	assert.Nil(t, cards[0].ThreeDSecureStatus)

	_, hasMore, err = service.Customer.ListCard("cus_121673955bd7aa144de5a8f6c262").Do()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
