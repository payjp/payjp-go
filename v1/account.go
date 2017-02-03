package payjp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// AccountService はあなたのアカウント情報を返します。
type AccountService struct {
	service *Service
}

func newAccountService(service *Service) *AccountService {
	return &AccountService{service}
}

// Get はあなたのアカウント情報を取得します。
func (t *AccountService) Retrieve() (*AccountResponse, error) {
	request, err := http.NewRequest("GET", t.service.apiBase+"/accounts", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", t.service.apiKey)

	resp, err := t.service.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	result := &AccountResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AccountResponse はAccount.Retrieve()メソッドが返す構造体です
type AccountResponse struct {
	ID        string           // acct_で始まる一意なオブジェクトを示す文字列
	Email     string           // メールアドレス
	CreatedAt time.Time        // このアカウント作成時のタイムスタンプ
	Merchant  struct {
		ID                  string    // acct_mch_で始まる一意なオブジェクトを示す文字列
		BankEnabled         bool      // 入金先銀行口座情報が設定済みかどうか
		BrandsAccepted      []string  // 本番環境で利用可能なカードブランドのリスト
		CurrenciesSupported []string  // 対応通貨のリスト
		DefaultCurrency     string    // 3文字のISOコード(現状 “jpy” のみサポート)
		BusinessType        string    // 業務形態
		ContactPhone        string    // 電話番号
		Country             string    // 所在国
		ChargeType          []string  // 支払い方法種別のリスト
		ProductDetail       string    // 販売商品の詳細
		ProductName         string    // 販売商品名
		ProductType         []string  // 販売商品の種類リスト
		DetailsSubmitted    bool      // 本番環境申請情報が提出済みかどうか
		LiveModeEnabled     bool      // 本番環境が有効かどうか
		LiveModeActivatedAt time.Time // 本番環境が許可された日時のタイムスタンプ
		SitePublished       bool      // 申請対象のサイトがオープン済みかどうか

		URL       string    // 申請対象サイトのURL
		CreatedAt time.Time // 作成時のタイムスタンプ
	} // マーチャントアカウントの詳細情報
}

type accountResponseParser struct {
	CreatedEpoch int             `json:"created"`
	Email        string          `json:"email"`
	ID           string          `json:"id"`
	Merchant     struct {
		BankEnabled            bool     `json:"bank_enabled"`
		BrandsAccepted         []string `json:"brands_accepted"`
		BusinessType           string   `json:"business_type"`
		ChargeType             []string `json:"charge_type"`
		ContactPhone           string   `json:"contact_phone"`
		Country                string   `json:"country"`
		CreatedEpoch           int      `json:"created"`
		CurrenciesSupported    []string `json:"currencies_supported"`
		DefaultCurrency        string   `json:"default_currency"`
		DetailsSubmitted       bool     `json:"details_submitted"`
		ID                     string   `json:"id"`
		LiveModeActivatedEpoch int      `json:"livemode_activated_at"`
		LiveModeEnabled        bool     `json:"livemode_enabled"`
		Object                 string   `json:"object"`
		ProductDetail          string   `json:"product_detail"`
		ProductName            string   `json:"product_name"`
		ProductType            []string `json:"product_type"`
		SitePublished          bool     `json:"site_published"`
		URL                    string   `json:"url"`
	} `json:"merchant"`
	Object string `json:"object"`
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (a *AccountResponse) UnmarshalJSON(b []byte) error {
	raw := accountResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "account" {
		a.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		a.Email = raw.Email
		a.ID = raw.ID
		m := &a.Merchant
		rm := &raw.Merchant
		m.BankEnabled = rm.BankEnabled
		m.BrandsAccepted = rm.BrandsAccepted
		m.BusinessType = rm.BusinessType
		m.ChargeType = rm.ChargeType
		m.ContactPhone = rm.ContactPhone
		m.Country = rm.Country
		m.CreatedAt = time.Unix(int64(rm.CreatedEpoch), 0)
		m.CurrenciesSupported = rm.CurrenciesSupported
		m.DefaultCurrency = rm.DefaultCurrency
		m.DetailsSubmitted = rm.DetailsSubmitted
		m.ID = rm.ID
		m.LiveModeActivatedAt = time.Unix(int64(rm.LiveModeActivatedEpoch), 0)
		m.LiveModeEnabled = rm.LiveModeEnabled
		m.ProductDetail = rm.ProductDetail
		m.ProductName = rm.ProductName
		m.ProductType = rm.ProductType
		m.SitePublished = rm.SitePublished
		m.URL = rm.URL
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
