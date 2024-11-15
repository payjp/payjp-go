package payjp

import (
	"encoding/json"
	"time"
)

type ThreeDSecureRequestService struct {
	service *Service
}

func newThreeDSecureRequestService(service *Service) *ThreeDSecureRequestService {
	return &ThreeDSecureRequestService{
		service: service,
	}
}

// ThreeDSecureRequest は3Dセキュアリクエストの作成時に使用する構造体です。
type ThreeDSecureRequest struct {
	ResourceID string  // 必須: 顧客カードID。
	TenantID   *string // テナントID
}

func (t ThreeDSecureRequestService) Create(threeDSecureRequest ThreeDSecureRequest) (*ThreeDSecureRequestResponse, error) {
	qb := newRequestBuilder()
	qb.Add("resource_id", threeDSecureRequest.ResourceID)
	qb.Add("tenant_id", threeDSecureRequest.TenantID)

	body, err := t.service.request("POST", "/three_d_secure_requests", qb.Reader())
	if err != nil {
		return nil, err
	}
	return parseThreeDSecureRequest(t.service, body, &ThreeDSecureRequestResponse{})
}

func parseThreeDSecureRequest(service *Service, body []byte, result *ThreeDSecureRequestResponse) (*ThreeDSecureRequestResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	return result, nil
}

func (s ThreeDSecureRequestService) Retrieve(id string) (*ThreeDSecureRequestResponse, error) {
	body, err := s.service.request("GET", "/three_d_secure_requests/"+id, nil)
	if err != nil {
		return nil, err
	}
	result := &ThreeDSecureRequestResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = s.service
	return result, nil
}

type ThreeDSecureRequestResponse struct {
	ID                 string `json:"id"`
	ResourceID         string `json:"resource_id"`
	Object             string `json:"object"`
	LiveMode           bool   `json:"livemode"`
	Created            *int   `json:"created"`
	CreatedAt          time.Time
	State              string  `form:"state"`
	StartedAt          *int    `json:"started_at"`
	ResultReceivedAt   *int    `json:"result_received_at"`
	FinishedAt         *int    `json:"finished_at"`
	ExpiredAt          *int    `json:"expired_at"`
	TenantId           *string `json:"tenant_id"`
	ThreeDSecureStatus string  `json:"three_d_secure_status"`
	service            *Service
}

func (s *ThreeDSecureRequestResponse) UnmarshalJSON(b []byte) error {
	type threeDSecureRequestResponseParser ThreeDSecureRequestResponse
	var raw threeDSecureRequestResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "three_d_secure_request" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.service = s.service
		*s = ThreeDSecureRequestResponse(raw)
		return nil
	}
	return parseError(b)
}

type ThreeDSecureRequestListParams struct {
	ListParams `form:"*"`
	ResourceID *string `form:"resource_id"`
	TenantID   *string `form:"tenant_id"`
}

func (c ThreeDSecureRequestService) All(params ...*ThreeDSecureRequestListParams) ([]*ThreeDSecureRequestResponse, bool, error) {
	p := &ThreeDSecureRequestListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/three_d_secure_requests"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*ThreeDSecureRequestResponse, len(raw.Data))
	for i, rawS := range raw.Data {
		json.Unmarshal(rawS, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}
