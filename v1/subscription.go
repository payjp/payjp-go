package payjp

import (
	"encoding/json"
	"time"
)

// SubscriptionStatus は定期購読のステータスを表すEnumです。
type SubscriptionStatus string

const (
	// SubscriptionActive はアクティブ状態を表す定数
	SubscriptionActive = SubscriptionStatus("active")
	// SubscriptionTrial はトライアル状態を表す定数
	SubscriptionTrial = SubscriptionStatus("trial")
	// SubscriptionCanceled はキャンセル状態を表す定数
	SubscriptionCanceled = SubscriptionStatus("canceled")
	// SubscriptionPaused は停止状態を表す定数
	SubscriptionPaused = SubscriptionStatus("paused")
)

func (s SubscriptionStatus) String() string {
	return string(s)
}

// SubscriptionService は月単位で定期的な支払い処理を行うサービスです。顧客IDとプランIDを指定して生成します。
type SubscriptionService struct {
	service *Service
}

func newSubscriptionService(service *Service) *SubscriptionService {
	return &SubscriptionService{
		service: service,
	}
}

// Subscription はSubscribeやUpdateの引数を設定するのに使用する構造体です。
type Subscription struct {
	TrialEnd        interface{}       // トライアルの終了時期 (time.Time or "now")
	TrialEndAt      time.Time         // deprecated
	SkipTrial       interface{}       // deprecated
	PlanID          interface{}       // プランID(string)
	NextCyclePlanID interface{}       // 次サイクルから適用するプランID(string, 更新時のみ設定可能)
	Prorate         interface{}       // 日割り課金をするかどうか(bool)
	Metadata        map[string]string // メタデータ
}

type SubscriptionDelete struct {
	Prorate *bool `form:"prorate"`
}

// Subscribe は顧客IDとプランIDを指定して、定期課金を開始することができます。
// TrialEndを指定することで、プラン情報を上書きするトライアル設定も可能です。
// 最初の支払いは定期課金作成時に実行されます。
// 支払い実行日(BillingDay)が指定されているプランの場合は日割り設定(Prorate)を有効化しない限り、
// 作成時よりもあとの支払い実行日に最初の課金が行われます。またトライアル設定がある場合は、
// トライアル終了時に支払い処理が行われ、そこを基準にして定期課金が開始されます。
func (s SubscriptionService) Subscribe(customerID string, subscription Subscription) (*SubscriptionResponse, error) {
	trialEnd := subscription.getTrialEnd()
	qb := newRequestBuilder()
	qb.Add("customer", customerID)
	qb.Add("plan", subscription.PlanID)
	qb.Add("prorate", subscription.Prorate)
	qb.Add("trial_end", trialEnd)
	qb.AddMetadata(subscription.Metadata)

	body, err := s.service.request("POST", "/subscriptions", qb.Reader())
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Retrieve subscription object. 特定の定期課金情報を取得します。
func (s SubscriptionService) Retrieve(customerID, id string) (*SubscriptionResponse, error) {
	body, err := s.service.request("GET", "/customers/"+customerID+"/subscriptions/"+id, nil)
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Update はトライアル期間を新たに設定したり、プランの変更を行うことができます。
func (s SubscriptionService) Update(subscriptionID string, subscription Subscription) (*SubscriptionResponse, error) {
	trialEnd := subscription.getTrialEnd()
	qb := newRequestBuilder()
	qb.Add("next_cycle_plan", subscription.NextCyclePlanID)
	qb.Add("plan", subscription.PlanID)
	qb.Add("prorate", subscription.Prorate)
	qb.Add("trial_end", trialEnd)
	qb.AddMetadata(subscription.Metadata)
	body, err := s.service.request("POST", "/subscriptions/"+subscriptionID, qb.Reader())
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Pause は引き落としの失敗やカードが不正である、また定期課金を停止したい場合はこのリクエストで定期購入を停止させます。
//
// 定期課金を停止させると、再開されるまで引き落とし処理は一切行われません。
func (s SubscriptionService) Pause(subscriptionID string) (*SubscriptionResponse, error) {
	body, err := s.service.request("POST", "/subscriptions/"+subscriptionID+"/pause", nil)
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Resume は停止もしくはキャンセル状態の定期課金を再開させます。
func (s SubscriptionService) Resume(subscriptionID string, subscription Subscription) (*SubscriptionResponse, error) {
	trialEnd := subscription.getTrialEnd()
	qb := newRequestBuilder()
	qb.Add("trial_end", trialEnd)
	qb.Add("prorate", subscription.Prorate)

	body, err := s.service.request("POST", "/subscriptions/"+subscriptionID+"/resume", qb.Reader())
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Cancel は定期課金をキャンセルし、現在の周期の終了日をもって定期課金を終了させます。
func (s SubscriptionService) Cancel(subscriptionID string) (*SubscriptionResponse, error) {
	body, err := s.service.request("POST", "/subscriptions/"+subscriptionID+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) Delete(subscriptionID string, params ...SubscriptionDelete) error {
	path := "/subscriptions/" + subscriptionID
	if len(params) > 0 {
		path = path + s.service.getQuery(&params[0])
	}
	return s.service.delete(path)
}

// deprecated
func (s SubscriptionService) List() *subscriptionListCaller {
	return &subscriptionListCaller{
		service: s,
	}
}

func (subscription Subscription) getTrialEnd() interface{} {
	if subscription.TrialEnd != nil {
		return subscription.TrialEnd
	}
	skipTrial, ok := subscription.SkipTrial.(bool)
	if ok && skipTrial {
		return "now"
	}
	var isZero time.Time
	if subscription.TrialEndAt != isZero {
		return subscription.TrialEndAt
	}
	return nil
}

func parseSubscription(service *Service, body []byte, result *SubscriptionResponse) (*SubscriptionResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	return result, nil
}

// SubscriptionResponse はSubscriptionService.GetやSubscriptionService.Listで返される
// 定期購読情報持つ構造体です。
type SubscriptionResponse struct {
	CreatedAt            time.Time          // この定期課金作成時のタイムスタンプ
	Created              *int               `json:"created"`
	StartAt              time.Time          // この定期課金開始時のタイムスタンプ
	Start                *int               `json:"start"`
	CurrentPeriodStartAt time.Time          // 現在の購読期間開始時のタイムスタンプ
	CurrentPeriodStart   *int               `json:"current_period_start"`
	CurrentPeriodEndAt   time.Time          // 現在の購読期間終了時のタイムスタンプ
	CurrentPeriodEnd     *int               `json:"current_period_end"`
	TrialStartAt         time.Time          // トライアル期間開始時のタイムスタンプ
	TrialStart           *int               `json:"trial_start"`
	TrialEndAt           time.Time          // 	トライアル期間終了時のタイムスタンプ
	TrialEnd             *int               `json:"trial_end"`
	PausedAt             time.Time          // 定期課金が停止状態になった時のタイムスタンプ
	RawPausedAt          *int               `json:"paused_at"`
	CanceledAt           time.Time          // 定期課金がキャンセル状態になった時のタイムスタンプ
	RawCanceledAt        *int               `json:"canceled_at"`
	ResumedAt            time.Time          // 停止またはキャンセル状態の定期課金が有効状態になった時のタイムスタンプ
	RawResumedAt         *int               `json:"resumed_at"`
	Customer             string             `json:"customer"`
	ID                   string             `json:"id"`
	LiveMode             bool               `json:"livemode"`
	Object               string             `json:"object"`
	RawPlan              json.RawMessage    `json:"plan"`
	Plan                 Plan               // この定期課金のプラン情報
	NextCyclePlan        *Plan              `json:"next_cycle_plan"`
	Prorate              bool               `json:"prorate"`
	Status               SubscriptionStatus `json:"status"`
	Metadata             map[string]string  `json:"metadata"`

	service *Service
}

func (s *SubscriptionResponse) updateResponse(r *SubscriptionResponse, err error) error {
	if err != nil {
		return err
	}
	*s = *r
	return nil
}

// Update をSubscriptionResponseから実行します。
func (s *SubscriptionResponse) Update(subscription Subscription) error {
	return s.updateResponse(s.service.Subscription.Update(s.ID, subscription))
}

// Pause をSubscriptionResponseから実行します。
func (s *SubscriptionResponse) Pause() error {
	return s.updateResponse(s.service.Subscription.Pause(s.ID))
}

// Resume をSubscriptionResponseから実行します。
func (s *SubscriptionResponse) Resume(subscription Subscription) error {
	return s.updateResponse(s.service.Subscription.Resume(s.ID, subscription))
}

// Cancel をSubscriptionResponseから実行します。
func (s *SubscriptionResponse) Cancel() error {
	return s.updateResponse(s.service.Subscription.Cancel(s.ID))
}

// Delete をSubscriptionResponseから実行します。
func (s *SubscriptionResponse) Delete(params ...SubscriptionDelete) error {
	if len(params) > 0 {
		return s.service.Subscription.Delete(s.ID, params[0])
	}
	return s.service.Subscription.Delete(s.ID)
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (s *SubscriptionResponse) UnmarshalJSON(b []byte) error {
	type subscriptionResponseParser SubscriptionResponse
	var raw subscriptionResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "subscription" {
		raw.CanceledAt = time.Unix(IntValue(raw.RawCanceledAt), 0)
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.CurrentPeriodEndAt = time.Unix(IntValue(raw.CurrentPeriodEnd), 0)
		raw.CurrentPeriodStartAt = time.Unix(IntValue(raw.CurrentPeriodStart), 0)
		raw.PausedAt = time.Unix(IntValue(raw.RawPausedAt), 0)
		json.Unmarshal(raw.RawPlan, &raw.Plan)
		raw.ResumedAt = time.Unix(IntValue(raw.RawResumedAt), 0)
		raw.StartAt = time.Unix(IntValue(raw.Start), 0)
		raw.TrialEndAt = time.Unix(IntValue(raw.TrialEnd), 0)
		raw.TrialStartAt = time.Unix(IntValue(raw.TrialStart), 0)

		raw.service = s.service
		*s = SubscriptionResponse(raw)
		return nil
	}
	return parseError(b)
}

// SubscriptionListCaller はリスト取得に使用する構造体です。
type subscriptionListCaller struct {
	service SubscriptionService
	SubscriptionListParams
}

type SubscriptionListParams struct {
	ListParams `form:"*"`
	Customer   *string             `form:"customer"`
	Plan       *string             `form:"plan"`
	Status     *SubscriptionStatus `form:"status"`
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *subscriptionListCaller) Limit(limit int) *subscriptionListCaller {
	c.SubscriptionListParams.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *subscriptionListCaller) Offset(offset int) *subscriptionListCaller {
	c.SubscriptionListParams.ListParams.Offset = &offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *subscriptionListCaller) Since(since time.Time) *subscriptionListCaller {
	p := int(since.Unix())
	c.SubscriptionListParams.ListParams.Since = &p
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *subscriptionListCaller) Until(until time.Time) *subscriptionListCaller {
	p := int(until.Unix())
	c.SubscriptionListParams.ListParams.Until = &p
	return c
}

// Do は指定されたクエリーを元に顧客のリストを配列で取得します。
func (c *subscriptionListCaller) Do() ([]*SubscriptionResponse, bool, error) {
	return c.service.All(&c.SubscriptionListParams)
}

func (c SubscriptionService) All(params ...*SubscriptionListParams) ([]*SubscriptionResponse, bool, error) {
	p := &SubscriptionListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/subscriptions"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*SubscriptionResponse, len(raw.Data))
	for i, rawSubscription := range raw.Data {
		json.Unmarshal(rawSubscription, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}
