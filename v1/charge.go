package payjp

import (
	"encoding/json"
	"fmt"
	"time"
)

// ChargeService 都度の支払いや定期購入の引き落としのときに生成される、支払い情報を取り扱います。
type ChargeService struct {
	service *Service
}

func newChargeService(service *Service) *ChargeService {
	return &ChargeService{
		service: service,
	}
}

// Charge 構造体はCharge.Createのパラメータを設定するのに使用します
type Charge struct {
	Currency       string            // 必須: 3文字のISOコード(現状 “jpy” のみサポート)
	Product        interface{}       // プロダクトID (指定された場合、amountとcurrencyは無視されます)
	CustomerID     string            // 顧客ID (CardかCustomerのどちらかは必須パラメータ)
	CardToken      string            // トークンID (CardかCustomerのどちらかは必須パラメータ)
	CustomerCardID string            // 顧客のカードID
	Capture        bool              // 支払い処理を確定するかどうか (falseの場合、カードの認証と支払い額の確保のみ行う)
	Description    string            // 概要
	ExpireDays     interface{}       // デフォルトで7日となっており、1日~60日の間で設定が可能
	ThreeDSecure   interface{}       // 3DSecureを実施するか否か (bool)
	Metadata       map[string]string // メタデータ
}

// Create はトークンID、カードを保有している顧客IDのいずれかのパラメーターを指定して支払いを作成します。
// 顧客IDを使って支払いを作成する場合は CustomerCardID に顧客の保有するカードのIDを指定でき、省略された場合はデフォルトカードとして登録されているものが利用されます。
// テスト用のキーでは、本番用の決済ネットワークへは接続されず、実際の請求が行われることもありません。 本番用のキーでは、決済ネットワークで処理が行われ、実際の請求が行われます。
//
// 支払いを確定せずに、カードの認証と支払い額のみ確保する場合は、 Capture に false を指定してください。 このとき ExpireDays を指定することで、認証の期間を定めることができます。 ExpireDays はデフォルトで7日となっており、1日~60日の間で設定が可能です。
func (c ChargeService) Create(amount int, charge Charge) (*ChargeResponse, error) {
	qb := newRequestBuilder()
	product, ok := charge.Product.(string)
	if ok {
		qb.Add("product", product)
	} else {
		qb.Add("amount", amount)
		if charge.Currency == "" {
			qb.Add("currency", "jpy")
		} else {
			qb.Add("currency", charge.Currency)
		}
	}
	if charge.CustomerID != "" {
		qb.Add("customer", charge.CustomerID)
	}
	if charge.CustomerCardID != "" {
		qb.Add("card", charge.CustomerCardID)
	}
	if charge.CardToken != "" {
		qb.Add("card", charge.CardToken)
	}

	qb.Add("capture", charge.Capture)
	qb.Add("expiry_days", charge.ExpireDays)
	if charge.Description != "" {
		qb.Add("description", charge.Description)
	}
	qb.Add("three_d_secure", charge.ThreeDSecure)
	qb.AddMetadata(charge.Metadata)

	body, err := c.service.request("POST", "/charges", qb.Reader())
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})

}

// Retrieve charge object. 支払い情報を取得します。
func (c ChargeService) Retrieve(chargeID string) (*ChargeResponse, error) {
	body, err := c.service.request("GET", "/charges/"+chargeID, nil)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) update(chargeID, description string, metadata map[string]string) ([]byte, error) {
	qb := newRequestBuilder()
	qb.Add("description", description)
	qb.AddMetadata(metadata)
	return c.service.request("POST", "/charges/"+chargeID, qb.Reader())
}

// Update は支払い情報のDescriptionを更新します。
func (c ChargeService) Update(chargeID, description string, metadata ...map[string]string) (*ChargeResponse, error) {
	var md map[string]string
	switch len(metadata) {
	case 0:
	case 1:
		md = metadata[0]
	default:
		return nil, fmt.Errorf("Update can accept zero or one metadata map, but %d are passed", len(metadata))
	}
	body, err := c.update(chargeID, description, md)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) refund(id string, reason string, amount []int) ([]byte, error) {
	qb := newRequestBuilder()
	if len(amount) > 0 {
		qb.Add("amount", amount[0])
	}
	qb.Add("refund_reason", reason)

	return c.service.request("POST", "/charges/"+id+"/refund", qb.Reader())
}

// Refund は支払い済みとなった処理を返金します。
// Amount省略時は全額返金、指定時に金額の部分返金を行うことができます。
func (c ChargeService) Refund(chargeID, reason string, amount ...int) (*ChargeResponse, error) {
	body, err := c.refund(chargeID, reason, amount)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) capture(chargeID string, amount []int) ([]byte, error) {
	qb := newRequestBuilder()
	if len(amount) > 0 {
		qb.Add("amount", amount[0])
	}
	return c.service.request("POST", "/charges/"+chargeID+"/capture", qb.Reader())
}

// Capture は認証状態となった処理待ちの支払い処理を確定させます。具体的には Captured="false" となった支払いが該当します。
//
// amount をセットすることで、支払い生成時の金額と異なる金額の支払い処理を行うことができます。 ただし amount は、支払い生成時の金額よりも少額である必要があるためご注意ください。
//
// amount をセットした場合、AmountRefunded に認証時の amount との差額が入ります。
//
// 例えば、認証時に amount=500 で作成し、 amount=400 で支払い確定を行った場合、 AmountRefunded=100 となり、確定金額が400円に変更された状態で支払いが確定されます。
func (c ChargeService) Capture(chargeID string, amount ...int) (*ChargeResponse, error) {
	body, err := c.capture(chargeID, amount)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

// TdsFinish は3Dセキュア認証が終了した支払いに対し、決済を行います。
// https://pay.jp/docs/api/#3d%E3%82%BB%E3%82%AD%E3%83%A5%E3%82%A2%E3%83%95%E3%83%AD%E3%83%BC%E3%82%92%E5%AE%8C%E4%BA%86%E3%81%99%E3%82%8B
func (c ChargeService) TdsFinish(id string) (*ChargeResponse, error) {
	body, err := c.service.request("POST", "/charges/"+id+"/tds_finish", nil)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

// Deprecated
func (c ChargeService) List() *ChargeListCaller {
	p := &ChargeListParams{}
	return &ChargeListCaller{
		service:          c,
		ChargeListParams: *p,
	}
}

type ChargeListParams struct {
	ListParams   `form:"*"`
	Customer     *string `form:"customer"`
	Subscription *string `form:"subscription"`
	Tenant       *string `form:"tenant"`
}

type ChargeListCaller struct {
	service ChargeService
	ChargeListParams
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *ChargeListCaller) Limit(limit int) *ChargeListCaller {
	c.ChargeListParams.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *ChargeListCaller) Offset(offset int) *ChargeListCaller {
	c.ChargeListParams.ListParams.Offset = &offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *ChargeListCaller) Since(since time.Time) *ChargeListCaller {
	p := int(since.Unix())
	c.ChargeListParams.ListParams.Since = &p
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *ChargeListCaller) Until(until time.Time) *ChargeListCaller {
	p := int(until.Unix())
	c.ChargeListParams.ListParams.Until = &p
	return c
}

// CustomerID を指定すると、指定した顧客の支払いのみを取得します
func (c *ChargeListCaller) CustomerID(id string) *ChargeListCaller {
	c.ChargeListParams.Customer = &id
	return c
}

// SubscriptionID を指定すると、指定した定期購読の支払いのみを取得します
func (c *ChargeListCaller) SubscriptionID(id string) *ChargeListCaller {
	c.ChargeListParams.Subscription = &id
	return c
}

func (c *ChargeListCaller) Do() ([]*ChargeResponse, bool, error) {
	return c.service.All(&c.ChargeListParams)
}

func (c ChargeService) All(params ...*ChargeListParams) ([]*ChargeResponse, bool, error) {
	p := &ChargeListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/charges"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*ChargeResponse, len(raw.Data))
	for i, r := range raw.Data {
		json.Unmarshal(r, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}

func parseCharge(service *Service, data []byte, result *ChargeResponse) (*ChargeResponse, error) {
	err := json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	return result, nil
}

func (c *ChargeResponse) updateResponse(r *ChargeResponse, err error) error {
	if err != nil {
		return err
	}
	*c = *r
	return nil
}

// Update は支払い情報のDescriptionとメタデータ(オプション)を更新します
func (c *ChargeResponse) Update(description string, metadata ...map[string]string) error {
	var md map[string]string
	switch len(metadata) {
	case 0:
	case 1:
		md = metadata[0]
	default:
		return fmt.Errorf("Update can accept zero or one metadata map, but %d are passed", len(metadata))
	}
	body, err := c.service.Charge.update(c.ID, description, md)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

// Refund 支払い済みとなった処理を返金します。
// 全額返金、及び amount を指定することで金額の部分返金を行うことができます。ただし部分返金を最初に行った場合、2度目の返金は全額返金しか行うことができないため、ご注意ください。
func (c *ChargeResponse) Refund(reason string, amount ...int) error {
	var body []byte
	var err error
	body, err = c.service.Charge.refund(c.ID, reason, amount)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

// Capture は認証状態となった処理待ちの支払い処理を確定させます。具体的には Captured="false" となった支払いが該当します。
//
// amount をセットすることで、支払い生成時の金額と異なる金額の支払い処理を行うことができます。 ただし amount は、支払い生成時の金額よりも少額である必要があるためご注意ください。
//
// amount をセットした場合、AmountRefunded に認証時の amount との差額が入ります。
//
// 例えば、認証時に amount=500 で作成し、 amount=400 で支払い確定を行った場合、 AmountRefunded=100 となり、確定金額が400円に変更された状態で支払いが確定されます。
func (c *ChargeResponse) Capture(amount ...int) error {
	body, err := c.service.Charge.capture(c.ID, amount)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

// TdsFinish をChrgeResponseから実行します。
func (c *ChargeResponse) TdsFinish() error {
	return c.updateResponse(c.service.Charge.TdsFinish(c.ID))
}

// ChargeResponse はCharge.Getなどで返される、支払いに関する情報を持った構造体です
type ChargeResponse struct {
	ID                 string          `json:"id"`       // ch_で始まる一意なオブジェクトを示す文字列
	LiveMode           bool            `json:"livemode"` // 本番環境かどうか
	Created            *int            `json:"created"`  // この支払い作成時のタイムスタンプ
	CreatedAt          time.Time       // この支払い作成時のタイムスタンプ
	Amount             int             `json:"amount"`     // 支払額
	Currency           string          `json:"currency"`   // 3文字のISOコード(現状 “jpy” のみサポート)
	Paid               bool            `json:"paid"`       // 認証処理が成功しているかどうか。
	RawExpiredAt       *int            `json:"expired_at"` // 認証状態が自動的に失効される日時のタイムスタンプ
	ExpiredAt          time.Time       // 認証状態が自動的に失効される日時のタイムスタンプ
	Captured           bool            `json:"captured"`    // 支払い処理を確定しているかどうか
	RawCapturedAt      *int            `json:"captured_at"` // 支払い処理確定時のタイムスタンプ
	CapturedAt         time.Time       // 支払い処理確定時のタイムスタンプ
	RawCard            json.RawMessage `json:"card"` // 支払いされたクレジットカードの情報
	Card               CardResponse    // 支払いされたクレジットカードの情報
	Customer           *string         `json:"customer"` // 顧客ID
	CustomerID         string          // 顧客ID
	RawDescription     *string         `json:"description"` // 概要
	Description        string
	RawFailureCode     *string `json:"failure_code"`    // 失敗した支払いのエラーコード
	RawFailureMessage  *string `json:"failure_message"` // 失敗した支払いの説明
	FailureCode        string  // 失敗した支払いのエラーコード
	FailureMessage     string  // 失敗した支払いの説明
	Refunded           bool    `json:"refunded"`        // 返金済みかどうか
	AmountRefunded     int     `json:"amount_refunded"` // この支払いに対しての返金額
	RawRefundReason    *string `json:"refund_reason"`   // 返金理由
	RefundReason       string  // 返金理由
	Subscription       *string `json:"subscription"` // sub_から始まる定期課金のID
	SubscriptionID     string
	Metadata           map[string]string `json:"metadata"`
	FeeRate            string            `json:"fee_rate"`              // 決済手数料率
	ThreeDSecureStatus *string           `json:"three_d_secure_status"` // 3Dセキュアの実施状況
	Object             string            `json:"object"`

	service *Service
}

type chargeResponseParser ChargeResponse

// UnmarshalJSON はJSONパース用の内部APIです。
func (c *ChargeResponse) UnmarshalJSON(b []byte) error {
	raw := chargeResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "charge" {
		raw.CapturedAt = time.Unix(IntValue(raw.RawCapturedAt), 0)
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.ExpiredAt = time.Unix(IntValue(raw.RawExpiredAt), 0)
		raw.CustomerID = StringValue(raw.Customer)
		raw.SubscriptionID = StringValue(raw.Subscription)
		raw.Description = StringValue(raw.RawDescription)
		raw.FailureCode = StringValue(raw.RawFailureCode)
		raw.FailureMessage = StringValue(raw.RawFailureMessage)
		json.Unmarshal(raw.RawCard, &raw.Card)
		raw.service = c.service
		*c = ChargeResponse(raw)
		return nil
	}
	return parseError(b)
}
