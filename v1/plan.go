package payjp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type planService struct {
	service *Service
}

func newPlanService(service *Service) *planService {
	return &planService{
		service: service,
	}
}

func (p planService) Create(plan Plan) (*PlanResponse, error) {
	var errors []string
	if plan.Amount < 50 || plan.Amount > 9999999 {
		errors = append(errors, fmt.Sprintf("Plan.Amount should be between 50 and 9,999,999, but %d.", plan.Amount))
	}
	if plan.Currency == "" {
		plan.Currency = "jpy"
	} else if plan.Currency != "jpy" {
		// todo: if pay.jp supports other currency, fix this condition
		errors = append(errors, fmt.Sprintf("payjp.Plan.Create() only supports 'jpy' as currency, but '%s'.", plan.Currency))
	}
	if plan.Interval == "" {
		plan.Interval = "month"
	} else if plan.Interval != "month" {
		// todo: if pay.jp supports other interval options, fix this condition
		errors = append(errors, fmt.Sprintf("payjp.Plan.Create() only supports 'month' as interval, but '%s'.", plan.Interval))
	}
	if plan.BillingDay < 0 || plan.BillingDay > 31 {
		errors = append(errors, fmt.Sprintf("Plan.BillingDay should be between 1 and 31, but %d.", plan.BillingDay))
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

	resp, err := p.service.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &PlanResponse{
		service: p.service,
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p planService) Get(id string) (*PlanResponse, error) {
	body, err := p.service.get("/plans/" + id)
	if err != nil {
		return nil, err
	}
	result := &PlanResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = p.service
	return result, nil
}

func (p planService) update(id, name string) ([]byte, error) {
	qb := newRequestBuilder()
	qb.Add("name=%s", name)
	request, err := http.NewRequest("POST", p.service.apiBase+"/plans/"+id, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", p.service.apiKey)

	return parseResponseError(p.service.Client.Do(request))
}

func (p planService) Update(id, name string) (*PlanResponse, error) {
	body, err := p.update(id, name)
	if err != nil {
		return nil, err
	}
	result := &PlanResponse{
		service: p.service,
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p planService) Delete(id string) error {
	return p.service.delete("/plans/" + id)
}

func (p planService) List() *planListCaller {
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
	result := &PlanListResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, false, err
	}
	for _, plan := range result.Data {
		plan.service = c.service
	}
	return result.Data, result.HasMore, nil
}

type PlanResponse struct {
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
	CreatedAt    time.Time
	service      *Service
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

type planResponse PlanResponse

func (p *PlanResponse) UnmarshalJSON(b []byte) error {
	raw := planResponse{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "plan" {
		*p = PlanResponse(raw)
		p.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

type PlanListResponse struct {
	Count   int             `json:"count"`
	Data    []*PlanResponse `json:"data"`
	HasMore bool            `json:"has_more"`
	Object  string          `json:"object"`
	URL     string          `json:"url"`
}

type planList PlanListResponse

func (p *PlanListResponse) UnmarshalJSON(b []byte) error {
	raw := planList{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "list" {
		*p = PlanListResponse(raw)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
