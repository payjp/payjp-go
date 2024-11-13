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
	Number       interface{} // カード番号
	ExpMonth     interface{} // 有効期限月
	ExpYear      interface{} // 有効期限年
	CVC          interface{} // CVCコード
	ThreeDSecure *bool       // 3DSecureを実施するか否か
}

func newTokenService(service *Service) *TokenService {
	return &TokenService{
		service: service,
	}
}

func parseToken(service *Service, data []byte, err error) (*TokenResponse, error) {
	if err != nil {
		return nil, err
	}
	result := &TokenResponse{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	result.service = service
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
	qb.Add("card[email]", card.Email)
	qb.Add("card[phone]", card.Phone)
	qb.Add("three_d_secure", card.ThreeDSecure)

	body, err := t.service.request("POST", "/tokens", qb.Reader(), map[string]string{
		"X-Payjp-Direct-Token-Generate": "true",
	})

	return parseToken(t.service, body, err)
}

// Retrieve token object. 特定のトークン情報を取得します。
func (t TokenService) Retrieve(id string) (*TokenResponse, error) {
	body, err := t.service.request("GET", "/tokens/"+id, nil)
	return parseToken(t.service, body, err)
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

// TdsFinish を TokenResponse から実行します。
func (t *TokenResponse) TdsFinish() error {
	return t.updateResponse(t.service.Token.TdsFinish(t.ID))
}

func (t *TokenResponse) updateResponse(r *TokenResponse, err error) error {
	if err != nil {
		return err
	}
	*t = *r
	return nil
}

// TdsFinish は3Dセキュア認証が終了した支払いに対し、決済を行います。
// https://pay.jp/docs/api/#%E3%83%88%E3%83%BC%E3%82%AF%E3%83%B3%E3%81%AB%E5%AF%BE%E3%81%99%E3%82%8B3d%E3%82%BB%E3%82%AD%E3%83%A5%E3%82%A2%E3%83%95%E3%83%AD%E3%83%BC%E3%82%92%E5%AE%8C%E4%BA%86%E3%81%99%E3%82%8B
func (t TokenService) TdsFinish(id string) (*TokenResponse, error) {
	body, err := t.service.request("POST", "/tokens/"+id+"/tds_finish", nil)
	return parseToken(t.service, body, err)
}
