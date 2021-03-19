package payjp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	Currency    string            // 必須: 3文字のISOコード(現状 “jpy” のみサポート)
	CustomerID  string            // 顧客ID (CardかCustomerのどちらかは必須パラメータ)
	Card        Card              // カードオブジェクト(cardかcustomerのどちらかは必須)
	CardToken   string            // トークンID (CardかCustomerのどちらかは必須パラメータ)
	CustomerCardID	string        // 顧客のカードID
	Capture     bool              // 支払い処理を確定するかどうか (falseの場合、カードの認証と支払い額の確保のみ行う)
	Description string            // 	概要
	ExpireDays  interface{}       // デフォルトで7日となっており、1日~60日の間で設定が可能
	Metadata    map[string]string // メタデータ
}

// Create Deprecated: use CreateContext instead
func (c ChargeService) Create(amount int, charge Charge) (*ChargeResponse, error) {
	return c.CreateContext(context.Background(), amount, charge)
}

// CreateContext はトークンID、カードを保有している顧客ID、カードオブジェクトのいずれかのパラメーターを指定して支払いを作成します。
// 顧客IDを使って支払いを作成する場合は CustomerCardID に顧客の保有するカードのIDを指定でき、省略された場合はデフォルトカードとして登録されているものが利用されます。
// テスト用のキーでは、本番用の決済ネットワークへは接続されず、実際の請求が行われることもありません。 本番用のキーでは、決済ネットワークで処理が行われ、実際の請求が行われます。
//
// 支払いを確定せずに、カードの認証と支払い額のみ確保する場合は、 Capture に false を指定してください。 このとき ExpireDays を指定することで、認証の期間を定めることができます。 ExpireDays はデフォルトで7日となっており、1日~60日の間で設定が可能です。
func (c ChargeService) CreateContext(ctx context.Context, amount int, charge Charge) (*ChargeResponse, error) {
	var errorMessages []string
	if amount < 50 || amount > 9999999 {
		errorMessages = append(errorMessages, fmt.Sprintf("Amount should be between 50 and 9,999,999, but %d.", amount))
	}
	counter := 0
	if charge.CustomerID != "" {
		counter++
	}
	if charge.CardToken != "" {
		counter++
	}
	if charge.Card.valid() {
		counter++
	}
	switch counter {
	case 0:
		errorMessages = append(errorMessages, "One of the following parameters is required: CustomerID, CardToken, Card")
	case 1:
	case 2, 3:
		errorMessages = append(errorMessages, "The following parameters are exclusive: CustomerID, CardToken, Card")
	}
	if charge.Currency == "" {
		charge.Currency = "jpy"
	} else if charge.Currency != "jpy" {
		// todo: if pay.jp supports other currency, fix this condition
		errorMessages = append(errorMessages, fmt.Sprintf("Only supports 'jpy' as currency, but '%s'.", charge.Currency))
	}
	expireDays, ok := charge.ExpireDays.(int)
	if ok && (expireDays < -1 || expireDays > 60) {
		errorMessages = append(errorMessages, fmt.Sprintf("ExpireDays should be between 1 and 60, but %d.", expireDays))
	}
	if len(errorMessages) > 0 {
		return nil, fmt.Errorf("Charge.Create() parameter error: %s", strings.Join(errorMessages, ", "))
	}
	qb := newRequestBuilder()
	qb.Add("amount", amount)
	qb.Add("currency", charge.Currency)
	if charge.CustomerID != "" {
		qb.Add("customer", charge.CustomerID)
		if charge.CustomerCardID != "" {
			qb.Add("card", charge.CustomerCardID)
		}
	} else if charge.CardToken != "" {
		qb.Add("card", charge.CardToken)
	}
	qb.AddCard(charge.Card)
	qb.Add("description", charge.Description)
	qb.Add("capture", charge.Capture)
	qb.Add("expiry_days", charge.ExpireDays)
	qb.AddMetadata(charge.Metadata)

	request, err := http.NewRequestWithContext(ctx, "POST", c.service.apiBase+"/charges", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	body, err := respToBody(c.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})

}

// Retrieve Deprecated: use RetrieveContext instead
func (c ChargeService) Retrieve(chargeID string) (*ChargeResponse, error) {
	return c.RetrieveContext(context.Background(), chargeID)
}

// RetrieveContext charge object. 支払い情報を取得します。
func (c ChargeService) RetrieveContext(ctx context.Context, chargeID string) (*ChargeResponse, error) {
	body, err := c.service.retrieve(ctx, "/charges/" + chargeID)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) update(ctx context.Context, chargeID, description string, metadata map[string]string) ([]byte, error) {
	qb := newRequestBuilder()
	qb.Add("name", description)
	qb.AddMetadata(metadata)
	request, err := http.NewRequestWithContext(ctx, "POST", c.service.apiBase+"/charges/"+chargeID, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

// Update Deprecated: use UpdateContext instead
func (c ChargeService) Update(chargeID, description string, metadata ...map[string]string) (*ChargeResponse, error) {
	return c.UpdateContext(context.Background(), chargeID, description, metadata...)
}

// UpdateContext は支払い情報のDescriptionを更新します。
func (c ChargeService) UpdateContext(ctx context.Context, chargeID, description string, metadata ...map[string]string) (*ChargeResponse, error) {
	var md map[string]string
	switch len(metadata) {
	case 0:
	case 1:
		md = metadata[0]
	default:
		return nil, fmt.Errorf("Update can accept zero or one metadata map, but %d are passed", len(metadata))
	}
	body, err := c.update(ctx, chargeID, description, md)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

func (c ChargeService) refund(ctx context.Context, id string, reason string, amount []int) ([]byte, error) {
	qb := newRequestBuilder()
	if len(amount) > 0 {
		qb.Add("amount", amount[0])
	}
	qb.Add("refund_reason", reason)
	request, err := http.NewRequestWithContext(ctx, "POST", c.service.apiBase+"/charges/"+id+"/refund", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

// Refund Deprecated: use RefundContext instead
func (c ChargeService) Refund(chargeID, reason string, amount ...int) (*ChargeResponse, error) {
	return c.RefundContext(context.Background(), chargeID, reason, amount...)
}

// RefundContext は支払い済みとなった処理を返金します。
// Amount省略時は全額返金、指定時に金額の部分返金を行うことができます。
func (c ChargeService) RefundContext(ctx context.Context, chargeID, reason string, amount ...int) (*ChargeResponse, error) {
	body, err := c.refund(ctx, chargeID, reason, amount)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}



func (c ChargeService) capture(ctx context.Context, chargeID string, amount []int) ([]byte, error) {
	qb := newRequestBuilder()
	if len(amount) > 0 {
		qb.Add("amount", amount[0])
	}
	request, err := http.NewRequestWithContext(ctx, "POST", c.service.apiBase+"/charges/"+chargeID+"/capture", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

// Capture Deprecated: use CaptureContext instead
func (c ChargeService) Capture(chargeID string, amount ...int) (*ChargeResponse, error) {
	return c.CaptureContext(context.Background(), chargeID, amount...)
}

// CaptureContext は認証状態となった処理待ちの支払い処理を確定させます。具体的には Captured="false" となった支払いが該当します。
//
// amount をセットすることで、支払い生成時の金額と異なる金額の支払い処理を行うことができます。 ただし amount は、支払い生成時の金額よりも少額である必要があるためご注意ください。
//
// amount をセットした場合、AmountRefunded に認証時の amount との差額が入ります。
//
// 例えば、認証時に amount=500 で作成し、 amount=400 で支払い確定を行った場合、 AmountRefunded=100 となり、確定金額が400円に変更された状態で支払いが確定されます。
func (c ChargeService) CaptureContext(ctx context.Context, chargeID string, amount ...int) (*ChargeResponse, error) {
	body, err := c.capture(ctx, chargeID, amount)
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})
}

// List は生成した支払い情報のリストを取得します。リストは、直近で生成された順番に取得されます。
func (c ChargeService) List() *ChargeListCaller {
	return &ChargeListCaller{
		service: c.service,
	}
}

// ChargeListCaller はリスト取得に使用する構造体です。
//
// Fluentインタフェースを提供しており、最後にDoを呼ぶことでリストが取得できます:
//
//     pay := payjp.New("api-key", nil)
//     charges, err := pay.Charge.List().Limit(50).Offset(150).Do()
type ChargeListCaller struct {
	service        *Service
	limit          int
	offset         int
	since          int
	until          int
	customerID     string
	subscriptionID string
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *ChargeListCaller) Limit(limit int) *ChargeListCaller {
	c.limit = limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *ChargeListCaller) Offset(offset int) *ChargeListCaller {
	c.offset = offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *ChargeListCaller) Since(since time.Time) *ChargeListCaller {
	c.since = int(since.Unix())
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *ChargeListCaller) Until(until time.Time) *ChargeListCaller {
	c.until = int(until.Unix())
	return c
}

// CustomerID を指定すると、指定した顧客の支払いのみを取得します
func (c *ChargeListCaller) CustomerID(id string) *ChargeListCaller {
	c.customerID = id
	return c
}

// SubscriptionID を指定すると、指定した定期購読の支払いのみを取得します
func (c *ChargeListCaller) SubscriptionID(id string) *ChargeListCaller {
	c.subscriptionID = id
	return c
}

// Do Deprecated: use DoContext instead
func (c *ChargeListCaller) Do() ([]*ChargeResponse, bool, error) {
	return c.DoContext(context.Background())
}

// DoContext は指定されたクエリーを元に支払いのリストを配列で取得します。
func (c *ChargeListCaller) DoContext(ctx context.Context) ([]*ChargeResponse, bool, error) {
	body, err := c.service.queryList(ctx, "/charges", c.limit, c.offset, c.since, c.until, func(values *url.Values) bool {
		result := false
		if c.customerID != "" {
			values.Add("customer", c.customerID)
			result = true
		}
		if c.subscriptionID != "" {
			values.Add("subscription", c.subscriptionID)
			result = true
		}
		return result
	})
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*ChargeResponse, len(raw.Data))
	for i, rawCharge := range raw.Data {
		charge := &ChargeResponse{}
		json.Unmarshal(rawCharge, charge)
		charge.service = c.service
		result[i] = charge
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

// ChargeResponse はCharge.Getなどで返される、支払いに関する情報を持った構造体です
type ChargeResponse struct {
	ID             string            // ch_で始まる一意なオブジェクトを示す文字列
	LiveMode       bool              // 本番環境かどうか
	CreatedAt      time.Time         // この支払い作成時のタイムスタンプ
	Amount         int               // 支払額
	Currency       string            // 3文字のISOコード(現状 “jpy” のみサポート)
	Paid           bool              // 認証処理が成功しているかどうか。
	ExpiredAt      time.Time         // 認証状態が自動的に失効される日時のタイムスタンプ
	Captured       bool              // 支払い処理を確定しているかどうか
	CapturedAt     time.Time         // 支払い処理確定時のタイムスタンプ
	Card           CardResponse      // 支払いされたクレジットカードの情報
	CustomerID     string            // 顧客ID
	Description    string            // 概要
	FailureCode    string            // 失敗した支払いのエラーコード
	FailureMessage string            // 失敗した支払いの説明
	Refunded       bool              // 返金済みかどうか
	AmountRefunded int               // この支払いに対しての返金額
	RefundReason   string            // 返金理由
	SubscriptionID string            // sub_から始まる定期課金のID
	Metadata       map[string]string // メタデータ

	service *Service
}

// Update Deprecated: use UpdateContext instead
func (c *ChargeResponse) Update(description string, metadata ...map[string]string) error {
	return c.UpdateContext(context.Background(), description, metadata...)
}

// UpdateContext は支払い情報のDescriptionとメタデータ(オプション)を更新します
func (c *ChargeResponse) UpdateContext(ctx context.Context, description string, metadata ...map[string]string) error {
	var md map[string]string
	switch len(metadata) {
	case 0:
	case 1:
		md = metadata[0]
	default:
		return fmt.Errorf("Update can accept zero or one metadata map, but %d are passed", len(metadata))
	}
	body, err := c.service.Charge.update(ctx, c.ID, description, md)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

// Refund Deprecated: use RefundContext instead
func (c *ChargeResponse) Refund(reason string, amount ...int) error {
	return c.RefundContext(context.Background(), reason, amount...)
}

// RefundContext 支払い済みとなった処理を返金します。
// 全額返金、及び amount を指定することで金額の部分返金を行うことができます。ただし部分返金を最初に行った場合、2度目の返金は全額返金しか行うことができないため、ご注意ください。
func (c *ChargeResponse) RefundContext(ctx context.Context, reason string, amount ...int) error {
	var body []byte
	var err error
	body, err = c.service.Charge.refund(ctx, c.ID, reason, amount)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

// Capture Deprecated: use CaptureContext instead
func (c *ChargeResponse) Capture(amount ...int) error {
	return c.CaptureContext(context.Background(), amount...)
}

// CaptureContext は認証状態となった処理待ちの支払い処理を確定させます。具体的には Captured="false" となった支払いが該当します。
//
// amount をセットすることで、支払い生成時の金額と異なる金額の支払い処理を行うことができます。 ただし amount は、支払い生成時の金額よりも少額である必要があるためご注意ください。
//
// amount をセットした場合、AmountRefunded に認証時の amount との差額が入ります。
//
// 例えば、認証時に amount=500 で作成し、 amount=400 で支払い確定を行った場合、 AmountRefunded=100 となり、確定金額が400円に変更された状態で支払いが確定されます。
func (c *ChargeResponse) CaptureContext(ctx context.Context, amount ...int) error {
	body, err := c.service.Charge.capture(ctx, c.ID, amount)
	if err != nil {
		return err
	}
	_, err = parseCharge(c.service, body, c)
	return err
}

type chargeResponseParser struct {
	Amount         int               `json:"amount"`
	AmountRefunded int               `json:"amount_refunded"`
	Captured       bool              `json:"captured"`
	CapturedEpoch  int               `json:"captured_at"`
	Card           json.RawMessage   `json:"card"`
	CreatedEpoch   int               `json:"created"`
	Currency       string            `json:"currency"`
	Customer       string            `json:"customer"`
	Description    string            `json:"description"`
	ExpiredEpoch   int               `json:"expired_at"`
	FailureCode    string            `json:"failure_code"`
	FailureMessage string            `json:"failure_message"`
	ID             string            `json:"id"`
	LiveMode       bool              `json:"livemode"`
	Object         string            `json:"object"`
	Paid           bool              `json:"paid"`
	RefundReason   string            `json:"refund_reason"`
	Refunded       bool              `json:"refunded"`
	Subscription   string            `json:"subscription"`
	Metadata       map[string]string `json:"metadata"`
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (c *ChargeResponse) UnmarshalJSON(b []byte) error {
	raw := chargeResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "charge" {
		c.Amount = raw.Amount
		c.AmountRefunded = raw.AmountRefunded
		c.Captured = raw.Captured
		c.CapturedAt = time.Unix(int64(raw.CapturedEpoch), 0)
		json.Unmarshal(raw.Card, &c.Card)
		c.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		c.Currency = raw.Currency
		c.CustomerID = raw.Customer
		c.Description = raw.Description
		c.ExpiredAt = time.Unix(int64(raw.ExpiredEpoch), 0)
		c.FailureCode = raw.FailureCode
		c.FailureMessage = raw.FailureMessage
		c.ID = raw.ID
		c.LiveMode = raw.LiveMode
		c.Paid = raw.Paid
		c.RefundReason = raw.RefundReason
		c.Refunded = raw.Refunded
		c.SubscriptionID = raw.Subscription
		c.Metadata = raw.Metadata
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}
	return nil
}
