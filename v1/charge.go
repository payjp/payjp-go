package payjp

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	Product     interface{}       // プロダクトID (指定された場合、amountとcurrencyは無視されます)
	CustomerID  string            // 顧客ID (CardかCustomerのどちらかは必須パラメータ)
	CardToken   string            // トークンID (CardかCustomerのどちらかは必須パラメータ)
	CustomerCardID	string        // 顧客のカードID
	Capture     bool              // 支払い処理を確定するかどうか (falseの場合、カードの認証と支払い額の確保のみ行う)
	Description string            // 概要
	ExpireDays  interface{}       // デフォルトで7日となっており、1日~60日の間で設定が可能
	ThreeDSecure interface{}      // 3DSecureを実施するか否か (bool)
	Metadata    map[string]string // メタデータ
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
		} else  {
			qb.Add("currency", charge.Currency)
		}
	}
	if charge.CustomerID != "" {
	    if charge.CardToken != "" {
	    	return nil, fmt.Errorf("The following parameters are exclusive: CustomerID, CardToken.")
	    }
		qb.Add("customer", charge.CustomerID)
		if charge.CustomerCardID != "" {
			qb.Add("card", charge.CustomerCardID)
		}
	} else if charge.CardToken != "" {
		qb.Add("card", charge.CardToken)
	} else {
		return nil, fmt.Errorf("One of the following parameters is required: CustomerID or CardToken")
	}

	qb.Add("capture", charge.Capture)
	expireDays, ok := charge.ExpireDays.(int)
	if ok {
		qb.Add("expiry_days", expireDays)
	}
	if charge.Description != "" {
		qb.Add("description", charge.Description)
	}
	ThreeDSecure, ok := charge.ThreeDSecure.(bool)
	if ok {
		qb.Add("three_d_secure", ThreeDSecure)
	}
	qb.AddMetadata(charge.Metadata)

	body, err := c.service.request("POST", "/charges", qb.Reader())
	if err != nil {
		return nil, err
	}
	return parseCharge(c.service, body, &ChargeResponse{})

}

// Retrieve charge object. 支払い情報を取得します。
func (c ChargeService) Retrieve(chargeID string) (*ChargeResponse, error) {
	body, err := c.service.retrieve("/charges/" + chargeID)
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

// Do は指定されたクエリーを元に支払いのリストを配列で取得します。
func (c *ChargeListCaller) Do() ([]*ChargeResponse, bool, error) {
	body, err := c.service.queryList("/charges", c.limit, c.offset, c.since, c.until, func(values *url.Values) bool {
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
	FeeRate        string            // 決済手数料率
	ThreeDSecureStatus *string       // 3Dセキュアの実施状況

	service *Service
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
	FeeRate        string            `json:"fee_rate"`
	ThreeDSecureStatus *string       `json:"three_d_secure_status"`
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
		c.FeeRate = raw.FeeRate
		c.ThreeDSecureStatus = raw.ThreeDSecureStatus
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}
	return nil
}
