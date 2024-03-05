package payjp

import (
	"encoding/json"
	"time"
)

// TokenService はカード情報を代替するトークンオブジェクトを扱います。
type TokenService struct {
	service *Service
}

type Token struct {
	Card
	Number   interface{} // カード番号
	ExpMonth interface{} // 有効期限月
	ExpYear  interface{} // 有効期限年
	CVC      interface{} // CVCコード
}

func newTokenService(service *Service) *TokenService {
	return &TokenService{
		service: service,
	}
}

func parseToken(data []byte, err error) (*TokenResponse, error) {
	if err != nil {
		return nil, err
	}
	result := &TokenResponse{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, err
}

// Create はテストモードのみ利用可能で、トークンを生成します。
func (t TokenService) Create(card Token) (*TokenResponse, error) {
	qb := newRequestBuilder()
	qb.Add("card[number]", card.Number)
	qb.Add("card[exp_month]", card.ExpMonth)
	qb.Add("card[exp_year]", card.ExpYear)
	qb.Add("card[cvc]", card.CVC)
	qb.Add("card[address_state]", card.AddressState)
	qb.Add("card[address_city]", card.AddressCity)
	qb.Add("card[address_line1]", card.AddressLine1)
	qb.Add("card[address_line2]", card.AddressLine2)
	qb.Add("card[address_zip]", card.AddressZip)
	qb.Add("card[country]", card.Country)
	qb.Add("card[name]", card.Name)

	return parseToken(t.service.request("POST", "/tokens", qb.Reader()))
}

// Retrieve token object. 特定のトークン情報を取得します。
func (t TokenService) Retrieve(id string) (*TokenResponse, error) {
	return parseToken(t.service.request("GET", "/tokens/"+id, nil))
}

// TokenResponse はToken.Create(), Token.Retrieve()が返す構造体です。
type TokenResponse struct {
	CreatedAt time.Time       // このトークン作成時間
	RawCard   json.RawMessage `json:"card"`
	Card      CardResponse    // クレジットカードの情報
	Created   *int            `json:"created"`
	ID        string          `json:"id"`
	LiveMode  bool            `json:"livemode"`
	Object    string          `json:"object"`
	Used      bool            `json:"used"`

	service *Service
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (t *TokenResponse) UnmarshalJSON(b []byte) error {
	type tokenResponseParser TokenResponse
	var raw tokenResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "token" {
		json.Unmarshal(raw.RawCard, &raw.Card)
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)

		raw.service = t.service
		*t = TokenResponse(raw)
		return nil
	}
	return parseError(b)
}
