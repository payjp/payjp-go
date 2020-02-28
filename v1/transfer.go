package payjp

import (
	"encoding/json"
	"net/url"
	"time"
)

// TransferStatus は入金状態を示すステータスです
type TransferStatus int

const (
	noTransferStatus TransferStatus = iota
	// TransferPending は支払い前のステータスを表す定数
	TransferPending
	// TransferPaid は支払い済みのステータスを表す定数
	TransferPaid
	// TransferFailed は支払い失敗のステータスを表す定数
	TransferFailed
	// TransferCanceled は支払いキャンセルのステータスを表す定数
	TransferCanceled
	// TransferRecombination は組戻ステータスを表す定数
	TransferRecombination
)

func (t TransferStatus) status() interface{} {
	switch t {
	case TransferPending:
		return "pending"
	case TransferPaid:
		return "paid"
	case TransferFailed:
		return "failed"
	case TransferCanceled:
		return "canceled"
	case TransferRecombination:
		return "recombination"
	}
	return nil
}

// TransferService は入金に関するサービスです。
//
// 入金は毎月15日と月末に締め、翌月15日と月末に入金されます。入金は、締め日までのデータがそれぞれ生成されます。
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
	body, err := t.service.retrieve("/transfers/" + transferID)
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

// List は入金リストを取得します。リストは、直近で生成された順番に取得されます。
func (t TransferService) List() *TransferListCaller {
	return &TransferListCaller{
		status:  noTransferStatus,
		service: t.service,
	}
}

// TransferListCaller は支払いのリスト取得に使用する構造体です。
type TransferListCaller struct {
	service *Service
	limit   int
	offset  int
	since   int
	until   int
	status  TransferStatus
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *TransferListCaller) Limit(limit int) *TransferListCaller {
	c.limit = limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *TransferListCaller) Offset(offset int) *TransferListCaller {
	c.offset = offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *TransferListCaller) Since(since time.Time) *TransferListCaller {
	c.since = int(since.Unix())
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *TransferListCaller) Until(until time.Time) *TransferListCaller {
	c.until = int(until.Unix())
	return c
}

// Status はここで指定されたステータスのデータを取得します
func (c *TransferListCaller) Status(status TransferStatus) *TransferListCaller {
	c.status = status
	return c
}

// Do は指定されたクエリーを元に入金のリストを配列で取得します。
func (c *TransferListCaller) Do() ([]*TransferResponse, bool, error) {
	body, err := c.service.queryList("/transfers", c.limit, c.offset, c.since, c.until, func(values *url.Values) bool {
		if c.status != noTransferStatus {
			values.Add("status", c.status.status().(string))
			return true
		}
		return false
	})
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*TransferResponse, len(raw.Data))
	for i, rawCharge := range raw.Data {
		charge := &TransferResponse{}
		json.Unmarshal(rawCharge, charge)
		charge.service = c.service
		result[i] = charge
	}
	return result, raw.HasMore, nil
}

// ChargeList は支払いは入金内訳リストを取得します。リストは、直近で生成された順番に取得されます。
func (t TransferService) ChargeList(transferID string) *TransferChargeListCaller {
	return &TransferChargeListCaller{
		service:    t.service,
		transferID: transferID,
	}
}

// TransferChargeListCaller は入金内訳のリスト取得に使用する構造体です。
type TransferChargeListCaller struct {
	service    *Service
	transferID string
	limit      int
	offset     int
	since      int
	until      int
	customerID string
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *TransferChargeListCaller) Limit(limit int) *TransferChargeListCaller {
	c.limit = limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *TransferChargeListCaller) Offset(offset int) *TransferChargeListCaller {
	c.offset = offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *TransferChargeListCaller) Since(since time.Time) *TransferChargeListCaller {
	c.since = int(since.Unix())
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *TransferChargeListCaller) Until(until time.Time) *TransferChargeListCaller {
	c.until = int(until.Unix())
	return c
}

// CustomerID はここに指定した顧客IDを持つデータを取得します
func (c *TransferChargeListCaller) CustomerID(ID string) *TransferChargeListCaller {
	c.customerID = ID
	return c
}

// Do は指定されたクエリーを元に入金内訳のリストを配列で取得します。
func (c *TransferChargeListCaller) Do() ([]*ChargeResponse, bool, error) {
	path := "/transfers/" + c.transferID + "/charges"
	body, err := c.service.queryList(path, c.limit, c.offset, c.since, c.until, func(values *url.Values) bool {
		if c.customerID != "" {
			values.Add("customer", c.customerID)
			return true
		}
		return false
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
		transfer := &ChargeResponse{}
		json.Unmarshal(rawCharge, transfer)
		transfer.service = c.service
		result[i] = transfer
	}
	return result, raw.HasMore, nil
}

// TransferResponse はTransferService.Get、TransferService.Listによって返される、
// 入金状態を示す構造体です。
type TransferResponse struct {
	ID             string            // tr_で始まる一意なオブジェクトを示す文字列
	LiveMode       bool              // 本番環境かどうか
	CreatedAt      time.Time         // この入金作成時のタイムスタンプ
	Amount         int               // 入金予定額
	CarriedBalance int               // 繰越金
	Currency       string            // 3文字のISOコード(現状 “jpy” のみサポート)
	Status         TransferStatus    // この入金の処理状態
	Charges        []*ChargeResponse // この入金に含まれる支払いのリスト
	ScheduledDate  string            // 入金予定日
	Summary        struct {
		ChargeCount  int // 支払い総数
		ChargeFee    int // 支払い手数料
		ChargeGross  int // 総売上
		Net          int // 差引額
		RefundAmount int // 返金総額
		RefundCount  int // 返金総数
	} // この入金に関する集計情報
	Description    string    // 概要
	TermStartAt    time.Time // 集計期間開始時のタイムスタンプ
	TermEndAt      time.Time // 集計期間終了時のタイムスタンプ
	TransferAmount int       // 	入金額
	TransferDate   string    // 入金日

	service *Service
}

type transferResponseParser struct {
	Amount         int                `json:"amount"`
	CarriedBalance int                `json:"carried_balance"`
	Charges        listResponseParser `json:"charges"`
	CreatedEpoch   int                `json:"created"`
	Currency       string             `json:"currency"`
	Description    string             `json:"description"`
	ID             string             `json:"id"`
	LiveMode       bool               `json:"livemode"`
	Object         string             `json:"object"`
	ScheduledDate  string             `json:"scheduled_date"`
	Status         string             `json:"status"`
	Summary        struct {
		ChargeCount  int `json:"charge_count"`
		ChargeFee    int `json:"charge_fee"`
		ChargeGross  int `json:"charge_gross"`
		Net          int `json:"net"`
		RefundAmount int `json:"refund_amount"`
		RefundCount  int `json:"refund_count"`
	} `json:"summary"`
	TermEndEpoch   int    `json:"term_end"`
	TermStartEpoch int    `json:"term_start"`
	TransferAmount int    `json:"transfer_amount"`
	TransferDate   string `json:"transfer_date"`
}

type transfer TransferResponse

// UnmarshalJSON はJSONパース用の内部APIです。
func (t *TransferResponse) UnmarshalJSON(b []byte) error {
	raw := transferResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "transfer" {
		t.Amount = raw.Amount
		t.CarriedBalance = raw.CarriedBalance
		t.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		t.Currency = raw.Currency
		t.Description = raw.Description
		t.ID = raw.ID
		t.LiveMode = raw.LiveMode
		t.ScheduledDate = raw.ScheduledDate
		switch raw.Status {
		case "pending":
			t.Status = TransferPending
		case "paid":
			t.Status = TransferPaid
		case "failed":
			t.Status = TransferFailed
		case "canceled":
			t.Status = TransferCanceled
		case "recombination":
			t.Status = TransferRecombination
		}
		t.Summary.ChargeCount = raw.Summary.ChargeCount
		t.Summary.ChargeFee = raw.Summary.ChargeFee
		t.Summary.ChargeGross = raw.Summary.ChargeGross
		t.Summary.Net = raw.Summary.Net
		t.Summary.RefundAmount = raw.Summary.RefundAmount
		t.Summary.RefundCount = raw.Summary.RefundCount
		t.TermEndAt = time.Unix(int64(raw.TermEndEpoch), 0)
		t.TermStartAt = time.Unix(int64(raw.TermStartEpoch), 0)
		t.TransferAmount = raw.TransferAmount
		t.TransferDate = raw.TransferDate
		t.Charges = make([]*ChargeResponse, len(raw.Charges.Data))
		for i, rawCharge := range raw.Charges.Data {
			charge := &ChargeResponse{}
			json.Unmarshal(rawCharge, charge)
			t.Charges[i] = charge
		}

		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
