package payjp

import (
	"encoding/json"
)

type TermService struct {
	service *Service
}

func newTermService(service *Service) *TermService {
	return &TermService{
		service: service,
	}
}

func (s TermService) Retrieve(id string) (*TermResponse, error) {
	body, err := s.service.request("GET", "/terms/"+id, nil)
	if err != nil {
		return nil, err
	}
	result := &TermResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = s.service
	return result, nil
}

type TermResponse struct {
	ID           string `json:"id"`
	LiveMode     bool   `json:"livemode"`
	Object       string `json:"object"`
	ChargeCount  int    `json:"charge_count"`
	RefundCount  int    `json:"refund_count"`
	DisputeCount int    `json:"dispute_count"`
	StartAt      *int   `json:"start_at"`
	EndAt        *int   `json:"end_at"`

	service *Service
}

func (s *TermResponse) UnmarshalJSON(b []byte) error {
	type termResponseParser TermResponse
	var raw termResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "term" {
		raw.service = s.service
		*s = TermResponse(raw)
		return nil
	}
	return parseError(b)
}

type TermListParams struct {
	Limit        *int `form:"limit"`
	Offset       *int `form:"offset"`
	SinceStartAt *int `form:"since_start_at"`
	UntilStartAt *int `form:"until_start_at"`
}

func (c TermService) All(params ...*TermListParams) ([]*TermResponse, bool, error) {
	p := &TermListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/terms"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*TermResponse, len(raw.Data))
	for i, rawS := range raw.Data {
		json.Unmarshal(rawS, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}
