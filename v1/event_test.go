package payjp

import (
	"encoding/json"
	"testing"
	"time"
)

var eventResponseJSON = []byte(`
{
  "created": 1442288882,
  "data": {
    "cards": {
      "count": 0,
      "data": [],
      "has_more": false,
      "object": "list",
      "url": "/v1/customers/cus_a16c7b4df01168eb82557fe93de4/cards"
    },
    "created": 1441936720,
    "default_card": null,
    "description": "updated\n",
    "email": null,
    "id": "cus_a16c7b4df01168eb82557fe93de4",
    "livemode": false,
    "object": "customer",
    "subscriptions": {
      "count": 0,
      "data": [],
      "has_more": false,
      "object": "list",
      "url": "/v1/customers/cus_a16c7b4df01168eb82557fe93de4/subscriptions"
    }
  },
  "id": "evnt_54db4d63c7886256acdbc784ccf",
  "livemode": false,
  "object": "event",
  "pending_webhooks": 1,
  "type": "customer.updated"
}
`)

var eventListResponseJSON = []byte(`
{
  "count": 1,
  "data": [
    {
      "created": 1442298026,
      "data": {
        "amount": 5000,
        "amount_refunded": 5000,
        "captured": true,
        "captured_at": 1442212986,
        "card": {
          "address_city": null,
          "address_line1": null,
          "address_line2": null,
          "address_state": null,
          "address_zip": null,
          "address_zip_check": "unchecked",
          "brand": "Visa",
          "country": null,
          "created": 1442212986,
          "customer": null,
          "cvc_check": "passed",
          "exp_month": 1,
          "exp_year": 2016,
          "fingerprint": "e1d8225886e3a7211127df751c86787f",
          "id": "car_f0984a6f68a730b7e1814ceabfe1",
          "last4": "4242",
          "name": null,
          "object": "card"
        },
        "created": 1442212986,
        "currency": "jpy",
        "customer": null,
        "description": "hogehoe",
        "expired_at": null,
        "failure_code": null,
        "failure_message": null,
        "id": "ch_bcb7776459913c743c20e9f9351d4",
        "livemode": false,
        "object": "charge",
        "paid": true,
        "refund_reason": null,
        "refunded": true,
        "subscription": null
      },
      "id": "evnt_8064917698aa417a3c86d292266",
      "livemode": false,
      "object": "event",
      "pending_webhooks": 1,
      "type": "charge.updated"
    }
  ],
  "has_more": true,
  "object": "list",
  "url": "/v1/events"
}
`)

func TestParseEventResponseJSON(t *testing.T) {
	event := &EventResponse{}
	err := json.Unmarshal(eventResponseJSON, event)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if event.ID != "evnt_54db4d63c7886256acdbc784ccf" {
		t.Errorf("event.ID should be 'evnt_54db4d63c7886256acdbc784ccf', but '%s'", event.ID)
	}
	if event.Type != "customer.updated" {
		t.Errorf("parse error: %s", event.Type)
	}
	if event.ResultType != CustomerEvent {
		t.Errorf("parse error: %v", event.ResultType)
	}

	card, err := event.CardData()
	if card != nil {
		t.Errorf("card should be nil, but %v", card)
	}
	if err == nil {
		t.Error("error should not be nil")
	}

	customer, err := event.CustomerData()
	if customer == nil {
		t.Error("customer should not be nil")
	}
	if err != nil {
		t.Errorf("error should be nil, but %v", err)
	}
}

func TestEventGet(t *testing.T) {
	mock, transport := NewMockClient(200, eventResponseJSON)
	service := New("api-key", mock)
	event, err := service.Event.Get("evnt_54db4d63c7886256acdbc784ccf")
	if transport.URL != "https://api.pay.jp/v1/events/evnt_54db4d63c7886256acdbc784ccf" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if event == nil {
		t.Error("event should not be nil")
	} else if event.PendingWebHooks != 1 {
		t.Errorf("parse error: event.PendingWebHooks should be 1, but %d.", event.PendingWebHooks)
	}
}

func TestEventList(t *testing.T) {
	mock, transport := NewMockClient(200, eventListResponseJSON)
	service := New("api-key", mock)
	events, hasMore, err := service.Event.List().
		Limit(10).
		Offset(15).
		Type("charge.updated").
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/events?limit=10&offset=15&since=1455328095&type=charge.updated&until=1455500895" {
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
	if len(events) != 1 {
		t.Error("parse error: plans")
	} else if events[0].PendingWebHooks != 1 {
		t.Errorf("parse error: event.PendingWebHooks should be 1, but %d.", events[0].PendingWebHooks)
	}
}
