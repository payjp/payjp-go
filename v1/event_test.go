package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func eventResponseJSONStr(s string) string {
	return `
{
  "created": 1442288882,
  "data": ` + s + `,
  "id": "evnt_54db4d63c7886256acdbc784ccf",
  "livemode": false,
  "object": "event",
  "pending_webhooks": 1,
  "type": "customer.updated"
}`
}

var chargeEventResponseJSONStr = eventResponseJSONStr(chargeNewResponseJSONStr)
var chargeEventResponseJSON = []byte(chargeEventResponseJSONStr)
var customerEventResponseJSON = []byte(eventResponseJSONStr(customerResponseJSONStr))
var deletedEventResponseJSON = []byte(eventResponseJSONStr(deleteResponseJSONStr))

var eventListResponseJSON = []byte(`
{
  "count": 1,
  "data": [` + chargeEventResponseJSONStr + `
  ],
  "has_more": true,
  "object": "list",
  "url": "/v1/events"
}`)

func TestParseEventResponseJSON(t *testing.T) {
	service := &Service{}
	s := &EventResponse{
		service: service,
	}
	err := json.Unmarshal(customerEventResponseJSON, s)
	assert.NoError(t, err)
	assert.Equal(t, "evnt_54db4d63c7886256acdbc784ccf", s.ID)
	assert.False(t, s.LiveMode)
	assert.Equal(t, 1442288882, *s.Created)
	assert.IsType(t, time.Unix(0, 0), s.CreatedAt)
	assert.Equal(t, "customer.updated", s.Type)
	assert.Equal(t, 1, s.PendingWebHooks)
	assert.Equal(t, "event", s.Object)
	assert.Equal(t, service, s.service)
	v, ok := s.DataParser.(CustomerResponse)
	assert.True(t, ok)
	err = json.Unmarshal(s.Data, &v)
	assert.Equal(t, "cus_121673955bd7aa144de5a8f6c262", v.ID)
	assert.Equal(t, "cus_121673955bd7aa144de5a8f6c262", s.GetDataValue("id"))
	assert.Equal(t, "1", s.GetDataValue("cards", "count"))

	err = json.Unmarshal(chargeEventResponseJSON, s)
	assert.NoError(t, err)
	v2, ok := s.DataParser.(ChargeResponse)
	assert.True(t, ok)
	err = json.Unmarshal(s.Data, &v2)
	assert.Equal(t, "ch_6421ddf0e12a5e5641d7426f2a2c9", v2.ID)
	assert.Equal(t, "ch_6421ddf0e12a5e5641d7426f2a2c9", s.GetDataValue("id"))
	assert.Equal(t, "car_7a79b41fed704317ec0deb4ebf93", s.GetDataValue("card", "id"))

	err = json.Unmarshal(deletedEventResponseJSON, s)
	assert.NoError(t, err)
	v3, ok := s.DataParser.(DeleteResponse)
	assert.True(t, ok)
	err = json.Unmarshal(s.Data, &v3)
	assert.Equal(t, "xxx", v3.ID)
	assert.True(t, v3.Deleted)
	assert.Equal(t, "false", s.GetDataValue("livemode"))
}

func TestEventRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, customerEventResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	event, err := service.Event.Retrieve("evnt_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/events/evnt_xxx", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, event)

	event, err = service.Event.Retrieve("hoge")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	assert.Nil(t, event)
}

func TestEventList(t *testing.T) {
	mock, transport := newMockClient(200, eventListResponseJSON)
	transport.AddResponse(200, eventListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)
	e, hasMore, err := service.Event.List().
		Limit(10).
		Offset(0).
		Type("charge.updated").
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/events?limit=10&offset=0&since=1455328095&type=charge.updated&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(e), 1)
	assert.Equal(t, 1, e[0].PendingWebHooks)
	assert.Equal(t, service, e[0].service)

	params := &EventListParams{
		ListParams: ListParams{
			Limit:  Int(10),
			Offset: Int(0),
			Since:  Int(1455328095),
			Until:  Int(1455500895),
		},
		Type:       String("a"),
		ResourceID: String("b"),
		Object:     String("c"),
	}
	e, hasMore, err = service.Event.All(params)

	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/events?limit=10&object=c&offset=0&resource_id=b&since=1455328095&type=a&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(e), 1)
	assert.Equal(t, 1, e[0].PendingWebHooks)
	assert.Equal(t, service, e[0].service)

	_, hasMore, err = service.Charge.List().Do()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
