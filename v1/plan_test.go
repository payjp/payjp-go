package payjp

import (
	"encoding/json"
	"testing"
	"time"
)

var planResponseJSON = []byte(`{
  "amount": 500,
  "billing_day": null,
  "created": 1433127983,
  "currency": "jpy",
  "id": "pln_45dd3268a18b2837d52861716260",
  "interval": "month",
  "livemode": false,
  "name": null,
  "object": "plan",
  "trial_days": 30
}`)

var planListResponseJSON = []byte(`
{
  "count": 1,
  "data": [
    {
      "amount": 500,
      "billing_day": null,
      "created": 1433127983,
      "currency": "jpy",
      "id": "pln_45dd3268a18b2837d52861716260",
      "interval": "month",
      "livemode": false,
      "name": null,
      "object": "plan",
      "trial_days": 30
    }
  ],
  "object": "list",
  "has_more": true,
  "url": "/v1/customers/cus_4df4b5ed720933f4fb9e28857517/cards"
}
`)

var planErrorResponseJSON = []byte(`
{
  "error": {
    "message": "There is no plan with ID: dummy",
    "param": "id",
    "status": 404,
    "type": "client_error"
  }
}
`)

func TestParsePlanResponseJSON(t *testing.T) {
	plan := &PlanResponse{}
	err := json.Unmarshal(planResponseJSON, plan)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if plan.CreatedAt.Unix() != 1433127983 {
		t.Errorf("plan.Created should be '1433127983', but '%d'", plan.CreatedAt.Unix())
	}
}

func TestPlanCreate(t *testing.T) {
	mock, transport := NewMockClient(200, planResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Plan.Create(Plan{
		Amount:   500,
		Currency: "jpy",
		Interval: "month",
	})
	if transport.URL != "https://api.pay.jp/v1/plans" {
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
	} else if plan.Amount != 500 {
		t.Errorf("plan.Amount should be 500, but %d.", plan.Amount)
	}
}

func TestPlanRetrieve(t *testing.T) {
	mock, transport := NewMockClient(200, planResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Plan.Retrieve("pln_45dd3268a18b2837d52861716260")
	if transport.URL != "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260" {
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
	} else if plan.Amount != 500 {
		t.Errorf("parse error: plan.Amount should be 500, but %d.", plan.Amount)
	}
}

func TestPlanGetError(t *testing.T) {
	mock, _ := NewMockClient(200, planErrorResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Plan.Retrieve("pln_45dd3268a18b2837d52861716260")
	if err == nil {
		t.Error("err should not be nil")
	}
	if plan != nil {
		t.Errorf("plan should be nil, but %v", plan)
	}
}

func TestPlanUpdate(t *testing.T) {
	mock, transport := NewMockClient(200, planResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Plan.Update("pln_45dd3268a18b2837d52861716260", "new name")
	if transport.URL != "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260" {
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
	} else if plan.Amount != 500 {
		t.Errorf("parse error: plan.Amount should be 500, but %d.", plan.Amount)
	}
}

func TestPlanUpdate2(t *testing.T) {
	mock, transport := NewMockClient(200, planResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Plan.Retrieve("pln_45dd3268a18b2837d52861716260")
	if plan == nil {
		t.Error("plan should not be nil")
		return
	}
	err = plan.Update("new name")
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "POST" {
		t.Errorf("Method should be POST, but %s", transport.Method)
	}
}

func TestPlanDelete(t *testing.T) {
	mock, transport := NewMockClient(200, planResponseJSON)
	service := New("api-key", mock)
	err := service.Plan.Delete("pln_45dd3268a18b2837d52861716260")
	if transport.URL != "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "DELETE" {
		t.Errorf("Method should be DELETE, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
}

func TestPlanDelete2(t *testing.T) {
	mock, transport := NewMockClient(200, planResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Plan.Retrieve("pln_45dd3268a18b2837d52861716260")
	if plan == nil {
		t.Error("plan should not be nil")
		return
	}
	err = plan.Delete()
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transport.URL != "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "DELETE" {
		t.Errorf("Method should be DELETE, but %s", transport.Method)
	}
}

func TestPlanList(t *testing.T) {
	mock, transport := NewMockClient(200, planListResponseJSON)
	service := New("api-key", mock)
	plans, hasMore, err := service.Plan.List().
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/plans?limit=10&offset=15&since=1455328095&until=1455500895" {
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
