package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var planResponseJSONStr = `{
  "amount": 500,
  "billing_day": null,
  "created": 1433127983,
  "currency": "jpy",
  "id": "pln_45dd3268a18b2837d52861716260",
  "interval": "month",
  "livemode": false,
  "metadata": {},
  "name": "name",
  "object": "plan",
  "trial_days": 30
}`
var planResponseJSON = []byte(planResponseJSONStr)
var planNewResponseJSON = []byte(`{
  "amount": 1000,
  "billing_day": null,
  "created": 1433127984,
  "currency": "jpy",
  "id": "pln_xxx",
  "interval": "month",
  "livemode": false,
  "metadata": {"hoge":"fuga"},
  "name": "new_name",
  "object": "plan",
  "trial_days": 30
}`)

var planListResponseJSON = []byte(`
{
  "count": 1,
  "data": [` + planResponseJSONStr +
	`],
  "object": "list",
  "has_more": true,
  "url": "/v1/customers/cus_4df4b5ed720933f4fb9e28857517/cards"
}
`)

func TestParsePlanResponseJSON(t *testing.T) {
	plan := &PlanResponse{}
	err := json.Unmarshal(planResponseJSON, plan)

	assert.NoError(t, err)
	assert.True(t, 1433127983 == plan.CreatedAt.Unix())
	assert.Equal(t, 0, plan.BillingDay)
}

func TestPlanCreate(t *testing.T) {
	mock, transport := newMockClient(200, planResponseJSON)
	service := New("api-key", mock)
	plan, err := service.Plan.Create(Plan{
		Amount:   500,
		Currency: "jpy",
		Interval: "month",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "amount=500&currency=jpy&interval=month", *transport.Body)
	assert.NotNil(t, plan)
	assert.Equal(t, 500, plan.Amount)
}

func TestPlanRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, planResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	plan, err := service.Plan.Retrieve("pln_req")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans/pln_req", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
	assert.Equal(t, 500, plan.Amount)

	plan, err = service.Plan.Retrieve("error")
	assert.Nil(t, plan)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestPlanUpdate(t *testing.T) {
	mock, transport := newMockClient(200, planResponseJSON)
	transport.AddResponse(200, planNewResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	plan, err := service.Plan.Update("pln_req", Plan{
		Name: "name",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans/pln_req", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "name=name", *transport.Body)
	assert.Equal(t, 500, plan.Amount)

	err = plan.Update(Plan{
		Name: "new_name",
		Metadata: map[string]string{
			"hoge": "fuga",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "name=new_name&metadata[hoge]=fuga", *transport.Body)
	assert.Equal(t, 1000, plan.Amount)

	err = plan.Update(Plan{})
	assert.Equal(t, "https://api.pay.jp/v1/plans/pln_xxx", transport.URL)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestPlanDelete(t *testing.T) {
	mock, transport := newMockClient(200, []byte(`{}`))
	service := New("api-key", mock)

	err := service.Plan.Delete("pln_45dd3268a18b2837d52861716260")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.Equal(t, "", transport.Header.Get("Content-Type"))
}

func TestPlanResponseDelete(t *testing.T) {
	mock, transport := newMockClient(200, planResponseJSON)
	transport.AddResponse(200, []byte(`{}`))
	service := New("api-key", mock)
	plan, err := service.Plan.Retrieve("pln_45dd3268a18b2837d52861716260")

	err = plan.Delete()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans/pln_45dd3268a18b2837d52861716260", transport.URL)
	assert.Equal(t, "DELETE", transport.Method)
	assert.NotNil(t, plan)
}

func TestPlanList(t *testing.T) {
	mock, transport := newMockClient(200, planListResponseJSON)
	transport.AddResponse(200, planListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	params := &PlanListParams{
		ListParams: ListParams{
			Limit:  Int(10),
			Offset: Int(0),
			Since:  Int(1455328095),
			Until:  Int(1455500895),
		},
	}
	plans, hasMore, err := service.Plan.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans?limit=10&offset=0&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(plans), 1)
	assert.Equal(t, 500, plans[0].Amount)

	plans, hasMore, err = service.Plan.List().
		Limit(1).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/plans?limit=1&offset=15&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(plans), 1)
	assert.Equal(t, 500, plans[0].Amount)

	_, hasMore, err = service.Plan.All()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
