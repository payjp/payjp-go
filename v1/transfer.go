package payjp

import (
	"encoding/json"
	"time"
)

type Transfer struct {
	Amount         int                `json:"amount"`
	CarriedBalance int                `json:"carried_balance"`
	Charges        listResponseParser `json:"charges"`
	CreatedEpoch   int                `json:"created"`
	Currency       string             `json:"currency"`
	Description    string             `json:"description"`
	ID             string             `json:"id"`
	LiveMode       bool               `json:"livemode"`
	Object         string             `json:"object"`
	ScheduledDate  string             `json:"scheduled_date"`
	Status         string             `json:"status"`
	Summary        struct {
		ChargeCount  int `json:"charge_count"`
		ChargeFee    int `json:"charge_fee"`
		ChargeGross  int `json:"charge_gross"`
		Net          int `json:"net"`
		RefundAmount int `json:"refund_amount"`
		RefundCount  int `json:"refund_count"`
	} `json:"summary"`
	TermEndEpoch   int    `json:"term_end"`
	TermStartEpoch int    `json:"term_start"`
	TransferAmount int    `json:"transfer_amount"`
	TransferDate   string `json:"transfer_date"`

	CreatedAt   time.Time
	TermEndAt   time.Time
	TermStartAt time.Time
}

type transfer Transfer

func (t *Transfer) UnmarshalJSON(b []byte) error {
	raw := transfer{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "transfer" {
		*t = Transfer(raw)
		t.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		t.TermEndAt = time.Unix(int64(raw.TermEndEpoch), 0)
		t.TermStartAt = time.Unix(int64(raw.TermStartEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
