package payjp

import (
	"encoding/json"
	"time"
)

// AccountService はあなたのアカウント情報を返します。
type AccountService struct {
	service *Service
}

func newAccountService(service *Service) *AccountService {
	return &AccountService{service}
}

// Retrieve account object. あなたのアカウント情報を取得します。
func (t *AccountService) Retrieve() (*AccountResponse, error) {
	body, err := t.service.request("GET", "/accounts", nil)
	result := &AccountResponse{}
	err = json.Unmarshal(body, result)
	return result, err
}

// AccountResponse はAccount.Retrieve()メソッドが返す構造体です
type AccountResponse struct {
	CreatedAt time.Time // このアカウント作成時のタイムスタンプ
	Created   *int      `json:"created"`
	Email     string    `json:"email"`
	ID        string    `json:"id"`
	Merchant  struct {
		BankEnabled            bool      `json:"bank_enabled"`
		BrandsAccepted         []string  `json:"brands_accepted"`
		BusinessType           string    `json:"business_type"`
		ChargeType             []string  `json:"charge_type"`
		ContactPhone           string    `json:"contact_phone"`
		Country                string    `json:"country"`
		Created                *int      `json:"created"`
		CreatedAt              time.Time // 作成時のタイムスタンプ
		CurrenciesSupported    []string  `json:"currencies_supported"`
		DefaultCurrency        string    `json:"default_currency"`
		DetailsSubmitted       bool      `json:"details_submitted"`
		ID                     string    `json:"id"`
		RawLiveModeActivatedAt *int      `json:"livemode_activated_at"`
		LiveModeActivatedAt    time.Time
		LiveModeEnabled        bool     `json:"livemode_enabled"`
		Object                 string   `json:"object"`
		ProductDetail          string   `json:"product_detail"`
		ProductName            string   `json:"product_name"`
		ProductType            []string `json:"product_type"`
		SitePublished          bool     `json:"site_published"`
		URL                    string   `json:"url"`
	} `json:"merchant"`
	Object  string `json:"object"`
	TeamID  string `json:"team_id"`
	service *Service
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (a *AccountResponse) UnmarshalJSON(b []byte) error {
	type accountResponseParser AccountResponse
	var raw accountResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "account" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.Merchant.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.Merchant.LiveModeActivatedAt = time.Unix(IntValue(raw.Merchant.RawLiveModeActivatedAt), 0)
		raw.service = a.service
		*a = AccountResponse(raw)
		return nil
	}
	return parseError(b)
}
