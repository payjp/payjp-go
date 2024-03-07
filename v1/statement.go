package payjp

import (
	"encoding/json"
	"time"
)

type StatementService struct {
	service *Service
}

func newStatementService(service *Service) *StatementService {
	return &StatementService{
		service: service,
	}
}

type StatementUrls struct {
	Platformer *bool // プラットフォーム手数料に関する明細か否か
}

func (s *StatementResponse) StatementUrls(p ...StatementUrls) (*StatementUrlResponse, error) {
	qb := newRequestBuilder()
	if len(p) > 0 {
		qb.Add("platformer", p[0].Platformer)
	}

	body, err := s.service.request("POST", "/statements/"+s.ID+"/statement_urls", qb.Reader())
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

// Retrieve Statement object.
func (s StatementService) Retrieve(id string) (*StatementResponse, error) {
	body, err := s.service.request("GET", "/statements/"+id, nil)
	if err != nil {
		return nil, err
	}
	result := &StatementResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = s.service
	return result, nil
}

// StatementResponse
type StatementResponse struct {
	ID        string  `json:"id"`
	LiveMode  bool    `json:"livemode"`
	Object    string  `json:"object"`
	Title     *string `json:"title"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Items     []StatementItem

	Created  *int              `json:"created"`
	Updated  *int              `json:"updated"`
	RawItems []json.RawMessage `json:"items"`

	service *Service
}

// StatementUrlResponse
type StatementUrlResponse struct {
	Expires int    `json:"expires"`
	Object  string `json:"object"`
	URL     string `json:"url"`
}

// StatementItem
type StatementItem struct {
	Amount  int    `json:"amount"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
	TaxRate string `json:"tax_rate"`
}

func (s *StatementResponse) UnmarshalJSON(b []byte) error {
	type statementResponseParser StatementResponse
	var raw statementResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "statement" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.UpdatedAt = time.Unix(IntValue(raw.Updated), 0)
		raw.Items = make([]StatementItem, len(raw.RawItems))
		for i, r := range raw.RawItems {
			json.Unmarshal(r, &(raw.Items[i]))
		}
		raw.service = s.service
		*s = StatementResponse(raw)
		return nil
	}
	return parseError(b)
}

type StatementListParams struct {
	ListParams     `form:"*"`
	Owner          *string `form:"owner"`
	SourceTransfer *string `form:"source_transfer"`
	Tenant         *string `form:"tenant"`
}

func (c StatementService) All(params ...*StatementListParams) ([]*StatementResponse, bool, error) {
	p := &StatementListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/statements"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*StatementResponse, len(raw.Data))
	for i, rawS := range raw.Data {
		json.Unmarshal(rawS, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}
