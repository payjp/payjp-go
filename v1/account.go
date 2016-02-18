package payjp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type AccountService struct {
	service *Service
}

func newAccountService(service *Service) *AccountService {
	return &AccountService{service}
}

func (t *AccountService) Get() (*AccountResponse, error) {
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

type AccountResponse struct {
	CreatedAt time.Time
	Customer  json.RawMessage
	Email     string
	ID        string
	Merchant  struct {
		BankEnabled         bool
		BrandsAccepted      []string
		BusinessType        interface{}
		ChargeType          interface{}
		ContactPhone        string
		Country             string
		CreatedAt           time.Time
		CurrenciesSupported []string
		DefaultCurrency     string
		DetailsSubmitted    bool
		ID                  string
		LiveModeActivatedAt time.Time
		LiveModeEnabled     bool
		ProductDetail       string
		ProductName         string
		ProductType         interface{}
		SitePublished       bool
		URL                 string
	}
}

type accountResponseParser struct {
	CreatedEpoch int             `json:"created"`
	Customer     json.RawMessage `json:"customer"`
	Email        string          `json:"email"`
	ID           string          `json:"id"`
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
	} `json:"merchant"`
	Object string `json:"object"`
}

func (a *AccountResponse) UnmarshalJSON(b []byte) error {
	raw := accountResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "account" {
		a.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		json.Unmarshal(raw.Customer, &a.Customer)
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
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
