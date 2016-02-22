package payjp

import (
	"encoding/json"
	"net/url"
	"time"
)

type TransferStatus int

const (
	NoStatus TransferStatus = iota
	Pending
	Paid
	Failed
	Canceled
)

func (t TransferStatus) String() string {
	result := "unknown"
	switch t {
	case Pending:
		result = "pending"
	case Paid:
		result = "paid"
	case Failed:
		result = "failed"
	case Canceled:
		result = "canceled"
	}
	return result
}

type TransferService struct {
	service *Service
}

func newTransferService(service *Service) *TransferService {
	return &TransferService{
		service: service,
	}
}

func (t TransferService) Get(transferID string) (*TransferResponse, error) {
	body, err := t.service.get("/transfers/" + transferID)
	if err != nil {
		return nil, err
	}
	result := &TransferResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = t.service
	return result, nil
}

func (t TransferService) List() *transferListCaller {
	return &transferListCaller{
		status:  NoStatus,
		service: t.service,
	}
}

type transferListCaller struct {
	service *Service
	limit   int
	offset  int
	since   int
	until   int
	status  TransferStatus
}

func (c *transferListCaller) Limit(limit int) *transferListCaller {
	c.limit = limit
	return c
}

func (c *transferListCaller) Offset(offset int) *transferListCaller {
	c.offset = offset
	return c
}

func (c *transferListCaller) Since(since time.Time) *transferListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *transferListCaller) Until(until time.Time) *transferListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *transferListCaller) Status(status TransferStatus) *transferListCaller {
	c.status = status
	return c
}

func (c *transferListCaller) Do() ([]*TransferResponse, bool, error) {
	body, err := c.service.queryList("/transfers", c.limit, c.offset, c.since, c.until, func(values *url.Values) bool {
		if c.status != NoStatus {
			values.Add("status", c.status.String())
			return true
		}
		return false
	})
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*TransferResponse, len(raw.Data))
	for i, rawCharge := range raw.Data {
		charge := &TransferResponse{}
		json.Unmarshal(rawCharge, charge)
		charge.service = c.service
		result[i] = charge
	}
	return result, raw.HasMore, nil
}

func (t TransferService) ChargeList(transferID string) *transferChargeListCaller {
	return &transferChargeListCaller{
		service:    t.service,
		transferID: transferID,
	}
}

type transferChargeListCaller struct {
	service    *Service
	transferID string
	limit      int
	offset     int
	since      int
	until      int
	customerID string
}

func (c *transferChargeListCaller) Limit(limit int) *transferChargeListCaller {
	c.limit = limit
	return c
}

func (c *transferChargeListCaller) Offset(offset int) *transferChargeListCaller {
	c.offset = offset
	return c
}

func (c *transferChargeListCaller) Since(since time.Time) *transferChargeListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *transferChargeListCaller) Until(until time.Time) *transferChargeListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *transferChargeListCaller) CustomerID(ID string) *transferChargeListCaller {
	c.customerID = ID
	return c
}

func (c *transferChargeListCaller) Do() ([]*ChargeResponse, bool, error) {
	path := "/transfers/" + c.transferID + "/charges"
	body, err := c.service.queryList(path, c.limit, c.offset, c.since, c.until, func(values *url.Values) bool {
		if c.customerID != "" {
			values.Add("customer", c.customerID)
			return true
		}
		return false
	})
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
		transfer := &ChargeResponse{}
		json.Unmarshal(rawCharge, transfer)
		transfer.service = c.service
		result[i] = transfer
	}
	return result, raw.HasMore, nil
}

func (c TransferService) ListCharge(transferID string) *transferListCaller {
	return &transferListCaller{
		service: c.service,
	}
}

type TransferResponse struct {
	Amount         int
	CarriedBalance int
	Charges        []*ChargeResponse
	CreatedAt      time.Time
	Currency       string
	Description    string
	ID             string
	LiveMode       bool
	ScheduledDate  string
	Status         string
	Summary        struct {
		ChargeCount  int
		ChargeFee    int
		ChargeGross  int
		Net          int
		RefundAmount int
		RefundCount  int
	}
	TermEndAt      time.Time
	TermStartAt    time.Time
	TransferAmount int
	TransferDate   string

	service *Service
}

type transferResponseParser struct {
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
}

type transfer TransferResponse

func (t *TransferResponse) UnmarshalJSON(b []byte) error {
	raw := transferResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "transfer" {
		t.Amount = raw.Amount
		t.CarriedBalance = raw.CarriedBalance
		t.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		t.Currency = raw.Currency
		t.Description = raw.Description
		t.ID = raw.ID
		t.LiveMode = raw.LiveMode
		t.ScheduledDate = raw.ScheduledDate
		t.Status = raw.Status
		t.Summary.ChargeCount = raw.Summary.ChargeCount
		t.Summary.ChargeFee = raw.Summary.ChargeFee
		t.Summary.ChargeGross = raw.Summary.ChargeGross
		t.Summary.Net = raw.Summary.Net
		t.Summary.RefundAmount = raw.Summary.RefundAmount
		t.Summary.RefundCount = raw.Summary.RefundCount
		t.TermEndAt = time.Unix(int64(raw.TermEndEpoch), 0)
		t.TermStartAt = time.Unix(int64(raw.TermStartEpoch), 0)
		t.TransferAmount = raw.TransferAmount
		t.TransferDate = raw.TransferDate
		t.Charges = make([]*ChargeResponse, len(raw.Charges.Data))
		for i, rawCharge := range raw.Charges.Data {
			charge := &ChargeResponse{}
			json.Unmarshal(rawCharge, charge)
			t.Charges[i] = charge
		}

		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
