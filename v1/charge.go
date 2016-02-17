package payjp

import (
	"encoding/json"
	"time"
)

type Charge struct {
	Amount         int              `json:"amount"`
	AmountRefunded int              `json:"amount_refunded"`
	Captured       bool             `json:"captured"`
	CapturedEpoch  int              `json:"captured_at"`
	Card           CardList         `json:"card"`
	CreatedEpoch   int              `json:"created"`
	Currency       string           `json:"currency"`
	Customer       string           `json:"customer"`
	Description    string           `json:"description"`
	ExpiredEpoch   int              `json:"expired_at"`
	FailureCode    int              `json:"failure_code"`
	FailureMessage string           `json:"failure_message"`
	ID             string           `json:"id"`
	LiveMode       bool             `json:"livemode"`
	Object         string           `json:"object"`
	Paid           bool             `json:"paid"`
	RefundReason   string           `json:"refund_reason"`
	Refunded       bool             `json:"refunded"`
	Subscription   SubscriptionList `json:"subscription"`

	CapturedAt time.Time
	CreatedAt  time.Time
	ExpiredAt  time.Time
}

type charge Charge

func (c *Charge) UnmarshalJSON(b []byte) error {
	raw := charge{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "charge" {
		*c = Charge(raw)
		c.CapturedAt = time.Unix(int64(raw.CapturedEpoch), 0)
		c.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		c.ExpiredAt = time.Unix(int64(raw.ExpiredEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

type ChargeList struct {
	Count   int      `json:"count"`
	Data    []Charge `json:"data"`
	HasMore bool     `json:"has_more"`
	Object  string   `json:"object"`
	URL     string   `json:"url"`
}
