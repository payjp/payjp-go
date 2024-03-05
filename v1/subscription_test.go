package payjp

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

var subscriptionResponseJSONStr = `
{
  "canceled_at": null,
  "created": 1433127983,
  "current_period_end": 1435732422,
  "current_period_start": 1433140422,
  "customer": "cus_4df4b5ed720933f4fb9e28857517",
  "id": "sub_response1",
  "livemode": false,
  "object": "subscription",
  "paused_at": null,
  "next_cycle_plan": {
    "amount": 1000,
    "billing_day": null,
    "created": 1432965398,
    "currency": "jpy",
    "id": "next_plan",
    "interval": "month",
    "name": "next plan",
    "object": "plan",
    "metadata": {},
    "trial_days": 0
  },
  "plan": {
    "amount": 1000,
    "billing_day": null,
    "created": 1432965397,
    "currency": "jpy",
    "id": "pln_9589006d14aad86aafeceac06b60",
    "interval": "month",
    "name": "test plan",
    "object": "plan",
    "metadata": {},
    "trial_days": 0
  },
  "resumed_at": null,
  "start": 1433140422,
  "status": "active",
  "trial_end": null,
  "trial_start": null,
  "metadata": {},
  "prorate": false
}
`
var subscriptionResponseJSON = []byte(subscriptionResponseJSONStr)

var nextCyclePlanNullResponseJSON = []byte(`
{
  "canceled_at": null,
  "created": 1433127983,
  "current_period_end": 1435732422,
  "current_period_start": 1433140422,
  "customer": "cus_4df4b5ed720933f4fb9e28857517",
  "id": "sub_response2",
  "livemode": false,
  "object": "subscription",
  "paused_at": null,
  "next_cycle_plan": null,
  "plan": {
    "amount": 1000,
    "billing_day": null,
    "created": 1432965397,
    "currency": "jpy",
    "id": "pln_9589006d14aad86aafeceac06b60",
    "interval": "month",
    "name": "test plan",
    "object": "plan",
    "metadata": {},
    "trial_days": 0
  },
  "resumed_at": null,
  "start": 1433140422,
  "status": "active",
  "trial_end": null,
  "trial_start": null,
  "metadata": {},
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
      "id": "sub_response3",
      "livemode": false,
      "object": "subscription",
      "paused_at": null,
      "next_cycle_plan": null,
      "plan": {
        "amount": 1000,
        "billing_day": null,
        "created": 1432965397,
        "currency": "jpy",
        "id": "pln_9589006d14aad86aafeceac06b60",
        "interval": "month",
        "name": "test plan",
        "object": "plan",
        "metadata": {},
        "trial_days": 0
      },
      "resumed_at": null,
      "start": 1433140422,
      "status": "active",
      "trial_end": null,
      "trial_start": null,
      "metadata": {},
      "prorate": false
    }
  ],
  "has_more": true,
  "object": "list",
  "url": "/v1/customers/cus_121673955bd7aa144de5a8f6c262/subscriptions"
}
`)

func TestParseSubscriptionResponseJSON(t *testing.T) {
	service := &Service{}
	subscription := &SubscriptionResponse{
		service: service,
	}
	err := json.Unmarshal(subscriptionResponseJSON, subscription)

	assert.NoError(t, err)
	assert.Equal(t, "sub_response1", subscription.ID)
	assert.Equal(t, SubscriptionActive, subscription.Status)
	assert.Equal(t, "active", subscription.Status.String())
	assert.Equal(t, 1000, subscription.Plan.Amount)
	assert.Equal(t, "next_plan", subscription.NextCyclePlan.ID)
	assert.Equal(t, "jpy", subscription.NextCyclePlan.Currency)
	assert.Equal(t, "jpy", subscription.NextCyclePlan.Currency)
	assert.Equal(t, "month", subscription.NextCyclePlan.Interval)
	assert.Equal(t, "next plan", subscription.NextCyclePlan.Name)
	assert.Equal(t, 0, subscription.NextCyclePlan.TrialDays)
	assert.NotNil(t, subscription.NextCyclePlan.Metadata)
	assert.Equal(t, service, subscription.service)
}

func TestCustomerGetSubscription(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	service := New("api-key", mock)

	subscription, err := service.Customer.GetSubscription("cus_xxx", "sub_req")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_xxx/subscriptions/sub_req", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.NotNil(t, subscription)
	assert.Equal(t, 1000, subscription.Plan.Amount)
}

func TestCustomerListSubscription(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	subscriptions, hasMore, err := service.Customer.ListSubscription("cus_xxx").
		Limit(1).
		Since(time.Unix(1, 0)).
		Until(time.Unix(2, 0)).Do()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions?customer=cus_xxx&limit=1&since=1&until=2", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(subscriptions), 1)
	assert.Nil(t, subscriptions[0].NextCyclePlan)

	_, hasMore, err = service.Customer.ListSubscription("error").Do()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestSubscriptionCreate(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	subscription, err := service.Subscription.Subscribe("cus_xxx", Subscription{
		PlanID:  "pln_yyy",
		Prorate: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "customer=cus_xxx&plan=pln_yyy&prorate=true", *transport.Body)
	assert.NotNil(t, subscription)
	assert.Equal(t, "pln_9589006d14aad86aafeceac06b60", subscription.Plan.ID)

	_, err = service.Subscription.Subscribe("error", Subscription{
		PlanID: "pln_yyy",
	})
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestSubscriptionCreateParamError(t *testing.T) {
	service := New("api-key", nil)

	subscription, err := service.Subscription.Subscribe("cus_xxx", Subscription{
		SkipTrial:  true,
		TrialEndAt: time.Now().AddDate(1, 0, 0),
	})
	assert.Nil(t, subscription)
	expected := fmt.Errorf("only either trial_end or SkipTrial is available")
	assert.Equal(t, expected, err)
}

func TestSubscriptionRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	subscription, err := service.Subscription.Retrieve("cus_xxx", "sub_req")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/customers/cus_xxx/subscriptions/sub_req", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, subscription)

	_, err = service.Subscription.Retrieve("cus_xxx", "sub_req")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestSubscriptionUpdate(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	transport.AddResponse(200, nextCyclePlanNullResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	subscription, err := service.Subscription.Update("sub_req", Subscription{
		PlanID:          "pln_xxx",
		NextCyclePlanID: "next_plan",
		Metadata: map[string]string{
			"hoge": "fuga",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_req", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "next_cycle_plan=next_plan&plan=pln_xxx&metadata[hoge]=fuga", *transport.Body)
	assert.NotNil(t, subscription)
	assert.Equal(t, "next_plan", subscription.NextCyclePlan.ID)

	err = subscription.Update(Subscription{
		NextCyclePlanID: "",
		Prorate:         "true",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_response1", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "next_cycle_plan=&prorate=true", *transport.Body)
	assert.Nil(t, subscription.NextCyclePlan)

	// next_cycle_plan未設定時はゼロ値ではなくrequest bodyに含まれないことをテスト
	err = subscription.Update(Subscription{
		Metadata: map[string]string{
			"hoge": "piyo",
		},
	})
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_response2", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "metadata[hoge]=piyo", *transport.Body)
	// Updateだけに行われていたparseResponseErrorが不要なことをテスト
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestSubscriptionPause(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	subscription, err := service.Subscription.Pause("sub_req")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_req/pause", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Nil(t, transport.Body)
	assert.NotNil(t, subscription)

	err = subscription.Pause()
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_response1/pause", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Nil(t, transport.Body)
	assert.NotNil(t, subscription)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestSubscriptionResume(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	nextYear := time.Now().AddDate(1, 0, 0)
	subscription, err := service.Subscription.Resume("sub_req", Subscription{
		TrialEndAt: nextYear,
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_req/resume", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	nextYearStr := strconv.Itoa(int(nextYear.Unix()))
	assert.Equal(t, "trial_end="+nextYearStr, *transport.Body)
	assert.NotNil(t, subscription)

	err = subscription.Resume(Subscription{})
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_response1/resume", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "", *transport.Body)
	assert.NotNil(t, subscription)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestSubscriptionCancel(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	subscription, err := service.Subscription.Cancel("sub_req")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_req/cancel", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Nil(t, transport.Body)
	assert.NotNil(t, subscription)

	err = subscription.Cancel()
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_response1/cancel", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Nil(t, transport.Body)
	assert.NotNil(t, subscription)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestSubscriptionDelete(t *testing.T) {
	mock, transport := newMockClient(200, []byte(deleteResponseJSONStr))
	service := New("api-key", mock)

	err := service.Subscription.Delete("sub_req", SubscriptionDelete{
		Prorate: String("false"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_req?prorate=false", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
}

func TestSubscriptionResponseDelete(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionResponseJSON)
	transport.AddResponse(200, []byte(deleteResponseJSONStr))
	service := New("api-key", mock)
	subscription, err := service.Subscription.Retrieve("cus_xxx", "sub_req")
	assert.NoError(t, err)

	err = subscription.Delete()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions/sub_response1", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.NotNil(t, subscription)
}

func TestSubscriptionList(t *testing.T) {
	mock, transport := newMockClient(200, subscriptionListResponseJSON)
	transport.AddResponse(200, subscriptionListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	subscriptions, hasMore, err := service.Subscription.List().
		Limit(1).
		Offset(0).Do()

	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions?limit=1&offset=0", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(subscriptions), 1)
	assert.Nil(t, subscriptions[0].NextCyclePlan)

	status := SubscriptionActive
	params := &SubscriptionListParams{
		ListParams: ListParams{
			Limit:  Int(1),
			Offset: Int(0),
		},
		Plan:     String("pln_xxx"),
		Status:   &status,
		Customer: String("cus_xxx"),
	}
	subscriptions, hasMore, err = service.Subscription.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/subscriptions?customer=cus_xxx&limit=1&offset=0&plan=pln_xxx&status=active", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(subscriptions), 1)
	assert.Nil(t, subscriptions[0].NextCyclePlan)

	_, hasMore, err = service.Subscription.List().Do()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
