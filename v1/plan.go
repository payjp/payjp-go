package payjp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Plan struct {
	Amount     int
	Currency   string
	Interval   string
	ID         string
	Name       string
	TrialDays  int
	BillingDay int
}

type PlanService struct {
	service *Service
}

func newPlanService(service *Service) *PlanService {
	return &PlanService{
		service: service,
	}
}

func (p PlanService) Create(plan Plan) (*PlanResponse, error) {
	var errors []string
	if plan.Amount < 50 || plan.Amount > 9999999 {
		errors = append(errors, fmt.Sprintf("Amount should be between 50 and 9,999,999, but %d.", plan.Amount))
	}
	if plan.Currency == "" {
		plan.Currency = "jpy"
	} else if plan.Currency != "jpy" {
		// todo: if pay.jp supports other currency, fix this condition
		errors = append(errors, fmt.Sprintf("Only supports 'jpy' as currency, but '%s'.", plan.Currency))
	}
	if plan.Interval == "" {
		plan.Interval = "month"
	} else if plan.Interval != "month" {
		// todo: if pay.jp supports other interval options, fix this condition
		errors = append(errors, fmt.Sprintf("Only supports 'month' as interval, but '%s'.", plan.Interval))
	}
	if plan.BillingDay < 0 || plan.BillingDay > 31 {
		errors = append(errors, fmt.Sprintf("BillingDay should be between 1 and 31, but %d.", plan.BillingDay))
	}
	if len(errors) != 0 {
		return nil, fmt.Errorf("payjp.Plan.Create() parameter error: %s", strings.Join(errors, ", "))
	}
	qb := newRequestBuilder()
	qb.Add("amount", strconv.Itoa(plan.Amount))
	qb.Add("currency", plan.Currency)
	qb.Add("interval", plan.Interval)
	if plan.ID != "" {
		qb.Add("id", plan.ID)
	}
	if plan.Name != "" {
		qb.Add("name", plan.Name)
	}
	if plan.TrialDays != 0 {
		qb.Add("trial_days", strconv.Itoa(plan.TrialDays))
	}
	if plan.BillingDay != 0 {
		qb.Add("billing_day", strconv.Itoa(plan.BillingDay))
	}
	request, err := http.NewRequest("POST", p.service.apiBase+"/plans", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", p.service.apiKey)

	body, err := respToBody(p.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parsePlan(p.service, body, &PlanResponse{})
}

func (p PlanService) Get(id string) (*PlanResponse, error) {
	body, err := p.service.get("/plans/" + id)
	if err != nil {
		return nil, err
	}
	return parsePlan(p.service, body, &PlanResponse{})
}

func parsePlan(service *Service, body []byte, result *PlanResponse) (*PlanResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	return result, nil
}

func (p PlanService) update(id, name string) ([]byte, error) {
	qb := newRequestBuilder()
	qb.Add("name", name)
	request, err := http.NewRequest("POST", p.service.apiBase+"/plans/"+id, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", p.service.apiKey)

	return parseResponseError(p.service.Client.Do(request))
}

func (p PlanService) Update(id, name string) (*PlanResponse, error) {
	body, err := p.update(id, name)
	if err != nil {
		return nil, err
	}
	return parsePlan(p.service, body, &PlanResponse{})
}

func (p PlanService) Delete(id string) error {
	return p.service.delete("/plans/" + id)
}

func (p PlanService) List() *planListCaller {
	return &planListCaller{
		service: p.service,
	}
}

type planListCaller struct {
	service *Service
	limit   int
	offset  int
	since   int
	until   int
}

func (c *planListCaller) Limit(limit int) *planListCaller {
	c.limit = limit
	return c
}

func (c *planListCaller) Offset(offset int) *planListCaller {
	c.offset = offset
	return c
}

func (c *planListCaller) Since(since time.Time) *planListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *planListCaller) Until(until time.Time) *planListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *planListCaller) Do() ([]*PlanResponse, bool, error) {
	body, err := c.service.queryList("/plans", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*PlanResponse, len(raw.Data))
	for i, rawPlan := range raw.Data {
		plan := &PlanResponse{}
		json.Unmarshal(rawPlan, plan)
		plan.service = c.service
		result[i] = plan
	}
	return result, raw.HasMore, nil
}

type PlanResponse struct {
	Amount     int
	BillingDay int
	Currency   string
	ID         string
	Interval   string
	LiveMode   bool
	Name       string
	TrialDays  int
	CreatedAt  time.Time
	service    *Service
}

type planResponseParser struct {
	Amount       int    `json:"amount"`
	BillingDay   int    `json:"billing_day"`
	CreatedEpoch int    `json:"created"`
	Currency     string `json:"currency"`
	ID           string `json:"id"`
	Interval     string `json:"interval"`
	LiveMode     bool   `json:"livemode"`
	Name         string `json:"name"`
	Object       string `json:"object"`
	TrialDays    int    `json:"trial_days"`
}

func (p *PlanResponse) Update(name string) error {
	body, err := p.service.Plan.update(p.ID, name)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, p)
}

func (p *PlanResponse) Delete() error {
	return p.service.Plan.Delete(p.ID)
}

func (p *PlanResponse) UnmarshalJSON(b []byte) error {
	raw := planResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "plan" {
		p.Amount = raw.Amount
		p.BillingDay = raw.BillingDay
		p.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		p.Currency = raw.Currency
		p.ID = raw.ID
		p.Interval = raw.Interval
		p.LiveMode = raw.LiveMode
		p.Name = raw.Name
		p.TrialDays = raw.TrialDays
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
