package payjp

import (
	"encoding/json"
	"errors"
	"time"
)

// Card はCustomerやTokenのAPIでカード情報を設定する時に使う構造体です
type Card struct {
	Name         interface{} // カード保有者名(e.g. YUI ARAGAKI)
	Number       interface{} // カード番号
	ExpMonth     int         // 有効期限月
	ExpYear      int         // 有効期限年
	CVC          int         // CVCコード
	Country      interface{} // 2桁のISOコード(e.g. JP)
	AddressZip   interface{} // 郵便番号
	AddressState interface{} // 都道府県
	AddressCity  interface{} // 市区町村
	AddressLine1 interface{} // 番地など
	AddressLine2 interface{} // 建物名など
}

func (c Card) valid() bool {
	_, ok := c.Number.(string)
	return ok && c.ExpYear > 0 && c.ExpMonth > 0
}

func (c Card) empty() bool {
	_, ok := c.Number.(string)
	return !ok && c.ExpYear == 0 && c.ExpMonth == 0
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

// CardResponse はCustomerやTokenのAPIが返す構造体です
type CardResponse struct {
	CreatedAt       time.Time // カード作成時のタイムスタンプ
	ID              string    // car_で始まる一意なオブジェクトを示す文字列
	Name            string    // カード保有者名(e.g. YUI ARAGAKI)
	Last4           string    // カード番号の下四桁
	ExpMonth        int       // 有効期限月
	ExpYear         int       // 有効期限年
	Brand           string    // カードブランド名(e.g. Visa)
	CvcCheck        string    // CVCコードチェックの結果
	Fingerprint     string    // このクレジットカード番号に紐づけられた一意（他と重複しない）キー
	Country         string    // 2桁のISOコード(e.g. JP)
	AddressZip      string    // 郵便番号
	AddressZipCheck string    // 郵便番号存在チェックの結果
	AddressState    string    // 都道府県
	AddressCity     string    // 市区町村
	AddressLine1    string    // 番地など
	AddressLine2    string    // 建物名など

	customerID string
	service    *Service
}

type cardResponseParser struct {
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
}

// Update メソッドはカードの内容を更新します
// Customer情報から得られるカードでしか更新はできません
func (c *CardResponse) Update(card Card) error {
	if c.customerID == "" {
		return errors.New("Token's card doens't support Update()")
	}
	_, err := c.service.Customer.postCard(c.customerID, "/"+c.ID, card, c)
	return err
}

// Delete メソッドは顧客に登録されているカードを削除します
// Customer情報から得られるカードでしか削除はできません
func (c *CardResponse) Delete() error {
	if c.customerID == "" {
		return errors.New("Token's card doens't support Delete()")
	}
	return c.service.delete("/customers/" + c.customerID + "/cards/" + c.ID)
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (c *CardResponse) UnmarshalJSON(b []byte) error {
	raw := cardResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "card" {
		c.AddressCity = raw.AddressCity
		c.AddressLine1 = raw.AddressLine1
		c.AddressLine2 = raw.AddressLine2
		c.AddressState = raw.AddressState
		c.AddressZip = raw.AddressZip
		c.AddressZipCheck = raw.AddressZipCheck
		c.Brand = raw.Brand
		c.Country = raw.Country
		c.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		c.CvcCheck = raw.CvcCheck
		c.ExpMonth = raw.ExpMonth
		c.ExpYear = raw.ExpYear
		c.Fingerprint = raw.Fingerprint
		c.ID = raw.ID
		c.Last4 = raw.Last4
		c.Name = raw.Name
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
