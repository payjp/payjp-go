package payjp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type accountService struct {
	service *Service
}

func newAccountService(service *Service) *accountService {
	return &accountService{service}
}

func (t *accountService) Get() (*Account, error) {
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
	result := &Account{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type Account struct {
	CreatedEpoch int              `json:"created"`
	Customer     CustomerResponse `json:"customer"`
	Email        string           `json:"email"`
	ID           string           `json:"id"`
	Merchant     struct {
		BankEnabled            bool        `json:"bank_enabled"`
		BrandsAccepted         []string    `json:"brands_accepted"`
		BusinessType           interface{} `json:"business_type"`
		ChargeType             interface{} `json:"charge_type"`
		ContactPhone           string      `json:"contact_phone"`
		Country                string      `json:"country"`
		CreatedEpoch           int         `json:"created"`
		CurrenciesSupported    []string    `json:"currencies_supported"`
		DefaultCurrency        string      `json:"default_currency"`
		DetailsSubmitted       bool        `json:"details_submitted"`
		ID                     string      `json:"id"`
		LiveModeActivatedEpoch int         `json:"livemode_activated_at"`
		LiveModeEnabled        bool        `json:"livemode_enabled"`
		Object                 string      `json:"object"`
		ProductDetail          string      `json:"product_detail"`
		ProductName            string      `json:"product_name"`
		ProductType            interface{} `json:"product_type"`
		SitePublished          bool        `json:"site_published"`
		URL                    string      `json:"url"`

		CreatedAt           time.Time
		LiveModeActivatedAt time.Time
	} `json:"merchant"`
	Object string `json:"object"`

	CreatedAt time.Time
}

type account Account

func (a *Account) UnmarshalJSON(b []byte) error {
	raw := account{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "account" {
		*a = Account(raw)
		a.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		a.Merchant.CreatedAt = time.Unix(int64(raw.Merchant.CreatedEpoch), 0)
		a.Merchant.LiveModeActivatedAt = time.Unix(int64(raw.Merchant.LiveModeActivatedEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
