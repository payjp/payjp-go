package payjp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ChargeService struct {
	service *Service
}

func newChargeService(service *Service) *ChargeService {
	return &ChargeService{
		service: service,
	}
}

type Charge struct {
	Currency    string
	CustomerID  string
	Card        Card
	CardToken   string
	Capture     bool
	Description string
	ExpireDays  int
}

func (c ChargeService) Create(amount int, charge Charge) (*ChargeResponse, error) {
	var errorMessages []string
	if amount < 50 || amount > 9999999 {
		errorMessages = append(errorMessages, fmt.Sprintf("Amount should be between 50 and 9,999,999, but %d.", amount))
	}
	counter := 0
	if charge.CustomerID != "" {
		counter++
	}
	if charge.CardToken != "" {
		counter++
	}
	if charge.Card.Valid() {
		counter++
	}
	switch counter {
	case 0:
		errorMessages = append(errorMessages, "One of the following parameters is required: CustomerID, CardToken, Card")
	case 1:
	case 2, 3:
		errorMessages = append(errorMessages, "The following parameters are exclusive: CustomerID, CardToken, Card")
	}
	if charge.Currency == "" {
		charge.Currency = "jpy"
	} else if charge.Currency != "jpy" {
		// todo: if pay.jp supports other currency, fix this condition
		errorMessages = append(errorMessages, fmt.Sprintf("Only supports 'jpy' as currency, but '%s'.", charge.Currency))
	}
	if len(errorMessages) > 0 {
		return nil, fmt.Errorf("Charge.Create() parameter error: %s", strings.Join(errorMessages, ", "))
	}
	qb := newRequestBuilder()
	qb.Add("amount", amount)
	qb.Add("currency", charge.Currency)
	qb.Add("customer", charge.CustomerID)
	qb.Add("card", charge.CardToken)
	qb.AddCard(charge.Card)
	qb.Add("description", charge.Description)
	qb.Add("capture", charge.Capture)
	qb.Add("expiry_days", charge.ExpireDays)

	request, err := http.NewRequest("POST", c.service.apiBase+"/charges", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	body, err := respToBody(c.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})

}

func (c ChargeService) Get(chargeID string) (*ChargeResponse, error) {
	body, err := c.service.get("/charges/" + chargeID)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) update(chargeID, description string) ([]byte, error) {
	qb := newRequestBuilder()
	qb.Add("name", description)
	request, err := http.NewRequest("POST", c.service.apiBase+"/charges/"+chargeID, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

func (c ChargeService) Update(chargeID, description string) (*ChargeResponse, error) {
	body, err := c.update(chargeID, description)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) refund(id string) ([]byte, error) {
	request, err := http.NewRequest("POST", c.service.apiBase+"/charges/"+id+"/refund", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

func (c ChargeService) Refund(chargeID string) (*ChargeResponse, error) {
	body, err := c.refund(chargeID)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) capture(chargeID string) ([]byte, error) {
	request, err := http.NewRequest("POST", c.service.apiBase+"/charges/"+chargeID+"/capture", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

func (c ChargeService) Capture(chargeID string) (*ChargeResponse, error) {
	body, err := c.capture(chargeID)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) List() *chargeListCaller {
	return &chargeListCaller{
		service: c.service,
	}
}

type chargeListCaller struct {
	service *Service
	limit   int
	offset  int
	since   int
	until   int
}

func (c *chargeListCaller) Limit(limit int) *chargeListCaller {
	c.limit = limit
	return c
}

func (c *chargeListCaller) Offset(offset int) *chargeListCaller {
	c.offset = offset
	return c
}

func (c *chargeListCaller) Since(since time.Time) *chargeListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *chargeListCaller) Until(until time.Time) *chargeListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *chargeListCaller) Do() ([]*ChargeResponse, bool, error) {
	body, err := c.service.queryList("/charges", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*ChargeResponse, len(raw.Data))
	for i, rawCharge := range raw.Data {
		charge := &ChargeResponse{}
		json.Unmarshal(rawCharge, charge)
		charge.service = c.service
		result[i] = charge
	}
	return result, raw.HasMore, nil
}

func parseCharge(service *Service, data []byte, result *ChargeResponse) (*ChargeResponse, error) {
	err := json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	return result, nil
}

type ChargeResponse struct {
	Amount         int
	AmountRefunded int
	Captured       bool
	CapturedAt     time.Time
	Card           CardResponse
	CreatedAt      time.Time
	Currency       string
	Customer       string
	Description    string
	ExpiredAt      time.Time
	FailureCode    int
	FailureMessage string
	ID             string
	LiveMode       bool
	Paid           bool
	RefundReason   string
	Refunded       bool
	Subscription   string

	service *Service
}

func (c *ChargeResponse) Update(description string) error {
	body, err := c.service.Charge.update(c.ID, description)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

func (c *ChargeResponse) Refund() error {
	body, err := c.service.Charge.refund(c.ID)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

func (c *ChargeResponse) Capture() error {
	body, err := c.service.Charge.capture(c.ID)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

type chargeResponseParser struct {
	Amount         int             `json:"amount"`
	AmountRefunded int             `json:"amount_refunded"`
	Captured       bool            `json:"captured"`
	CapturedEpoch  int             `json:"captured_at"`
	Card           json.RawMessage `json:"card"`
	CreatedEpoch   int             `json:"created"`
	Currency       string          `json:"currency"`
	Customer       string          `json:"customer"`
	Description    string          `json:"description"`
	ExpiredEpoch   int             `json:"expired_at"`
	FailureCode    int             `json:"failure_code"`
	FailureMessage string          `json:"failure_message"`
	ID             string          `json:"id"`
	LiveMode       bool            `json:"livemode"`
	Object         string          `json:"object"`
	Paid           bool            `json:"paid"`
	RefundReason   string          `json:"refund_reason"`
	Refunded       bool            `json:"refunded"`
	Subscription   string          `json:"subscription"`
}

func (c *ChargeResponse) UnmarshalJSON(b []byte) error {
	raw := chargeResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "charge" {
		c.Amount = raw.Amount
		c.AmountRefunded = raw.AmountRefunded
		c.Captured = raw.Captured
		c.CapturedAt = time.Unix(int64(raw.CapturedEpoch), 0)
		json.Unmarshal(raw.Card, &c.Card)
		c.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		c.Currency = raw.Currency
		c.Customer = raw.Customer
		c.Description = raw.Description
		c.ExpiredAt = time.Unix(int64(raw.ExpiredEpoch), 0)
		c.FailureCode = raw.FailureCode
		c.FailureMessage = raw.FailureMessage
		c.ID = raw.ID
		c.LiveMode = raw.LiveMode
		c.Paid = raw.Paid
		c.RefundReason = raw.RefundReason
		c.Refunded = raw.Refunded
		c.Subscription = raw.Subscription
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}
	return nil
}
