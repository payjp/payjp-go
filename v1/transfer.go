package payjp

import (
	"encoding/json"
	"time"
)

// TransferStatus は入金状態を示すステータスです
type TransferStatus string

const (
	// TransferPending は支払い前のステータスを表す定数
	TransferPending = TransferStatus("pending")
	// TransferPaid は支払い済みのステータスを表す定数
	TransferPaid = TransferStatus("paid")
	// TransferFailed は支払い失敗のステータスを表す定数
	TransferFailed = TransferStatus("failed")
	// TransferRecombination は組戻ステータスを表す定数
	TransferRecombination = TransferStatus("recombination")
	// TransferCarriedOver は入金繰り越しを表す定数
	TransferCarriedOver = TransferStatus("carried_over")
	// TransferStop は入金停止を表す定数
	TransferStop = TransferStatus("stop")
)

func (t TransferStatus) status() interface{} {
	return string(t)
}

type TransferListParams struct {
	ListParams        `form:"*"`
	SinceSheduledDate *int            `form:"since_scheduled_date"`
	UntilSheduledDate *int            `form:"until_scheduled_date"`
	Status            *TransferStatus `form:"status"`
}

// TransferChargeListParams は入金内訳のリスト取得に使用する構造体です。
type TransferChargeListParams struct {
	ListParams `form:"*"`
	Customer   *string `form:"customer"`
}

type transferListCaller struct {
	service TransferService
	TransferListParams
}

type TransferChargeListCaller struct {
	caller *TransferResponse
	TransferChargeListParams
}

// TransferService は入金に関するサービスです。
type TransferService struct {
	service *Service
}

func newTransferService(service *Service) *TransferService {
	return &TransferService{
		service: service,
	}
}

// Retrieve transfer object. 入金情報を取得します。
func (t TransferService) Retrieve(transferID string) (*TransferResponse, error) {
	body, err := t.service.request("GET", "/transfers/"+transferID, nil)
	if err != nil {
		return nil, err
	}
	result := &TransferResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = t.service
	return result, nil
}

func (c TransferService) All(params ...*TransferListParams) ([]*TransferResponse, bool, error) {
	p := &TransferListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/transfers"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*TransferResponse, len(raw.Data))
	for i, raw := range raw.Data {
		json.Unmarshal(raw, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}

// Deprecated
func (t TransferService) List() *transferListCaller {
	return &transferListCaller{
		service: t,
	}
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *transferListCaller) Limit(limit int) *transferListCaller {
	c.TransferListParams.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *transferListCaller) Offset(offset int) *transferListCaller {
	c.TransferListParams.ListParams.Offset = &offset
	return c
}

// SinceSheduledDate は入金予定日がここに指定したタイムスタンプ以降のデータのみ取得します
func (c *transferListCaller) SinceSheduledDate(sinceSheduledDate time.Time) *transferListCaller {
	p := int(sinceSheduledDate.Unix())
	c.TransferListParams.SinceSheduledDate = &p
	return c
}

// UntilSheduledDate は入金予定日がここに指定したタイムスタンプ以前のデータのみ取得します
func (c *transferListCaller) UntilSheduledDate(untilSheduledDate time.Time) *transferListCaller {
	p := int(untilSheduledDate.Unix())
	c.TransferListParams.UntilSheduledDate = &p
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *transferListCaller) Since(since time.Time) *transferListCaller {
	p := int(since.Unix())
	c.TransferListParams.ListParams.Since = &p
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *transferListCaller) Until(until time.Time) *transferListCaller {
	p := int(until.Unix())
	c.TransferListParams.ListParams.Until = &p
	return c
}

// Status はここで指定されたステータスのデータを取得します
func (c *transferListCaller) Status(status TransferStatus) *transferListCaller {
	c.TransferListParams.Status = &status
	return c
}

// Do は指定されたクエリーを元に入金のリストを配列で取得します。
func (c *transferListCaller) Do() ([]*TransferResponse, bool, error) {
	return c.service.All(&c.TransferListParams)
}

// ChargeList は支払いは入金内訳リストを取得します。リストは、直近で生成された順番に取得されます。
func (t TransferService) ChargeList(id string) *TransferChargeListCaller {
	tr, err := t.Retrieve(id)
	if err != nil {
		return nil
	}
	return &TransferChargeListCaller{
		caller: tr,
	}
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *TransferChargeListCaller) Limit(limit int) *TransferChargeListCaller {
	c.TransferChargeListParams.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *TransferChargeListCaller) Offset(offset int) *TransferChargeListCaller {
	c.TransferChargeListParams.ListParams.Offset = &offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *TransferChargeListCaller) Since(since time.Time) *TransferChargeListCaller {
	p := int(since.Unix())
	c.TransferChargeListParams.ListParams.Since = &p
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *TransferChargeListCaller) Until(until time.Time) *TransferChargeListCaller {
	p := int(until.Unix())
	c.TransferChargeListParams.ListParams.Until = &p
	return c
}

// CustomerID はここに指定した顧客IDを持つデータを取得します
func (c *TransferChargeListCaller) CustomerID(ID string) *TransferChargeListCaller {
	c.TransferChargeListParams.Customer = &ID
	return c
}

// Do は指定されたクエリーを元に入金内訳のリストを配列で取得します。
func (c *TransferChargeListCaller) Do() ([]*ChargeResponse, bool, error) {
	return c.caller.All(&c.TransferChargeListParams)
}

func (c *TransferResponse) All(params ...*TransferChargeListParams) ([]*ChargeResponse, bool, error) {
	p := &TransferChargeListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	path := "/transfers/" + c.ID + "/charges" + c.service.getQuery(p)
	body, err := c.service.request("GET", path, nil)
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

// TransferResponse はTransferService.Get、TransferService.Listによって返される、
// 入金状態を示す構造体です。
type TransferResponse struct {
	ID                string             `json:"id"`       // tr_で始まる一意なオブジェクトを示す文字列
	LiveMode          bool               `json:"livemode"` // 本番環境かどうか
	Created           *int               `json:"created"`  // この入金作成時のタイムスタンプ
	CreatedAt         time.Time          // この入金作成時のタイムスタンプ
	Amount            int                `json:"amount"` // 入金予定額
	CarriedBalance    int                // 繰越金
	RawCarriedBalance *int               `json:"carried_balance"`
	Currency          string             `json:"currency"` // 3文字のISOコード(現状 “jpy” のみサポート)
	Status            TransferStatus     `json:"status"`   // この入金の処理状態
	Charges           []*ChargeResponse  // この入金に含まれる支払いのリスト
	RawCharges        listResponseParser `json:"charges"`
	ScheduledDate     string             `json:"scheduled_date"` // 入金予定日
	Summary           struct {
		ChargeCount   int `json:"charge_count"`   // 支払い総数
		ChargeFee     int `json:"charge_fee"`     // 支払い手数料
		ChargeGross   int `json:"charge_gross"`   // 総売上
		Net           int `json:"net"`            // 差引額
		RefundAmount  int `json:"refund_amount"`  // 返金総額
		RefundCount   int `json:"refund_count"`   // 返金総数
		DisputeAmount int `json:"dispute_amount"` // チャージバックにより相殺された金額の合計
		DisputeCount  int `json:"dispute_count"`  // チャージバック対象となったchargeの個数
	} `json:"summary"`
	Description       string    // 概要
	RawDescription    *string   `json:"description"`
	TermStartAt       time.Time // 集計期間開始時のタイムスタンプ
	TermEndAt         time.Time // 集計期間終了時のタイムスタンプ
	TermEnd           *int      `json:"term_end"`
	TermStart         *int      `json:"term_start"`
	TransferAmount    int
	TransferDate      string
	RawTransferAmount *int    `json:"transfer_amount"`
	RawTransferDate   *string `json:"transfer_date"`
	Object            string  `json:"object"`

	service *Service
}

func (t *TransferResponse) UnmarshalJSON(b []byte) error {
	type transferResponseParser TransferResponse
	var raw transferResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "transfer" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.TermEndAt = time.Unix(IntValue(raw.TermEnd), 0)
		raw.TermStartAt = time.Unix(IntValue(raw.TermStart), 0)
		raw.CarriedBalance = int(IntValue(raw.RawCarriedBalance))
		raw.Description = StringValue(raw.RawDescription)
		raw.TransferAmount = int(IntValue(raw.RawTransferAmount))
		raw.TransferDate = StringValue(raw.RawTransferDate)
		raw.Charges = make([]*ChargeResponse, len(raw.RawCharges.Data))
		for i, rawCharge := range raw.RawCharges.Data {
			json.Unmarshal(rawCharge, &(raw.Charges[i]))
		}
		raw.service = t.service
		*t = TransferResponse(raw)

		return nil
	}
	return parseError(b)
}
