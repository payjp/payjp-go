package payjp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var threeDSecureRequestJSONStr = `{
  "created": 1730084767,
  "expired_at": null,
  "finished_at": null,
  "id": "tdsr_125192559c91c4011c1ff56f50a",
  "livemode": true,
  "object": "three_d_secure_request",
  "resource_id": "car_4ec110e0700daf893160424fe03c",
  "result_received_at": null,
  "started_at": null,
  "state": "created",
  "tenant_id": null,
  "three_d_secure_status": "unverified"
}`

var threeDSecureRequestResponseJSON = []byte(threeDSecureRequestJSONStr)

var threeDSecureRequestsListJSONStr = `
{
  "count": 1,
  "data": [` + threeDSecureRequestJSONStr +
	`],
  "has_more": false,
  "object": "list",
  "url": "/v1/three_d_secure_requests"
}`
var threeDSecureRequestsListResponseJSON = []byte(threeDSecureRequestsListJSONStr)

func TestParseThreeDSecureRequest(t *testing.T) {
	service := &Service{}
	s := &ThreeDSecureRequestResponse{
		service: service,
	}
	err := json.Unmarshal(threeDSecureRequestResponseJSON, s)

	assert.NoError(t, err)
	assert.Equal(t, "tdsr_125192559c91c4011c1ff56f50a", s.ID)
	assert.True(t, s.LiveMode)
	assert.Equal(t, "three_d_secure_request", s.Object)
	assert.Equal(t, "car_4ec110e0700daf893160424fe03c", s.ResourceID)
	assert.Nil(t, s.StartedAt)
	assert.Nil(t, s.ResultReceivedAt)
	assert.Nil(t, s.FinishedAt)
	assert.Nil(t, s.ExpiredAt)
	assert.Nil(t, s.TenantId)
	assert.Equal(t, "unverified", s.ThreeDSecureStatus)
	assert.Equal(t, service, s.service)
}

func TestThreeDSecureRequestList(t *testing.T) {
	mock, transport := newMockClient(200, threeDSecureRequestsListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	limit := Int(10)
	offset := Int(0)
	params := &ThreeDSecureRequestListParams{
		ListParams: ListParams{
			Limit:  limit,
			Offset: offset,
			Since:  Int(1455328095),
			Until:  Int(1455328095),
		},
	}
	res, hasMore, err := service.ThreeDSecureRequest.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/three_d_secure_requests?limit=10&offset=0&since=1455328095&until=1455328095", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.False(t, hasMore)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, "tdsr_125192559c91c4011c1ff56f50a", res[0].ID)
	assert.Equal(t, service, res[0].service)

	_, hasMore, err = service.ThreeDSecureRequest.All()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestThreeDSecureRequest(t *testing.T) {
	mock, transport := newMockClient(200, threeDSecureRequestResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	v, err := service.ThreeDSecureRequest.Retrieve("tdsr_125192559c91c4011c1ff56f50a")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/three_d_secure_requests/tdsr_125192559c91c4011c1ff56f50a", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, v)
	assert.Equal(t, "tdsr_125192559c91c4011c1ff56f50a", v.ID)

	_, err = service.ThreeDSecureRequest.Retrieve("tdsr_125192559c91c4011c1ff56f50a")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestThreeDSecureRequestCreate(t *testing.T) {
	mock, transport := newMockClient(200, threeDSecureRequestResponseJSON)
	transport.AddResponse(200, threeDSecureRequestResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	threeDSecureRequest, err := service.ThreeDSecureRequest.Create(ThreeDSecureRequest{
		ResourceID: "car_4ec110e0700daf893160424fe03c",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/three_d_secure_requests", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "resource_id=car_4ec110e0700daf893160424fe03c", *transport.Body)
	assert.NotNil(t, threeDSecureRequest)
	assert.Equal(t, "tdsr_125192559c91c4011c1ff56f50a", threeDSecureRequest.ID)
	assert.Equal(t, "car_4ec110e0700daf893160424fe03c", threeDSecureRequest.ResourceID)

	threeDSecureRequest, err = service.ThreeDSecureRequest.Create(ThreeDSecureRequest{
		ResourceID: "car_4ec110e0700daf893160424fe03c",
		TenantID:   "ten_xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	})
	assert.NoError(t, err)
	assert.Equal(t, "resource_id=car_4ec110e0700daf893160424fe03c&tenant_id=ten_xxxxxxxxxxxxxxxxxxxxxxxxxxxx", *transport.Body)
	assert.NotNil(t, threeDSecureRequest)

	_, err = service.ThreeDSecureRequest.Create(ThreeDSecureRequest{
		ResourceID: "car_4ec110e0700daf893160424fe03c",
	})
	assert.Equal(t, "resource_id=car_4ec110e0700daf893160424fe03c", *transport.Body)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
