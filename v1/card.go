package payjp

import (
	"encoding/json"
	"time"
)

type Card struct {
	Number       string
	ExpMonth     int
	ExpYear      int
	CVC          int
	AddressState string
	AddressCity  string
	AddressLine1 string
	AddressLine2 string
	AddressZip   string
	Country      string
	Name         string
}

func parseCard(service *Service, body []byte, result *CardResponse, customerID string) (*CardResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	result.customerID = customerID
	return result, nil
}

type CardResponse struct {
	AddressCity     string `json:"address_city"`
	AddressLine1    string `json:"address_line1"`
	AddressLine2    string `json:"address_line2"`
	AddressState    string `json:"address_state"`
	AddressZip      string `json:"address_zip"`
	AddressZipCheck string `json:"address_zip_check"`
	Brand           string `json:"brand"`
	Country         string `json:"country"`
	CreatedEpoch    int    `json:"created"`
	CvcCheck        string `json:"cvc_check"`
	ExpMonth        int    `json:"exp_month"`
	ExpYear         int    `json:"exp_year"`
	Fingerprint     string `json:"fingerprint"`
	ID              string `json:"id"`
	Last4           string `json:"last4"`
	Name            string `json:"name"`
	Object          string `json:"object"`

	CreatedAt  time.Time
	customerID string
	service    *Service
}

func (c *CardResponse) Update(card Card) error {
	_, err := c.service.Customer.postCard(c.customerID, "/"+c.ID, card, c)
	return err
}

func (c *CardResponse) Delete() error {
	return c.service.delete("/customers/" + c.customerID + "/cards/" + c.ID)
}

type cardResponse CardResponse

func (c *CardResponse) UnmarshalJSON(b []byte) error {
	raw := cardResponse{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "card" {
		*c = CardResponse(raw)
		c.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

type CardList struct {
	Count   int             `json:"count"`
	Data    []*CardResponse `json:"data"`
	HasMore bool            `json:"has_more"`
	Object  string          `json:"object"`
	URL     string          `json:"url"`
}

type cardList CardList

func (c *CardList) UnmarshalJSON(b []byte) error {
	raw := cardList{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "list" {
		*c = CardList(raw)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
