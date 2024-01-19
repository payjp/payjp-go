package payjp

import (
	"encoding/json"
	"errors"
	"net/url"
	"time"
)

// EventType は、イベントのレスポンスの型を表す列挙型です。
// EventResponse.ResultTypeで種類を表すのに使用されます。
type EventType int

const (
	// ChargeEvent の場合はイベントに含まれるのがCharge型です
	ChargeEvent EventType = iota
	// TokenEvent の場合はイベントに含まれるのがToken型です
	TokenEvent
	// CustomerEvent の場合はイベントに含まれるのがCustomer型です
	CustomerEvent
	// CardEvent の場合はイベントに含まれるのがCard型です
	CardEvent
	// PlanEvent の場合はイベントに含まれるのがPlan型です
	PlanEvent
	// DeleteEvent の場合はイベントに含まれるのがDelete型です
	DeleteEvent
	// SubscriptionEvent の場合はイベントに含まれるのがSubscription型です
	SubscriptionEvent
	// TransferEvent の場合はイベントに含まれるのがTransfer型です
	TransferEvent
)

var eventTypes = map[string]EventType{
	"charge.succeeded":      ChargeEvent,
	"charge.failed":         ChargeEvent,
	"charge.updated":        ChargeEvent,
	"charge.refunded":       ChargeEvent,
	"charge.captured":       ChargeEvent,
	"token.create":          TokenEvent,
	"customer.created":      CustomerEvent,
	"customer.updated":      CustomerEvent,
	"customer.deleted":      DeleteEvent,
	"customer.card.created": CardEvent,
	"customer.card.updated": CardEvent,
	"customer.card.deleted": DeleteEvent,
	"plan.created":          PlanEvent,
	"plan.updated":          PlanEvent,
	"plan.deleted":          DeleteEvent,
	"subscription.created":  SubscriptionEvent,
	"subscription.updated":  SubscriptionEvent,
	"subscription.deleted":  DeleteEvent,
	"subscription.paused":   SubscriptionEvent,
	"subscription.resumed":  SubscriptionEvent,
	"subscription.canceled": SubscriptionEvent,
	"subscription.renewed":  SubscriptionEvent,
	"transfer.succeeded":    TransferEvent,
}

// EventService は作成、更新、削除などのイベントを表示するサービスです。
//
// イベント情報は、Webhookで任意のURLへ通知設定をすることができます。
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
	data, err := e.service.retrieve("/events/" + id)
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

// List はイベントリストを取得します。リストは、直近で生成された順番に取得されます。
func (e EventService) List() *EventListCaller {
	return &EventListCaller{
		service: e.service,
	}
}

// EventListCaller はイベントのリスト取得に使用する構造体です。
type EventListCaller struct {
	service    *Service
	limit      int
	offset     int
	resourceID string
	typeString string
	object     string
	since      int
	until      int
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (e *EventListCaller) Limit(limit int) *EventListCaller {
	e.limit = limit
	return e
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (e *EventListCaller) Offset(offset int) *EventListCaller {
	e.offset = offset
	return e
}

// ResourceID は取得するeventに紐づくAPIリソースのIDを設定します (e.g. customer.id)
func (e *EventListCaller) ResourceID(id string) *EventListCaller {
	e.resourceID = id
	return e
}

// Object は取得するeventに紐づくAPIリソースのオブジェクト名を設定します (e.g. customer, charge)
func (e *EventListCaller) Object(object string) *EventListCaller {
	e.object = object
	return e
}

// Type は取得するeventのtypeを設定します
func (e *EventListCaller) Type(typeString string) *EventListCaller {
	e.typeString = typeString
	return e
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (e *EventListCaller) Since(since time.Time) *EventListCaller {
	e.since = int(since.Unix())
	return e
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (e *EventListCaller) Until(until time.Time) *EventListCaller {
	e.until = int(until.Unix())
	return e
}

// Do は指定されたクエリーを元にイベントのリストを配列で取得します。
func (e *EventListCaller) Do() ([]*EventResponse, bool, error) {
	body, err := e.service.queryList("/events", e.limit, e.offset, e.since, e.until, func(values *url.Values) bool {
		hasParam := false
		if e.resourceID != "" {
			values.Set("resource_id", e.resourceID)
			hasParam = true
		}
		if e.object != "" {
			values.Set("object", e.object)
			hasParam = true
		}
		if e.typeString != "" {
			values.Set("type", e.typeString)
			hasParam = true
		}
		return hasParam
	})
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
		event := &EventResponse{}
		json.Unmarshal(rawPlan, event)
		result[i] = event
	}
	return result, raw.HasMore, nil
}

// EventResponse は、EventService.Retrieve()/EventService.List()が返す構造体です。
type EventResponse struct {
	CreatedAt       time.Time
	ID              string
	LiveMode        bool
	Type            string
	PendingWebHooks int
	ResultType      EventType

	data json.RawMessage
}

// ChargeData は、イベントの種類がChargeEventの時にChargeResponse構造体を返します。
func (e EventResponse) ChargeData() (*ChargeResponse, error) {
	if e.ResultType != ChargeEvent {
		return nil, errors.New("this event is not charge type")
	}
	result := &ChargeResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

// TokenData は、イベントの種類がTokenEventの時にTokenResponse構造体を返します。
func (e EventResponse) TokenData() (*TokenResponse, error) {
	if e.ResultType != TokenEvent {
		return nil, errors.New("this event is not token type")
	}
	result := &TokenResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

// CustomerData は、イベントの種類がCustomerEventの時にCustomerResponse構造体を返します。
func (e EventResponse) CustomerData() (*CustomerResponse, error) {
	if e.ResultType != CustomerEvent {
		return nil, errors.New("this event is not customer type")
	}
	result := &CustomerResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

// CardData は、イベントの種類がCardEventの時にCardResponse構造体を返します。
func (e EventResponse) CardData() (*CardResponse, error) {
	if e.ResultType != CardEvent {
		return nil, errors.New("this event is not card type")
	}
	result := &CardResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

// PlanData は、イベントの種類がPlanEventの時にPlanResponse構造体を返します。
func (e EventResponse) PlanData() (*PlanResponse, error) {
	if e.ResultType != PlanEvent {
		return nil, errors.New("this event is not plan type")
	}
	result := &PlanResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

// SubscriptionData は、イベントの種類がSubscriptionEventの時にSubscriptionResponse構造体を返します。
func (e EventResponse) SubscriptionData() (*SubscriptionResponse, error) {
	if e.ResultType != SubscriptionEvent {
		return nil, errors.New("this event is not subscription type")
	}
	result := &SubscriptionResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

// TransferData は、イベントの種類がTransferEventの時にTransferResponse構造体を返します。
func (e EventResponse) TransferData() (*TransferResponse, error) {
	if e.ResultType != TransferEvent {
		return nil, errors.New("this event is not tranfer type")
	}
	result := &TransferResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

// DeleteData は、イベントの種類がDeleteEventの時にDeleteResponse構造体を返します。
func (e EventResponse) DeleteData() (*DeleteResponse, error) {
	if e.ResultType != DeleteEvent {
		return nil, errors.New("this event is not delete type")
	}
	result := &DeleteResponse{}
	json.Unmarshal(e.data, result)
	return result, nil
}

type eventResponseParser struct {
	CreatedEpoch    int             `json:"created"`
	Data            json.RawMessage `json:"data"`
	ID              string          `json:"id"`
	LiveMode        bool            `json:"livemode"`
	Object          string          `json:"object"`
	PendingWebHooks int             `json:"pending_webhooks"`
	Type            string          `json:"type"`
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (e *EventResponse) UnmarshalJSON(b []byte) error {
	raw := eventResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "event" {
		e.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		e.data = raw.Data
		e.ID = raw.ID
		e.LiveMode = raw.LiveMode
		e.PendingWebHooks = raw.PendingWebHooks
		e.Type = raw.Type
		e.ResultType = eventTypes[raw.Type]

		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

// MarshalJSON はリクエストボディをAPI仕様のJson形式で返します。
func (e *EventResponse) MarshalJSON() ([]byte, error) {
	raw := eventResponseParser{}
	raw.Object = "event"
	raw.CreatedEpoch = int(e.CreatedAt.Unix())
	raw.Data = e.data
	raw.ID = e.ID
	raw.LiveMode = e.LiveMode
	raw.PendingWebHooks = e.PendingWebHooks
	raw.Type = e.Type
	return json.Marshal(&raw)
}

// DeleteResponse はイベントの種類がDeleteEventの時にDeleteData()が返す構造体です。
type DeleteResponse struct {
	Deleted  bool   `json:"deleted"`
	ID       string `json:"id"`
	LiveMode bool   `json:"livemode"`
}
