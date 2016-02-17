package payjp

import (
	"encoding/json"
	"testing"
)

var eventResponseJson []byte = []byte(`
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

func TestParseEventResponseJson(t *testing.T) {
	event := &Event{}
	err := json.Unmarshal(eventResponseJson, event)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if event.Object != "event" {
		t.Errorf("parse error")
	}
	if event.ID != "evnt_54db4d63c7886256acdbc784ccf" {
		t.Errorf("event.ID should be 'evnt_54db4d63c7886256acdbc784ccf', but '%s'", event.ID)
	}
}
