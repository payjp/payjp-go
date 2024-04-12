package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

var termJSONStr = `{
  "created": 1438354800,
  "id": "tm_b92b879e60f62b532d6756ae12af",
  "livemode": false,
  "object": "term",
  "charge_count": 158,
  "refund_count": 25,
  "dispute_count": 2,
  "end_at": null,
  "start_at": 1438354800
}`

var termResponseJSON = []byte(termJSONStr)

var termListJSONStr = `
{
  "count": 1,
  "data": [` + termJSONStr +
	`],
  "has_more": false,
  "object": "list",
  "url": "/v1/terms"
}`
var termListResponseJSON = []byte(termListJSONStr)

func TestParseTerm(t *testing.T) {
	service := &Service{}
	s := &TermResponse{
		service: service,
	}
	err := json.Unmarshal(termResponseJSON, s)

	assert.NoError(t, err)
	assert.Equal(t, "tm_b92b879e60f62b532d6756ae12af", s.ID)
	assert.False(t, s.LiveMode)
	assert.Equal(t, "term", s.Object)
	assert.Equal(t, 1438354800, *s.StartAt)
	assert.Nil(t, s.EndAt)
	assert.Equal(t, 158, s.ChargeCount)
	assert.Equal(t, 25, s.RefundCount)
	assert.Equal(t, 2, s.DisputeCount)
	assert.Equal(t, service, s.service)
}

func TestTermList(t *testing.T) {
	mock, transport := newMockClient(200, termListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	limit := Int(10)
	offset := Int(0)
	params := &TermListParams{
		Limit:        limit,
		Offset:       offset,
		SinceStartAt: Int(1455328095),
		UntilStartAt: Int(1455328095),
	}
	res, hasMore, err := service.Term.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/terms?limit=10&offset=0&since_start_at=1455328095&until_start_at=1455328095", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.False(t, hasMore)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, "tm_b92b879e60f62b532d6756ae12af", res[0].ID)
	assert.Equal(t, service, res[0].service)

	_, hasMore, err = service.Term.All()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestTerm(t *testing.T) {
	mock, transport := newMockClient(200, termResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	v, err := service.Term.Retrieve("tm_b92b879e60f62b532d6756ae12af")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/terms/tm_b92b879e60f62b532d6756ae12af", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, v)
	assert.Equal(t, "tm_b92b879e60f62b532d6756ae12af", v.ID)

	_, err = service.Term.Retrieve("tm_b92b879e60f62b532d6756ae12af")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
