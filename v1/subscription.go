package payjp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SubscriptionStatus は定期購読のステータスを表すEnumです。
type SubscriptionStatus int

const (
	noSubscriptionStatus SubscriptionStatus = iota
	// SubscriptionActive はアクティブ状態を表す定数
	SubscriptionActive
	// SubscriptionTrial はトライアル状態を表す定数
	SubscriptionTrial
	// SubscriptionCanceled はキャンセル状態を表す定数
	SubscriptionCanceled
	// SubscriptionPaused は停止状態を表す定数
	SubscriptionPaused
)

func (s SubscriptionStatus) status() interface{} {
	switch s {
	case SubscriptionActive:
		return "active"
	case SubscriptionTrial:
		return "trial"
	case SubscriptionCanceled:
		return "canceled"
	case SubscriptionPaused:
		return "paused"
	}
	return nil
}

// SubscriptionService は月単位で定期的な支払い処理を行うサービスです。顧客IDとプランIDを指定して生成します。
//
// stauts=SubscriptionTrial の場合は支払いは行われず、status=SubscriptionActive の場合のみ支払いが行われます。
//
// 支払い処理は、はじめに定期課金を生成した瞬間に行われ、そこを基準にして定期的な支払いが行われていきます。
// 定期課金は、顧客に複数紐付けるができ、生成した定期課金は停止・再開・キャンセル・削除することができます。
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
	TrialEndAt time.Time         // トライアルの終了時期
	SkipTrial  interface{}       // トライアルをしない(bool)
	PlanID     interface{}       // プランID(string)
	NextCyclePlanID interface{}  // 次サイクルから適用するプランID(string, 更新時のみ設定可能)
	Prorate    interface{}       // 日割り課金をするかどうか(bool)
	Metadata   map[string]string // メタデータ
}

// Subscribe は顧客IDとプランIDを指定して、定期課金を開始することができます。
// TrialEndを指定することで、プラン情報を上書きするトライアル設定も可能です。
// 最初の支払いは定期課金作成時に実行されます。
// 支払い実行日(BillingDay)が指定されているプランの場合は日割り設定(Prorate)を有効化しない限り、
// 作成時よりもあとの支払い実行日に最初の課金が行われます。またトライアル設定がある場合は、
// トライアル終了時に支払い処理が行われ、そこを基準にして定期課金が開始されます。
func (s SubscriptionService) Subscribe(customerID string, subscription Subscription) (*SubscriptionResponse, error) {
	return s.SubscribeContext(context.Background(), customerID, subscription)
}

func (s SubscriptionService) SubscribeContext(ctx context.Context, customerID string, subscription Subscription) (*SubscriptionResponse, error) {
	var errors []string
	planID, ok := subscription.PlanID.(string)
	if !ok || planID == "" {
		errors = append(errors, "PlanID is required, but empty.")
	}
	var defaultTime time.Time
	skipTrial, ok := subscription.SkipTrial.(bool)
	if subscription.TrialEndAt != defaultTime && ok {
		errors = append(errors, "TrialEndAt and SkipTrial are exclusive.")
	}
	if len(errors) != 0 {
		return nil, fmt.Errorf("Subscription.Subscribe() parameter error: %s", strings.Join(errors, ", "))
	}
	qb := newRequestBuilder()
	qb.Add("customer", customerID)
	qb.Add("plan", subscription.PlanID)
	if subscription.TrialEndAt != defaultTime {
		qb.Add("trial_end", strconv.Itoa(int(subscription.TrialEndAt.Unix())))
	} else if ok && skipTrial {
		qb.Add("trial_end", "now")
	}
	qb.Add("prorate", subscription.Prorate)
	qb.AddMetadata(subscription.Metadata)
	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", s.service.apiKey)

	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Retrieve subscription object. 特定の定期課金情報を取得します。
func (s SubscriptionService) Retrieve(customerID, subscriptionID string) (*SubscriptionResponse, error) {
	return s.RetrieveContext(context.Background(), customerID, subscriptionID)
}

func (s SubscriptionService) RetrieveContext(ctx context.Context, customerID, subscriptionID string) (*SubscriptionResponse, error) {
	body, err := s.service.retrieve(ctx, "/customers/" + customerID + "/subscriptions/" + subscriptionID)
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

func (s SubscriptionService) update(ctx context.Context, subscriptionID string, subscription Subscription) ([]byte, error) {
	var defaultTime time.Time
	_, ok := subscription.SkipTrial.(bool)
	if subscription.TrialEndAt != defaultTime && ok {
		return nil, errors.New("Subscription.Update() parameter error: TrialEndAt and SkipTrial are exclusive")
	}
	qb := newRequestBuilder()
	qb.Add("next_cycle_plan", subscription.NextCyclePlanID)
	qb.Add("plan", subscription.PlanID)
	if subscription.TrialEndAt != defaultTime {
		qb.Add("trial_end", strconv.Itoa(int(subscription.TrialEndAt.Unix())))
	} else if subscription.SkipTrial == true {
		qb.Add("trial_end", "now")
	}
	qb.Add("prorate", subscription.Prorate)
	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions/"+subscriptionID, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", s.service.apiKey)
	return parseResponseError(s.service.Client.Do(request))
}

// Update はトライアル期間を新たに設定したり、プランの変更を行うことができます。
//
// トライアル期間を更新する場合、トライアル期間終了時に支払い処理が行われ、
// そこを基準としてプランに沿った周期で定期課金が再開されます。
// このトライアル期間を利用すれば、定期課金の開始日を任意の日にずらすこともできます。
// また SkipTrial=true とする事により、トライアル期間中の定期課金を即時開始できます。
//
// プランを変更する場合は、 PlanID に新しいプランのIDを指定してください。
// 同時に Prorate=true とする事により、 日割り課金を有効化できます。
func (s SubscriptionService) Update(subscriptionID string, subscription Subscription) (*SubscriptionResponse, error) {
	return s.UpdateContext(context.Background(), subscriptionID, subscription)
}

func (s SubscriptionService) UpdateContext(ctx context.Context, subscriptionID string, subscription Subscription) (*SubscriptionResponse, error) {
	body, err := s.update(ctx, subscriptionID, subscription)
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Pause は引き落としの失敗やカードが不正である、また定期課金を停止したい場合はこのリクエストで定期購入を停止させます。
//
// 定期課金を停止させると、再開されるまで引き落とし処理は一切行われません。
func (s SubscriptionService) Pause(subscriptionID string) (*SubscriptionResponse, error) {
	return s.PauseContext(context.Background(), subscriptionID)
}

func (s SubscriptionService) PauseContext(ctx context.Context, subscriptionID string) (*SubscriptionResponse, error) {
	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions/"+subscriptionID+"/pause", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Resume は停止もしくはキャンセル状態の定期課金を再開させます。
// トライアル日数が残っていて再開日がトライアル終了日時より前の場合、
// トライアル状態で定期課金が再開されます。
//
// TrialEndを指定することで、トライアル終了日を任意の日時に再指定する事ができます。
//
// 支払いの失敗が原因で停止状態にある定期課金の再開時は未払い分の支払いを行います。
//
// 未払い分の支払いに失敗すると、定期課金は再開されません。 この場合は、有効なカードを顧客のデフォルトカードにセットしてから、
// 再度定期課金の再開を行ってください。
//
// またProrate を指定することで、日割り課金を有効化することができます。 日割り課金が有効な場合は、
// 再開日より課金日までの日数分で課金額を日割りします。
func (s SubscriptionService) Resume(subscriptionID string, subscription Subscription) (*SubscriptionResponse, error) {
	return s.ResumeContext(context.Background(), subscriptionID, subscription)
}

func (s SubscriptionService) ResumeContext(ctx context.Context, subscriptionID string, subscription Subscription) (*SubscriptionResponse, error) {
	var defaultTime time.Time
	qb := newRequestBuilder()
	if subscription.TrialEndAt != defaultTime {
		qb.Add("trial_end", strconv.Itoa(int(subscription.TrialEndAt.Unix())))
	}
	qb.Add("prorate", subscription.Prorate)

	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions/"+subscriptionID+"/resume", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Cancel は定期課金をキャンセルし、現在の周期の終了日をもって定期課金を終了させます。
//
// 終了日以前であれば、定期課金の再開リクエスト(Resume)を行うことで、
// キャンセルを取り消すことができます。終了日をむかえた定期課金は、
// 自動的に削除されますのでご注意ください。
func (s SubscriptionService) Cancel(subscriptionID string) (*SubscriptionResponse, error) {
	return s.CancelContext(context.Background(), subscriptionID)
}

func (s SubscriptionService) CancelContext(ctx context.Context, subscriptionID string) (*SubscriptionResponse, error) {
	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions/"+subscriptionID+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseSubscription(s.service, body, &SubscriptionResponse{})
}

// Delete は定期課金をすぐに削除します。次回以降の課金は行われずに、一度削除した定期課金は、
// 再び戻すことができません。
func (s SubscriptionService) Delete(subscriptionID string) error {
	return s.DeleteContext(context.Background(), subscriptionID)
}

func (s SubscriptionService) DeleteContext(ctx context.Context, subscriptionID string) error {
	request, err := http.NewRequestWithContext(ctx, "DELETE", s.service.apiBase+"/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	_, err = parseResponseError(s.service.Client.Do(request))
	return err
}

// List は顧客の定期課金リストを取得します。リストは、直近で生成された順番に取得されます。
func (s SubscriptionService) List() *SubscriptionListCaller {
	return &SubscriptionListCaller{
		service: s.service,
	}
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
	ID                   string             // sub_で始まる一意なオブジェクトを示す文字列
	LiveMode             bool               // 本番環境かどうか
	CreatedAt            time.Time          // この定期課金作成時のタイムスタンプ
	StartAt              time.Time          // この定期課金開始時のタイムスタンプ
	CustomerID           string             // この定期課金を購読している顧客のID
	Plan                 Plan               // この定期課金のプラン情報
	NextCyclePlan        *Plan              // この定期課金の次のサイクルから適用されるプラン情報
	Status               SubscriptionStatus // この定期課金の現在の状態
	Prorate              bool               // 日割り課金が有効かどうか
	CurrentPeriodStartAt time.Time          // 現在の購読期間開始時のタイムスタンプ
	CurrentPeriodEndAt   time.Time          // 現在の購読期間終了時のタイムスタンプ
	TrialStartAt         time.Time          // トライアル期間開始時のタイムスタンプ
	TrialEndAt           time.Time          // 	トライアル期間終了時のタイムスタンプ
	PausedAt             time.Time          // 定期課金が停止状態になった時のタイムスタンプ
	CanceledAt           time.Time          // 定期課金がキャンセル状態になった時のタイムスタンプ
	ResumedAt            time.Time          // 停止またはキャンセル状態の定期課金が有効状態になった時のタイムスタンプ
	Metadata             map[string]string  // メタデータ

	service *Service
}

type subscriptionResponseParser struct {
	CanceledEpoch           int               `json:"canceled_at"`
	CreatedEpoch            int               `json:"created"`
	CurrentPeriodEndEpoch   int               `json:"current_period_end"`
	CurrentPeriodStartEpoch int               `json:"current_period_start"`
	Customer                string            `json:"customer"`
	ID                      string            `json:"id"`
	LiveMode                bool              `json:"livemode"`
	Object                  string            `json:"object"`
	PausedEpoch             int               `json:"paused_at"`
	Plan                    json.RawMessage   `json:"plan"`
	NextCyclePlan           json.RawMessage   `json:"next_cycle_plan"`
	Prorate                 bool              `json:"prorate"`
	ResumedEpoch            int               `json:"resumed_at"`
	StartEpoch              int               `json:"start"`
	Status                  string            `json:"status"`
	TrialEndEpoch           int               `json:"trial_end"`
	TrialStartEpoch         int               `json:"trial_start"`
	Metadata                map[string]string `json:"metadata"`
}

// Update はトライアル期間を新たに設定したり、プランの変更を行うことができます。
func (s *SubscriptionResponse) Update(subscription Subscription) error {
	return s.UpdateContext(context.Background(), subscription)
}

func (s *SubscriptionResponse) UpdateContext(ctx context.Context, subscription Subscription) error {
	body, err := s.service.Subscription.update(ctx, s.ID, subscription)
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

// Pause は引き落としの失敗やカードが不正である、また定期課金を停止したい場合はこのリクエストで定期購入を停止させます。
func (s *SubscriptionResponse) Pause() error {
	return s.PauseContext(context.Background())
}

func (s *SubscriptionResponse) PauseContext(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions/"+s.ID+"/pause", nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

// Resume は停止もしくはキャンセル状態の定期課金を再開させます。
func (s *SubscriptionResponse) Resume(subscription Subscription) error {
	return s.ResumeContext(context.Background(), subscription)
}

func (s *SubscriptionResponse) ResumeContext(ctx context.Context, subscription Subscription) error {
	var defaultTime time.Time
	qb := newRequestBuilder()
	if subscription.TrialEndAt != defaultTime {
		qb.Add("trial_end", strconv.Itoa(int(subscription.TrialEndAt.Unix())))
	}
	qb.Add("prorate", subscription.Prorate)

	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions/"+s.ID+"/resume", qb.Reader())
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

// Cancel は定期課金をキャンセルし、現在の周期の終了日をもって定期課金を終了させます。
func (s *SubscriptionResponse) Cancel() error {
	return s.CancelContext(context.Background())
}

func (s *SubscriptionResponse) CancelContext(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "POST", s.service.apiBase+"/subscriptions/"+s.ID+"/cancel", nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	body, err := respToBody(s.service.Client.Do(request))
	if err != nil {
		return err
	}
	_, err = parseSubscription(s.service, body, s)
	return err
}

// Delete は定期課金をすぐに削除します。次回以降の課金は行われずに、一度削除した定期課金は、
// 再び戻すことができません。
func (s *SubscriptionResponse) Delete() error {
	return s.DeleteContext(context.Background())
}

func (s *SubscriptionResponse) DeleteContext(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, "DELETE", s.service.apiBase+"/subscriptions/"+s.ID, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.service.apiKey)
	_, err = parseResponseError(s.service.Client.Do(request))
	return err
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (s *SubscriptionResponse) UnmarshalJSON(b []byte) error {
	raw := subscriptionResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "subscription" {
		s.CanceledAt = time.Unix(int64(raw.CanceledEpoch), 0)
		s.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		s.CurrentPeriodEndAt = time.Unix(int64(raw.CurrentPeriodEndEpoch), 0)
		s.CurrentPeriodStartAt = time.Unix(int64(raw.CurrentPeriodStartEpoch), 0)
		s.CustomerID = raw.Customer
		s.ID = raw.ID
		s.LiveMode = raw.LiveMode
		s.PausedAt = time.Unix(int64(raw.PausedEpoch), 0)
		json.Unmarshal(raw.Plan, &s.Plan)
		json.Unmarshal(raw.NextCyclePlan, &s.NextCyclePlan)
		s.Prorate = raw.Prorate
		s.ResumedAt = time.Unix(int64(raw.ResumedEpoch), 0)
		s.StartAt = time.Unix(int64(raw.StartEpoch), 0)
		switch raw.Status {
		case "active":
			s.Status = SubscriptionActive
		case "trial":
			s.Status = SubscriptionTrial
		case "canceled":
			s.Status = SubscriptionCanceled
		case "paused":
			s.Status = SubscriptionPaused
		}
		s.TrialEndAt = time.Unix(int64(raw.TrialEndEpoch), 0)
		s.TrialStartAt = time.Unix(int64(raw.TrialStartEpoch), 0)
		s.Metadata = raw.Metadata
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

// SubscriptionListCaller はリスト取得に使用する構造体です。
type SubscriptionListCaller struct {
	service    *Service
	customerID string
	limit      int
	offset     int
	since      int
	until      int
	planID     string
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *SubscriptionListCaller) Limit(limit int) *SubscriptionListCaller {
	c.limit = limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *SubscriptionListCaller) Offset(offset int) *SubscriptionListCaller {
	c.offset = offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *SubscriptionListCaller) Since(since time.Time) *SubscriptionListCaller {
	c.since = int(since.Unix())
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *SubscriptionListCaller) Until(until time.Time) *SubscriptionListCaller {
	c.until = int(until.Unix())
	return c
}

// PlanID はプランIDで結果を絞ります
func (c *SubscriptionListCaller) PlanID(planID string) *SubscriptionListCaller {
	c.planID = planID
	return c
}

// Do は指定されたクエリーを元に顧客のリストを配列で取得します。
func (c *SubscriptionListCaller) Do() ([]*SubscriptionResponse, bool, error) {
	return c.DoContext(context.Background())
}

func (c *SubscriptionListCaller) DoContext(ctx context.Context) ([]*SubscriptionResponse, bool, error) {
	var url string
	if c.customerID == "" {
		url = "/subscriptions"
	} else {
		url = "/customers/" + c.customerID + "/subscriptions"
	}
	body, err := c.service.queryList(ctx, url, c.limit, c.offset, c.since, c.until)
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
		subscription := &SubscriptionResponse{}
		json.Unmarshal(rawSubscription, subscription)
		result[i] = subscription
	}
	return result, raw.HasMore, nil
}
