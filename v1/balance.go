package payjp

import (
	"encoding/json"
	"time"
)

type BalanceListParams struct {
	ListParams   `form:"*"`
	SinceDueDate *int    `form:"since_due_date"`
	UntilDueDate *int    `form:"until_due_date"`
	State        *string `form:"state"`
	Closed       *bool   `form:"closed"`
	Owner        *string `form:"owner"`
	Tenant       *string `form:"tenant"`
}

type BalanceService struct {
	service *Service
}

func newBalanceService(service *Service) *BalanceService {
	return &BalanceService{
		service: service,
	}
}

func (s *BalanceResponse) StatementUrls(p ...StatementUrls) (*StatementUrlResponse, error) {
	qb := newRequestBuilder()
	if len(p) > 0 {
		qb.Add("platformer", p[0].Platformer)
	}

	body, err := s.service.request("POST", "/balances/"+s.ID+"/statement_urls", qb.Reader())
	if err != nil {
		return nil, err
	}
	result := &StatementUrlResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Retrieve transfer object. 入金情報を取得します。
func (t BalanceService) Retrieve(id string) (*BalanceResponse, error) {
	body, err := t.service.request("GET", "/balances/"+id, nil)
	if err != nil {
		return nil, err
	}
	result := &BalanceResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = t.service
	return result, nil
}

func (c BalanceService) All(params ...*BalanceListParams) ([]*BalanceResponse, bool, error) {
	p := &BalanceListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/balances"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*BalanceResponse, len(raw.Data))
	for i, raw := range raw.Data {
		json.Unmarshal(raw, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}

type BankInfo struct {
	BankCode              string `json:"bank_code"`
	BankBranchCode        string `json:"bank_branch_code"`
	BankAccountType       string `json:"bank_account_type"`
	BankAccountNumber     string `json:"bank_account_number"`
	BankAccountHolderName string `json:"bank_account_holder_name"`
	BankAccountStatus     string `json:"bank_account_status"`
}

type BalanceResponse struct {
	ID            string `json:"id"`
	LiveMode      bool   `json:"livemode"`
	Created       *int   `json:"created"`
	CreatedAt     time.Time
	Net           int64  `json:"net"`
	Type          string `json:"type"`
	Closed        bool   `json:"closed"`
	Statements    []*StatementResponse
	RawStatements listResponseParser `json:"statements"`
	DueDate       string             `json:"due_date"`
	BankInfo      *BankInfo          `json:"bank_info"`
	Object        string             `json:"object"`

	service *Service
}

func (t *BalanceResponse) UnmarshalJSON(b []byte) error {
	type balaceResponseParser BalanceResponse
	var raw balaceResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "balance" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.Statements = make([]*StatementResponse, len(raw.RawStatements.Data))
		for i, r := range raw.RawStatements.Data {
			json.Unmarshal(r, &(raw.Statements[i]))
		}
		raw.service = t.service
		*t = BalanceResponse(raw)

		return nil
	}
	return parseError(b)
}
