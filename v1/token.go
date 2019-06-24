package payjp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TokenService はカード情報を代替するトークンオブジェクトを扱います。
//
// トークンは、カード番号やCVCなどのセキュアなデータを隠しつつも、カードと同じように扱うことができます。
//
// 顧客にカードを登録するときや、支払い処理を行うときにカード代わりとして使用します。
//
// 一度使用したトークンは再び使用することはできませんが、
// 顧客にカードを登録すれば、顧客IDを支払い手段として用いることで、
// 何度でも同じカードで支払い処理ができるようになります。
type TokenService struct {
	service *Service
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
	return result, nil
}

// Create メソッドカード情報を指定して、トークンを生成します。
//
// トークンはサーバーサイドからのリクエストでも生成可能ですが、通常は チェックアウトや
// payjp.js を利用して、ブラウザ経由でパブリックキーとカード情報を指定して生成します。
// トークンは二度以上使用することができません。
//
// チェックアウトやpayjp.jsを使ったトークン化の実装方法については チュートリアル -
// カード情報のトークン化(https://pay.jp/docs/cardtoken)をご覧ください。
//
// Card構造体で引数を設定しますが、Number/ExpMonth/ExpYearが必須パラメータです。
func (t TokenService) Create(card Card) (*TokenResponse, error) {
	var errors []string
	if card.Number == nil {
		errors = append(errors, "Number is required")
	}
	if card.ExpMonth == nil {
		errors = append(errors, "ExpMonth is required")
	}
	if card.ExpYear == nil {
		errors = append(errors, "ExpYear is required")
	}
	if len(errors) != 0 {
		return nil, fmt.Errorf("payjp.Token.Create() parameter error: %s", strings.Join(errors, ", "))
	}
	qb := newRequestBuilder()
	qb.AddCard(card)

	request, err := http.NewRequest("POST", t.service.apiBase+"/tokens", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", t.service.apiKey)

	return parseToken(respToBody(t.service.Client.Do(request)))
}

// Retrieve token object. 特定のトークン情報を取得します。
func (t TokenService) Retrieve(id string) (*TokenResponse, error) {
	return parseToken(t.service.retrieve("/tokens/" + id))
}

// TokenResponse はToken.Create(), Token.Retrieve()が返す構造体です。
type TokenResponse struct {
	Card      CardResponse // クレジットカードの情報
	CreatedAt time.Time    // このトークン作成時間
	ID        string       // tok_で始まる一意なオブジェクトを示す文字列
	LiveMode  bool         // 本番環境かどうか
	Used      bool         // トークンが使用済みかどうか
}

type tokenResponseParser struct {
	Card         json.RawMessage `json:"card"`
	CreatedEpoch int             `json:"created"`
	ID           string          `json:"id"`
	LiveMode     bool            `json:"livemode"`
	Object       string          `json:"object"`
	Used         bool            `json:"used"`
	CreatedAt    time.Time
}

// UnmarshalJSON はJSONパース用の内部APIです。
func (t *TokenResponse) UnmarshalJSON(b []byte) error {
	raw := tokenResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "token" {
		json.Unmarshal(raw.Card, &t.Card)
		t.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		t.ID = raw.ID
		t.LiveMode = raw.LiveMode
		t.Used = raw.Used
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
