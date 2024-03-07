package payjp

import (
	"encoding/json"
	"time"
)

// Card はCustomerやTokenのAPIでカード情報を設定する時に使う構造体です
type Card struct {
	Name         interface{}       // カード保有者名(e.g. YUI ARAGAKI)
	Country      interface{}       // 2桁のISOコード(e.g. JP)
	AddressZip   interface{}       // 郵便番号
	AddressState interface{}       // 都道府県
	AddressCity  interface{}       // 市区町村
	AddressLine1 interface{}       // 番地など
	AddressLine2 interface{}       // 建物名など
	Metadata     map[string]string // メタデータ
}

// CardResponse はCustomerやTokenのAPIが返す構造体です
type CardResponse struct {
	CreatedAt          time.Time         // カード作成時のタイムスタンプ
	Created            *int              `json:"created"`
	AddressCity        string            `json:"address_city"`
	AddressLine1       string            `json:"address_line1"`
	AddressLine2       string            `json:"address_line2"`
	AddressState       string            `json:"address_state"`
	AddressZip         string            `json:"address_zip"`
	AddressZipCheck    string            `json:"address_zip_check"`
	Brand              string            `json:"brand"`
	Country            string            `json:"country"`
	CvcCheck           string            `json:"cvc_check"`
	ExpMonth           int               `json:"exp_month"`
	ExpYear            int               `json:"exp_year"`
	Fingerprint        string            `json:"fingerprint"`
	ID                 string            `json:"id"`
	Last4              string            `json:"last4"`
	Name               string            `json:"name"`
	LiveMode           bool              `json:"livemode"`
	Object             string            `json:"object"`
	Metadata           map[string]string `json:"metadata"`
	ThreeDSecureStatus *string           `json:"three_d_secure_status"`

	customerID string
	service    *Service
}

// Update メソッドはカードの内容を更新します
// Customer情報から得られるカードでしか更新はできません
func (c *CardResponse) Update(card Card) error {
	r, err := c.service.Customer.UpdateCard(c.customerID, c.ID, card)
	if err != nil {
		return err
	}
	*c = *r
	return nil
}

// Delete メソッドは顧客に登録されているカードを削除します
// Customer情報から得られるカードでしか削除はできません
func (c *CardResponse) Delete() error {
	if c.customerID == "" {
		panic("token's card doens't support Delete()")
	}
	return c.service.delete("/customers/" + c.customerID + "/cards/" + c.ID)
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (c *CardResponse) UnmarshalJSON(b []byte) error {
	type cardResponseParser CardResponse
	var raw cardResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "card" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.service = c.service
		raw.customerID = c.customerID
		*c = CardResponse(raw)
		return nil
	}
	return parseError(b)
}
