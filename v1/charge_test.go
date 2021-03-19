package payjp

import (
	"context"
	"encoding/json"
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
  "customer": null,
  "description": null,
  "expired_at": null,
  "failure_code": null,
  "failure_message": null,
  "fee_rate": "3.00",
  "id": "ch_fa990a4c10672a93053a774730b0a",
  "livemode": false,
  "object": "charge",
  "paid": true,
  "refund_reason": null,
  "refunded": false,
  "subscription": null
}
`)

var chargeListResponseJSON = []byte(`
{
  "count": 1,
  "data": [
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
      "customer": "cus_67fab69c14d8888bba941ae2009b",
      "description": "test charge",
      "expired_at": null,
      "failure_code": null,
      "failure_message": null,
      "fee_rate": "3.00",
      "id": "ch_6421ddf0e12a5e5641d7426f2a2c9",
      "livemode": false,
      "object": "charge",
      "paid": true,
      "refund_reason": null,
      "refunded": false,
      "subscription": null
    }
  ],
  "has_more": true,
  "object": "list",
  "url": "/v1/charges"
}
`)

var chargeErrorResponseJSON = []byte(`
{
  "error": {
    "code": "invalid_number",
    "message": "Your card number is invalid.",
    "param": "card[number]",
    "status": 400,
    "type": "card_error"
  }
}
`)

func TestParseChargeResponseJSON(t *testing.T) {
	charge := &ChargeResponse{}
	err := json.Unmarshal(chargeResponseJSON, charge)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
}

func TestParseChargeErrorResponseJSON(t *testing.T) {
	charge := &ChargeResponse{}
	err := json.Unmarshal(chargeErrorResponseJSON, charge)

	if err == nil {
		t.Error("err should not be nil")
	}
}

func TestChargeCreate(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	charge, err := service.Charge.Create(1000, Charge{
		Card: Card{
			Number:   "4242424242424242",
			ExpMonth: 2,
			ExpYear:  2020,
		},
	})
	if transport.URL != "https://api.pay.jp/v1/charges" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if charge == nil {
		t.Error("charge should not be nil")
	} else if charge.Amount != 3500 {
		t.Errorf("charge.Amount should be 3500, but %d.", charge.Amount)
	}
}

func TestChargeCreateByNonDefaultard(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	charge, err := service.Charge.Create(1000, Charge{
		CustomerID:     "cus_xxxxxxxxx",
		CustomerCardID: "car_xxxxxxxxx",
	})

	if transport.URL != "https://api.pay.jp/v1/charges" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}

	if charge == nil {
		t.Error("charge should not be nil")
	}
}

func TestChargeRetrieve(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Charge.Retrieve("ch_fa990a4c10672a93053a774730b0a")
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if plan == nil {
		t.Error("plan should not be nil")
	} else if plan.Amount != 3500 {
		t.Errorf("parse error: plan.Amount should be 500, but %d.", plan.Amount)
	}
}

func TestChargeGetError(t *testing.T) {
	mock, _ := NewMockClient(200, chargeErrorResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Charge.Retrieve("ch_fa990a4c10672a93053a774730b0a")
	if err == nil {
		t.Error("err should not be nil")
	}
	if plan != nil {
		t.Errorf("plan should be nil, but %v", plan)
	}
}

func TestChargeUpdate(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Charge.Update("ch_fa990a4c10672a93053a774730b0a", "new description")
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if plan == nil {
		t.Error("plan should not be nil")
	} else if plan.Amount != 3500 {
		t.Errorf("parse error: plan.Amount should be 500, but %d.", plan.Amount)
	}
}

func TestChargeUpdate2(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Charge.Retrieve("ch_fa990a4c10672a93053a774730b0a")
	if plan == nil {
		t.Error("plan should not be nil")
		return
	}
	err = plan.UpdateContext(context.Background(), "new description")
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
}

func TestChargeRefund(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	_, err := service.Charge.Refund("ch_fa990a4c10672a93053a774730b0a", "reason")
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/refund" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
}

func TestChargeRefund2(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Charge.Retrieve("ch_fa990a4c10672a93053a774730b0a")
	if plan == nil {
		t.Error("plan should not be nil")
		return
	}
	err = plan.RefundContext(context.Background(), "reason")
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/refund" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
}

func TestChargeCapture(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	_, err := service.Charge.Capture("ch_fa990a4c10672a93053a774730b0a")
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/capture" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
}

func TestServiceChargeCaptureChangeAmount(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	_, err := service.Charge.Capture("ch_fa990a4c10672a93053a774730b0a", 300)
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/capture" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
}

func TestChargeCapture2(t *testing.T) {
	ctx := context.Background()
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Charge.RetrieveContext(ctx, "ch_fa990a4c10672a93053a774730b0a")
	if plan == nil {
		t.Error("plan should not be nil")
		return
	}
	err = plan.CaptureContext(ctx)
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/charges/ch_fa990a4c10672a93053a774730b0a/capture" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
}

func TestChargeCaptureChangedAmount(t *testing.T) {
	mock, transport := NewMockClient(200, chargeResponseJSON)
	service := New("api-key", mock)
	chargeID := "ch_fa990a4c10672a93053a774730b0a"
	charge, err := service.Charge.Retrieve(chargeID)
	newAmount := 100
	err = charge.CaptureContext(context.Background(), newAmount)
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/charges/"+chargeID+"/capture" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
}

func TestChargeList(t *testing.T) {
	mock, transport := NewMockClient(200, chargeListResponseJSON)
	service := New("api-key", mock)
	plans, hasMore, err := service.Charge.List().
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/charges?limit=10&offset=15&since=1455328095&until=1455500895" {
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
	if len(plans) != 1 {
		t.Error("parse error: plans")
	}
}
