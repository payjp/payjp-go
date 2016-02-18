package payjp

import (
	"encoding/json"
	"testing"
	"time"
)

var subscriptionResponseJSON = []byte(`
{
  "canceled_at": null,
  "created": 1433127983,
  "current_period_end": 1435732422,
  "current_period_start": 1433140422,
  "customer": "cus_4df4b5ed720933f4fb9e28857517",
  "id": "sub_567a1e44562932ec1a7682d746e0",
  "livemode": false,
  "object": "subscription",
  "paused_at": null,
  "plan": {
    "amount": 1000,
    "billing_day": null,
    "created": 1432965397,
    "currency": "jpy",
    "id": "pln_9589006d14aad86aafeceac06b60",
    "interval": "month",
    "name": "test plan",
    "object": "plan",
    "trial_days": 0
  },
  "resumed_at": null,
  "start": 1433140422,
  "status": "active",
  "trial_end": null,
  "trial_start": null,
  "prorate": false
}
`)

var subscriptionListResponseJSON = []byte(`
{
  "count": 1,
  "data": [
    {
      "canceled_at": null,
      "created": 1433127983,
      "current_period_end": 1435732422,
      "current_period_start": 1433140422,
      "customer": "cus_4df4b5ed720933f4fb9e28857517",
      "id": "sub_567a1e44562932ec1a7682d746e0",
      "livemode": false,
      "object": "subscription",
      "paused_at": null,
      "plan": {
        "amount": 1000,
        "billing_day": null,
        "created": 1432965397,
        "currency": "jpy",
        "id": "pln_9589006d14aad86aafeceac06b60",
        "interval": "month",
        "name": "test plan",
        "object": "plan",
        "trial_days": 0
      },
      "resumed_at": null,
      "start": 1433140422,
      "status": "active",
      "trial_end": null,
      "trial_start": null,
      "prorate": false
    }
  ],
  "has_more": true,
  "object": "list",
  "url": "/v1/customers/cus_121673955bd7aa144de5a8f6c262/subscriptions"
}
`)

func TestParseSubscriptionResponseJSON(t *testing.T) {
	subscription := &SubscriptionResponse{}
	err := json.Unmarshal(subscriptionResponseJSON, subscription)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if subscription.ID != "sub_567a1e44562932ec1a7682d746e0" {
		t.Errorf("subscription.ID should be 'sub_567a1e44562932ec1a7682d746e0', but '%s'", subscription.ID)
	}
	if subscription.Plan.Amount != 1000 {
		t.Errorf("subscription.Plan.Amount should be 1000 but %d", subscription.Plan.Amount)
	}
}

func TestCustomerGetSubscription(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	subscription, err := service.Customer.GetSubscription("cus_121673955bd7aa144de5a8f6c262", "sub_567a1e44562932ec1a7682d746e0")
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/subscriptions/sub_567a1e44562932ec1a7682d746e0" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if subscription == nil {
		t.Error("subscription should not be nil")
	} else if subscription.Plan.Amount != 1000 {
		t.Errorf("subscription.Plan.Amount should be 1000 but %d", subscription.Plan.Amount)
	}
}

func TestCustomerListSubscription(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionListResponseJSON)
	service := New("api-key", mock)
	subscriptions, hasMore, err := service.Customer.ListSubscription("cus_121673955bd7aa144de5a8f6c262").
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/subscriptions?limit=10&offset=15&since=1455328095&until=1455500895" {
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
	if len(subscriptions) != 1 {
		t.Error("parse error: plans")
	}
}

func TestSubscriptionCreate(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	subscription, err := service.Subscription.Subscribe("cus_4df4b5ed720933f4fb9e28857517", Subscription{
		PlanID: "pln_9589006d14aad86aafeceac06b60",
	})
	if transport.URL != "https://api.pay.jp/v1/subscriptions" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if subscription == nil {
		t.Error("subscription should not be nil")
	} else if subscription.Plan.ID != "pln_9589006d14aad86aafeceac06b60" {
		t.Errorf("subscription.Plan.ID is wrong: %s.", subscription.Plan.ID)
	}
}

func TestSubscriptionGet(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	subscription, err := service.Subscription.Get("cus_121673955bd7aa144de5a8f6c262", "sub_567a1e44562932ec1a7682d746e0")
	if transport.URL != "https://api.pay.jp/v1/customers/cus_121673955bd7aa144de5a8f6c262/subscriptions/sub_567a1e44562932ec1a7682d746e0" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if subscription == nil {
		t.Error("subscription should not be nil")
	} else if subscription.Plan.Amount != 1000 {
		t.Errorf("subscription.Plan.Amount should be 1000 but %d", subscription.Plan.Amount)
	}
}

func TestSubscriptionUpdate(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	subscription, err := service.Subscription.Update("sub_567a1e44562932ec1a7682d746e0", Subscription{
		PlanID: "pln_9589006d14aad86aafeceac06b60",
	})
	if transport.URL != "https://api.pay.jp/v1/subscriptions/sub_567a1e44562932ec1a7682d746e0" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if subscription == nil {
		t.Error("subscription should not be nil")
	} else if subscription.Plan.ID != "pln_9589006d14aad86aafeceac06b60" {
		t.Errorf("subscription.Plan.ID is wrong: %s.", subscription.Plan.ID)
	}
}

func TestSubscriptionPause(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	subscription, err := service.Subscription.Pause("sub_567a1e44562932ec1a7682d746e0")
	if transport.URL != "https://api.pay.jp/v1/subscriptions/sub_567a1e44562932ec1a7682d746e0/pause" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if subscription == nil {
		t.Error("subscription should not be nil")
	} else if subscription.Plan.ID != "pln_9589006d14aad86aafeceac06b60" {
		t.Errorf("subscription.Plan.ID is wrong: %s.", subscription.Plan.ID)
	}
}

func TestSubscriptionResume(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	subscription, err := service.Subscription.Resume("sub_567a1e44562932ec1a7682d746e0")
	if transport.URL != "https://api.pay.jp/v1/subscriptions/sub_567a1e44562932ec1a7682d746e0/resume" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if subscription == nil {
		t.Error("subscription should not be nil")
	} else if subscription.Plan.ID != "pln_9589006d14aad86aafeceac06b60" {
		t.Errorf("subscription.Plan.ID is wrong: %s.", subscription.Plan.ID)
	}
}

func TestSubscriptionCancel(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	subscription, err := service.Subscription.Cancel("sub_567a1e44562932ec1a7682d746e0")
	if transport.URL != "https://api.pay.jp/v1/subscriptions/sub_567a1e44562932ec1a7682d746e0/cancel" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if subscription == nil {
		t.Error("subscription should not be nil")
	} else if subscription.Plan.ID != "pln_9589006d14aad86aafeceac06b60" {
		t.Errorf("subscription.Plan.ID is wrong: %s.", subscription.Plan.ID)
	}
}

func TestSubscriptionDelete(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)
	err := service.Subscription.Delete("sub_567a1e44562932ec1a7682d746e0")
	if transport.URL != "https://api.pay.jp/v1/subscriptions/sub_567a1e44562932ec1a7682d746e0" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "DELETE" {
		t.Errorf("Method should be DELETE, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
}

func TestSubscriptionList(t *testing.T) {
	mock, transport := NewMockClient(200, subscriptionListResponseJSON)
	service := New("api-key", mock)
	subscriptions, hasMore, err := service.Subscription.List().
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/subscriptions?limit=10&offset=15&since=1455328095&until=1455500895" {
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
	if len(subscriptions) != 1 {
		t.Error("parse error: plans")
	}
}
