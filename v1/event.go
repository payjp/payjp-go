package payjp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// EventService は作成、更新、削除などのイベントを表示するサービスです。
type EventService struct {
	service *Service
}

func newEventService(service *Service) *EventService {
	return &EventService{
		service: service,
	}
}

// Retrieve event object. 特定のイベント情報を取得します。
func (e EventService) Retrieve(id string) (*EventResponse, error) {
	data, err := e.service.request("GET", "/events/"+id, nil)
	if err != nil {
		return nil, err
	}
	result := &EventResponse{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e EventService) All(params ...*EventListParams) ([]*EventResponse, bool, error) {
	p := &EventListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := e.service.request("GET", "/events"+e.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*EventResponse, len(raw.Data))
	for i, rawPlan := range raw.Data {
		json.Unmarshal(rawPlan, &result[i])
		result[i].service = e.service
	}
	return result, raw.HasMore, nil
}

// List はイベントリストを取得します。リストは、直近で生成された順番に取得されます。
func (c EventService) List() *EventListCaller {
	p := &EventListParams{}
	return &EventListCaller{
		service:         c,
		EventListParams: *p,
	}
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (e *EventListCaller) Limit(limit int) *EventListCaller {
	e.EventListParams.ListParams.Limit = &limit
	return e
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (e *EventListCaller) Offset(offset int) *EventListCaller {
	e.EventListParams.ListParams.Offset = &offset
	return e
}

// ResourceID は取得するeventに紐づくAPIリソースのIDを設定します (e.g. customer.id)
func (e *EventListCaller) ResourceID(id string) *EventListCaller {
	e.EventListParams.ResourceID = &id
	return e
}

// Object は取得するeventに紐づくAPIリソースのオブジェクト名を設定します (e.g. customer, charge)
func (e *EventListCaller) Object(object string) *EventListCaller {
	e.EventListParams.Object = &object
	return e
}

// Type は取得するeventのtypeを設定します
func (e *EventListCaller) Type(p string) *EventListCaller {
	e.EventListParams.Type = &p
	return e
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (e *EventListCaller) Since(since time.Time) *EventListCaller {
	p := int(since.Unix())
	e.EventListParams.ListParams.Since = &p
	return e
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (e *EventListCaller) Until(until time.Time) *EventListCaller {
	p := int(until.Unix())
	e.EventListParams.ListParams.Until = &p
	return e
}

type EventListParams struct {
	ListParams `form:"*"`
	ResourceID *string `form:"resource_id"`
	Type       *string `form:"type"`
	Object     *string `form:"object"`
}

type EventListCaller struct {
	service EventService
	EventListParams
}

func (e *EventListCaller) Do() ([]*EventResponse, bool, error) {
	return e.service.All(&e.EventListParams)
}

// EventResponse は、EventService.Retrieve()/EventService.List()が返す構造体です。
type EventResponse struct {
	CreatedAt       time.Time
	Created         *int            `json:"created"` // この支払い作成時のタイムスタンプ
	ID              string          `json:"id"`
	LiveMode        bool            `json:"livemode"`
	Type            string          `json:"type"`
	PendingWebHooks int             `json:"pending_webhooks"`
	Object          string          `json:"object"`
	Data            json.RawMessage `json:"data"`
	DataMap         map[string]interface{}
	DataParser      interface{}

	service *Service
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (e *EventResponse) UnmarshalJSON(b []byte) error {
	type eventResponseParser EventResponse
	var raw eventResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "event" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		data := payjpResponse{}
		err = json.Unmarshal(raw.Data, &data)
		raw.DataParser = data.Parser
		err = json.Unmarshal(raw.Data, &raw.DataMap)

		raw.service = e.service
		*e = EventResponse(raw)
		return nil
	}
	return parseError(b)
}

func (e *EventResponse) GetDataValue(keys ...string) string {
	return getValue(e.DataMap, keys)
}

// from https://github.com/stripe/stripe-go/releases/tag/v76.14.0
func getValue(m map[string]interface{}, keys []string) string {
	node := m[keys[0]]

	for i := 1; i < len(keys); i++ {
		key := keys[i]

		sliceNode, ok := node.([]interface{})
		if ok {
			intKey, err := strconv.Atoi(key)
			if err != nil {
				panic(fmt.Sprintf(
					"Cannot access nested slice element with non-integer key: %s",
					key))
			}
			node = sliceNode[intKey]
			continue
		}

		mapNode, ok := node.(map[string]interface{})
		if ok {
			node = mapNode[key]
			continue
		}

		panic(fmt.Sprintf(
			"Cannot descend into non-map non-slice object with key: %s", key))
	}

	if node == nil {
		return ""
	}

	return fmt.Sprintf("%v", node)
}
