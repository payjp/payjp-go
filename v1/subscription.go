package payjp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SubscriptionService struct {
	service *Service
}

func newSubscriptionService(service *Service) *SubscriptionService {
	return &SubscriptionService{
		service: service,
	}
}

type Subscription struct {
	TrialEndAt time.Time
	SkipTrial  interface{} // bool
	PlanID     interface{} // string
	Prorate    interface{} // bool
}

func (s SubscriptionService) Subscribe(customerID string, subscription Subscription) (*SubscriptionResponse, error) {
	var errors []string
	planID, ok := subscription.PlanID.(string)
	if !ok || planID == "" {
		errors = append(errors, "PlanID is required, but empty.")
	}
	var defaultTime time.Time
	skipTrial, ok := subscription.SkipTrial.(bool)
	if subscription.TrialEndAt != defaultTime && ok {
		errors = append(errors, "TrialEndAt and SkipTrial are exclusive.")
	}
	if len(errors) != 0 {
		return nil, fmt.Errorf("Subscription.Subscribe() parameter error: %s", strings.Join(errors, ", "))
	}
	qb := newRequestBuilder()
	qb.Add("customer", customerID)
	qb.Add("plan", subscription.PlanID)
	if subscription.TrialEndAt != defaultTime {
		qb.Add("trial_end", strconv.Itoa(int(subscription.TrialEndAt.Unix())))
	} else if ok && skipTrial {
		qb.Add("trial_end", "now")
	}
	qb.Add("prorate", subscription.Prorate)
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", s.service.apiKey)

	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) Get(customerID, subscriptionID string) (*SubscriptionResponse, error) {
	body, err := s.service.get("/customers/" + customerID + "/subscriptions/" + subscriptionID)
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) update(subscriptionID string, subscription Subscription) ([]byte, error) {
	var defaultTime time.Time
	_, ok := subscription.SkipTrial.(bool)
	if subscription.TrialEndAt != defaultTime && ok {
		return nil, errors.New("Subscription.Update() parameter error: TrialEndAt and SkipTrial are exclusive")
	}
	qb := newRequestBuilder()
	qb.Add("plan", subscription.PlanID)
	if subscription.TrialEndAt != defaultTime {
		qb.Add("trial_end", strconv.Itoa(int(subscription.TrialEndAt.Unix())))
	} else if subscription.SkipTrial == true {
		qb.Add("trial_end", "now")
	}
	qb.Add("prorate", subscription.Prorate)
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions/"+subscriptionID, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", s.service.apiKey)
	return parseResponseError(s.service.Client.Do(request))
}

func (s SubscriptionService) Update(subscriptionID string, subscription Subscription) (*SubscriptionResponse, error) {
	body, err := s.update(subscriptionID, subscription)
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) Pause(subscriptionID string) (*SubscriptionResponse, error) {
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions/"+subscriptionID+"/pause", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) Resume(subscriptionID string) (*SubscriptionResponse, error) {
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions/"+subscriptionID+"/resume", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) Cancel(subscriptionID string) (*SubscriptionResponse, error) {
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions/"+subscriptionID+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) Delete(subscriptionID string) error {
	request, err := http.NewRequest("DELETE", s.service.apiBase+"/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	_, err = parseResponseError(s.service.Client.Do(request))
	return err
}

func (s SubscriptionService) List() *subscriptionListCaller {
	return &subscriptionListCaller{
		service: s.service,
	}
}

func parseSubscription(service *Service, body []byte, result *SubscriptionResponse) (*SubscriptionResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	return result, nil
}

type SubscriptionResponse struct {
	CanceledAt           time.Time
	CreatedAt            time.Time
	CurrentPeriodEndAt   time.Time
	CurrentPeriodStartAt time.Time
	Customer             string
	ID                   string
	LiveMode             bool
	PausedAt             time.Time
	Plan                 Plan
	Prorate              bool
	ResumedAt            time.Time
	Start                int
	Status               string
	TrialEndAt           time.Time
	TrialStartAt         time.Time

	service *Service
}

type subscriptionResponseParser struct {
	CanceledEpoch           int             `json:"canceled_at"`
	CreatedEpoch            int             `json:"created"`
	CurrentPeriodEndEpoch   int             `json:"current_period_end"`
	CurrentPeriodStartEpoch int             `json:"current_period_start"`
	Customer                string          `json:"customer"`
	ID                      string          `json:"id"`
	LiveMode                bool            `json:"livemode"`
	Object                  string          `json:"object"`
	PausedEpoch             int             `json:"paused_at"`
	Plan                    json.RawMessage `json:"plan"`
	Prorate                 bool            `json:"prorate"`
	ResumedEpoch            int             `json:"resumed_at"`
	Start                   int             `json:"start"`
	Status                  string          `json:"status"`
	TrialEndEpoch           int             `json:"trial_end"`
	TrialStartEpoch         int             `json:"trial_start"`
}

func (s *SubscriptionResponse) Update(subscription Subscription) error {
	body, err := s.service.Subscription.update(s.ID, subscription)
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

func (s *SubscriptionResponse) Pause() error {
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions/"+s.ID+"/pause", nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

func (s *SubscriptionResponse) Resume() error {
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions/"+s.ID+"/resume", nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

func (s *SubscriptionResponse) Cancel() error {
	request, err := http.NewRequest("POST", s.service.apiBase+"/subscriptions/"+s.ID+"/cancel", nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

func (s *SubscriptionResponse) Delete() error {
	request, err := http.NewRequest("DELETE", s.service.apiBase+"/subscriptions/"+s.ID, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	_, err = parseResponseError(s.service.Client.Do(request))
	return err
}

func (s *SubscriptionResponse) UnmarshalJSON(b []byte) error {
	raw := subscriptionResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "subscription" {
		s.CanceledAt = time.Unix(int64(raw.CanceledEpoch), 0)
		s.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		s.CurrentPeriodEndAt = time.Unix(int64(raw.CurrentPeriodEndEpoch), 0)
		s.CurrentPeriodStartAt = time.Unix(int64(raw.CurrentPeriodStartEpoch), 0)
		s.Customer = raw.Customer
		s.ID = raw.ID
		s.LiveMode = raw.LiveMode
		s.PausedAt = time.Unix(int64(raw.PausedEpoch), 0)
		json.Unmarshal(raw.Plan, &s.Plan)
		s.Prorate = raw.Prorate
		s.ResumedAt = time.Unix(int64(raw.ResumedEpoch), 0)
		s.Start = raw.Start
		s.Status = raw.Status
		s.TrialEndAt = time.Unix(int64(raw.TrialEndEpoch), 0)
		s.TrialStartAt = time.Unix(int64(raw.TrialStartEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

type SubscriptionList struct {
	Count   int                     `json:"count"`
	Data    []*SubscriptionResponse `json:"data"`
	HasMore bool                    `json:"has_more"`
	Object  string                  `json:"object"`
	URL     string                  `json:"url"`
}

type subscriptionListCaller struct {
	service    *Service
	customerID string
	limit      int
	offset     int
	since      int
	until      int
}

func (c *subscriptionListCaller) Limit(limit int) *subscriptionListCaller {
	c.limit = limit
	return c
}

func (c *subscriptionListCaller) Offset(offset int) *subscriptionListCaller {
	c.offset = offset
	return c
}

func (c *subscriptionListCaller) Since(since time.Time) *subscriptionListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *subscriptionListCaller) Until(until time.Time) *subscriptionListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *subscriptionListCaller) Do() ([]*SubscriptionResponse, bool, error) {
	var url string
	if c.customerID == "" {
		url = "/subscriptions"
	} else {
		url = "/customers/" + c.customerID + "/subscriptions"
	}
	body, err := c.service.queryList(url, c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	result := &SubscriptionList{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, false, err
	}
	for _, customer := range result.Data {
		customer.service = c.service
	}
	return result.Data, result.HasMore, nil
}
