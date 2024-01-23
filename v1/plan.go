package payjp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// PlanService は定期購入のときに使用する静的なプラン情報を扱います。
//
// 金額、支払い実行日(1-31)、トライアル日数などを指定して、 あなたのビジネスに必要なさまざまなプランを生成することができます。
//
// 生成したプランは、顧客と紐付けて定期購入処理を行うことができます。
type PlanService struct {
	service *Service
}

func newPlanService(service *Service) *PlanService {
	return &PlanService{
		service: service,
	}
}

// Plan はプランの作成時に使用する構造体です。
type Plan struct {
	Amount     int               // 必須: 金額。50~9,999,999の整数
	Currency   string            // 3文字のISOコード(現状 “jpy” のみサポート)
	Interval   string            // 月次など
	ID         string            // プランID
	Name       string            // プランの名前
	TrialDays  int               // トライアル日数
	BillingDay int               // 支払いの実行日(1〜31)
	Metadata   map[string]string // メタデータ
}

// Create は金額や通貨などを指定して定期購入に利用するプランを生成します。
//
// トライアル日数を指定することで、トライアル付きのプランを生成することができます。
//
// また、支払いの実行日を指定すると、支払い日の固定されたプランを生成することができます。
func (p PlanService) Create(plan Plan) (*PlanResponse, error) {
	if plan.BillingDay < 0 || plan.BillingDay > 31 {
		return nil, fmt.Errorf("billing_day should be between 1 and 31, but %d", plan.BillingDay)
	}
	qb := newRequestBuilder()
	qb.Add("amount", strconv.Itoa(plan.Amount))
	if plan.Currency == "" {
		qb.Add("currency", "jpy")
	} else {
		qb.Add("currency", plan.Currency)
	}
	if plan.Interval == "" {
		qb.Add("interval", "month")
	} else {
		qb.Add("interval", plan.Interval)
	}
	if plan.ID != "" {
		qb.Add("id", plan.ID)
	}
	if plan.Name != "" {
		qb.Add("name", plan.Name)
	}
	if plan.TrialDays != 0 {
		qb.Add("trial_days", strconv.Itoa(plan.TrialDays))
	}
	if plan.BillingDay != 0 {
		qb.Add("billing_day", strconv.Itoa(plan.BillingDay))
	}
	qb.AddMetadata(plan.Metadata)

	body, err := p.service.request("POST", "/plans", qb.Reader())
	if err != nil {
		return nil, err
	}
	return parsePlan(p.service, body, &PlanResponse{})
}

// Retrieve plan object. 特定のプラン情報を取得します。
func (p PlanService) Retrieve(id string) (*PlanResponse, error) {
	body, err := p.service.request("GET", "/plans/"+id, nil)
	if err != nil {
		return nil, err
	}
	return parsePlan(p.service, body, &PlanResponse{})
}

func parsePlan(service *Service, body []byte, result *PlanResponse) (*PlanResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	return result, nil
}

// Update はプラン情報を更新します。
func (p PlanService) Update(id string, plan Plan) (*PlanResponse, error) {
	qb := newRequestBuilder()
	qb.Add("name", plan.Name)
	qb.AddMetadata(plan.Metadata)
	body, err := p.service.request("POST", "/plans/"+id, qb.Reader())
	if err != nil {
		return nil, err
	}
	return parsePlan(p.service, body, &PlanResponse{})
}

// Delete はプランを削除します。
func (p PlanService) Delete(id string) error {
	return p.service.delete("/plans/" + id)
}

// List は生成したプランのリストを取得します。リストは、直近で生成された順番に取得されます。
func (p PlanService) List() *planListCaller {
	return &planListCaller{
		service: p,
	}
}

// PlanListCaller はプランのリスト取得に使用する構造体です。
type PlanListParams struct {
	ListParams `form:"*"`
}

type planListCaller struct {
	service PlanService
	PlanListParams
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *planListCaller) Limit(limit int) *planListCaller {
	c.PlanListParams.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *planListCaller) Offset(offset int) *planListCaller {
	c.PlanListParams.ListParams.Offset = &offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *planListCaller) Since(since time.Time) *planListCaller {
	p := int(since.Unix())
	c.PlanListParams.ListParams.Since = &p
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *planListCaller) Until(until time.Time) *planListCaller {
	p := int(until.Unix())
	c.PlanListParams.ListParams.Until = &p
	return c
}

// Deprecated: Use Do instead
func (c *planListCaller) Do() ([]*PlanResponse, bool, error) {
	return c.service.All(&c.PlanListParams)
}
func (c PlanService) All(params ...*PlanListParams) ([]*PlanResponse, bool, error) {
	p := &PlanListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/plans"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*PlanResponse, len(raw.Data))
	for i, rawPlan := range raw.Data {
		json.Unmarshal(rawPlan, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}

// PlanResponse はPlanService.はPlanService.Listで返されるプランを表す構造体です
type PlanResponse struct {
	CreatedAt  time.Time         // このプラン作成時のタイムスタンプ
	Amount     int               `json:"amount"`
	BillingDay int               `json:"billing_day"`
	Created    *int              `json:"created"`
	Currency   string            `json:"currency"`
	ID         string            `json:"id"`
	Interval   string            `json:"interval"`
	LiveMode   bool              `json:"livemode"`
	Name       string            `json:"name"`
	Object     string            `json:"object"`
	TrialDays  int               `json:"trial_days"`
	Metadata   map[string]string `json:"metadata"`

	service *Service
}

func (p *PlanResponse) updateResponse(r *PlanResponse, err error) error {
	if err != nil {
		return err
	}
	*p = *r
	return nil
}

// Update はプラン情報を更新します。
func (p *PlanResponse) Update(plan Plan) error {
	return p.updateResponse(p.service.Plan.Update(p.ID, plan))
}

// Delete はプランを削除します。
func (p *PlanResponse) Delete() error {
	return p.service.Plan.Delete(p.ID)
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (p *PlanResponse) UnmarshalJSON(b []byte) error {
	type planResponseParser PlanResponse
	var raw planResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "plan" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.service = p.service
		*p = PlanResponse(raw)
		return nil
	}
	return parseError(b)
}
