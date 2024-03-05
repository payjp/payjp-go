package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var chargeResponseJSON = []byte(`
{
  "amount": 3500,
  "amount_refunded": 0,
  "captured": true,
  "captured_at": 1433127983,
  "card": {
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
    "id": "car_d0e44730f83b0a19ba6caee04160",
    "last4": "4242",
    "name": null,
    "object": "card"
  },
  "created": 1433127983,
  "currency": "jpy",
  "customer": "cus_xxx",
  "description": null,
  "expired_at": null,
  "failure_code": null,
  "failure_message": null,
  "metadata": null,
  "fee_rate": "3.00",
  "id": "ch_fa990a4c10672a93053a774730b0a",
  "livemode": false,
  "object": "charge",
  "paid": true,
  "refund_reason": null,
  "refunded": false,
  "three_d_secure_status": null,
  "subscription": null
}
`)

var chargeNewResponseJSONStr = `
{
  "amount": 1000,
  "amount_refunded": 0,
  "captured": true,
  "captured_at": 1432965397,
  "card": {
    "address_city": "\u8d64\u5742",
    "address_line1": "7-4",
    "address_line2": "203",
    "address_state": "\u6e2f\u533a",
    "address_zip": "1070050",
    "address_zip_check": "passed",
    "brand": "Visa",
    "country": "JP",
    "created": 1432965397,
    "cvc_check": "passed",
    "exp_month": 12,
    "exp_year": 2016,
    "fingerprint": "e1d8225886e3a7211127df751c86787f",
    "id": "car_7a79b41fed704317ec0deb4ebf93",
    "last4": "4242",
    "name": "Test Hodler",
    "object": "card"
  },
  "created": 1432965397,
  "currency": "jpy",
  "customer": null,
  "customer": "cus_67fab69c14d8888bba941ae2009b",
  "description": "new description",
  "expired_at": null,
  "failure_code": null,
  "failure_message": null,
  "metadata": null,
  "fee_rate": "3.00",
  "id": "ch_6421ddf0e12a5e5641d7426f2a2c9",
  "livemode": false,
  "object": "charge",
  "paid": true,
  "refund_reason": null,
  "refunded": false,
  "three_d_secure_status": "verified",
  "subscription": null
}`
var chargeNewResponseJSON = []byte(chargeNewResponseJSONStr)

var chargeListResponseJSON = []byte(`
{
  "count": 1,
  "data": [` + chargeNewResponseJSONStr +
	`],
  "has_more": true,
  "object": "list",
  "url": "/v1/charges"
}
`)

func TestParseChargeResponseJSON(t *testing.T) {
	service := &Service{}
	s := &ChargeResponse{
		service: service,
	}
	err := json.Unmarshal(chargeResponseJSON, s)

	assert.NoError(t, err)
	assert.Equal(t, "ch_fa990a4c10672a93053a774730b0a", s.ID)
	assert.False(t, s.LiveMode)
	assert.Equal(t, 1433127983, *s.Created)
	assert.IsType(t, time.Unix(0, 0), s.CreatedAt)
	assert.Equal(t, 3500, s.Amount)
	assert.Equal(t, "jpy", s.Currency)
	assert.True(t, s.Paid)
	assert.Nil(t, s.RawExpiredAt)
	assert.IsType(t, time.Unix(0, 0), s.ExpiredAt)
	assert.Equal(t, 1433127983, *s.RawCapturedAt)
	assert.IsType(t, time.Unix(0, 0), s.CapturedAt)
	assert.True(t, s.Captured)
	assert.Equal(t, "car_d0e44730f83b0a19ba6caee04160", s.Card.ID)
	// assert.Nil(t, s.Card.RawName)
	assert.Equal(t, "", s.Card.Name)
	assert.Equal(t, "cus_xxx", *s.Customer)
	assert.Equal(t, "cus_xxx", s.CustomerID)
	assert.Nil(t, s.RawDescription)
	assert.Equal(t, "", s.Description)
	assert.Nil(t, s.RawFailureCode)
	assert.Nil(t, s.RawFailureMessage)
	assert.Equal(t, "", s.FailureCode)
	assert.Equal(t, "", s.FailureMessage)
	assert.False(t, s.Refunded)
	assert.Equal(t, 0, s.AmountRefunded)
	assert.Nil(t, s.RawRefundReason)
	assert.Equal(t, "", s.RefundReason)
	assert.Nil(t, s.Subscription)
	assert.Equal(t, "", s.SubscriptionID)
	assert.Nil(t, s.Metadata)
	assert.Equal(t, "3.00", s.FeeRate)
	assert.Nil(t, s.ThreeDSecureStatus)
	assert.Equal(t, "charge", s.Object)
	assert.Equal(t, service, s.service)
}

func TestChargeCreate(t *testing.T) {
	mock, transport := newMockClient(200, chargeResponseJSON)
	transport.AddResponse(200, chargeResponseJSON)
	transport.AddResponse(200, chargeResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	charge, err := service.Charge.Create(3500, Charge{
		CardToken: "tok_req1",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "amount=3500&currency=jpy&card=tok_req1&capture=false", *transport.Body)
	assert.NotNil(t, charge)
	assert.Equal(t, 3500, charge.Amount)

	charge, err = service.Charge.Create(3500, Charge{
		Product:        "prd_req1",
		CustomerID:     "cus_req1",
		CustomerCardID: "car_req1",
		Description:    "desc",
		Capture:        true,
		Metadata: map[string]string{
			"hoge": "fuga",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t,
		"product=prd_req1&customer=cus_req1&card=car_req1&capture=true&description=desc&metadata[hoge]=fuga",
		*transport.Body)
	assert.NotNil(t, charge)

	charge, err = service.Charge.Create(3500, Charge{
		Product:      "prd_req1",
		CardToken:    "tok_req1",
		ExpireDays:   1,
		ThreeDSecure: false,
	})
	assert.NoError(t, err)
	assert.Equal(t, "product=prd_req1&card=tok_req1&capture=false&expiry_days=1&three_d_secure=false", *transport.Body)
	assert.NotNil(t, charge)

	_, err = service.Charge.Create(1000, Charge{
		ExpireDays:   "invalid",
		ThreeDSecure: "invalid",
	})
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, "amount=1000&currency=jpy&capture=false&expiry_days=invalid&three_d_secure=invalid", *transport.Body)
	assert.Equal(t, errorStr, err.Error())
}

func TestChargeRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, chargeResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	charge, err := service.Charge.Retrieve("ch_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_xxx", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, charge)

	charge, err = service.Charge.Retrieve("ch_err")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Nil(t, charge)
}

func TestChargeUpdate(t *testing.T) {
	mock, transport := newMockClient(200, chargeResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, chargeNewResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	charge, err := service.Charge.Update("ch_req", "new description")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_req", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "description=new+description", *transport.Body)
	assert.NotNil(t, charge)
	assert.Equal(t, "", charge.Description)

	chargeErr, err := service.Charge.Update("ch_req", "new description", map[string]string{
		"hoge": "fuga",
	})
	assert.Equal(t, "description=new+description&metadata[hoge]=fuga", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Nil(t, chargeErr)

	err = charge.Update("new description")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "description=new+description", *transport.Body)
	assert.Equal(t, "new description", charge.Description)

	err = charge.Update("new description2", map[string]string{
		"hoge": "fuga",
	})
	assert.Equal(t, "description=new+description2&metadata[hoge]=fuga", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Equal(t, "new description", charge.Description)
}

func TestChargeRefund(t *testing.T) {
	mock, transport := newMockClient(200, chargeResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, chargeNewResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	charge, err := service.Charge.Refund("ch_fa990a4c10672a93053a774730b0a", "reason")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/refund", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "refund_reason=reason", *transport.Body)
	assert.NotNil(t, charge)
	assert.Equal(t, "", charge.RefundReason)

	chargeErr, err := service.Charge.Refund("ch_fa990a4c10672a93053a774730b0a", "reason", 500)
	assert.Equal(t, "amount=500&refund_reason=reason", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Nil(t, chargeErr)

	err = charge.Refund("reason")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/refund", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "refund_reason=reason", *transport.Body)
	assert.Equal(t, 1000, charge.Amount)

	err = charge.Refund("reason", 500)
	assert.Equal(t, "amount=500&refund_reason=reason", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Equal(t, 1000, charge.Amount)
}

func TestChargeCapture(t *testing.T) {
	mock, transport := newMockClient(200, chargeResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, chargeNewResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	charge, err := service.Charge.Capture("ch_fa990a4c10672a93053a774730b0a")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/capture", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "", *transport.Body)
	assert.Equal(t, 3500, charge.Amount)

	chargeErr, err := service.Charge.Capture("ch_fa990a4c10672a93053a774730b0a", 300)
	assert.Equal(t, "amount=300", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Nil(t, chargeErr)

	err = charge.Capture()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/capture", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "", *transport.Body)

	err = charge.Capture(100)
	assert.Equal(t, "amount=100", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Equal(t, 1000, charge.Amount)
}

func TestChargeTdsFinish(t *testing.T) {
	mock, transport := newMockClient(200, chargeResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, chargeNewResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	charge, err := service.Charge.TdsFinish("ch_req")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_req/tds_finish", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Nil(t, transport.Body)

	chargeErr, err := service.Charge.TdsFinish("ch_req")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Nil(t, chargeErr)

	err = charge.TdsFinish()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/tds_finish", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Nil(t, transport.Body)
	assert.Equal(t, "verified", *charge.ThreeDSecureStatus)

	err = charge.TdsFinish()
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.NotNil(t, charge)
}

func TestChargeCaptureChangedAmount(t *testing.T) {
	mock, transport := newMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	chargeID := "ch_fa990a4c10672a93053a774730b0a"
	charge, err := service.Charge.Retrieve(chargeID)
	assert.NoError(t, err)
	newAmount := 100
	err = charge.Capture(newAmount)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges/"+chargeID+"/capture", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "amount=100", *transport.Body)
}

func TestChargeList(t *testing.T) {
	mock, transport := newMockClient(200, chargeListResponseJSON)
	transport.AddResponse(200, chargeListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	charges, hasMore, err := service.Charge.List().
		Limit(10).
		Offset(0).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()

	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges?limit=10&offset=0&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(charges), 1)
	assert.Equal(t, 1000, charges[0].Amount)
	assert.Equal(t, service, charges[0].service)

	params := &ChargeListParams{
		ListParams: ListParams{
			Limit:  Int(10),
			Offset: Int(0),
		},
		Customer:     String("a"),
		Subscription: String("b"),
		Tenant:       String("c"),
	}
	charges, hasMore, err = service.Charge.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/charges?customer=a&limit=10&offset=0&subscription=b&tenant=c", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(charges), 1)
	assert.Equal(t, 1000, charges[0].Amount)
	assert.Equal(t, service, charges[0].service)

	_, hasMore, err = service.Charge.List().Do()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
