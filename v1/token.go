package payjp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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

func (t TokenService) Create(card Card) (*TokenResponse, error) {
	var errors []string
	if card.Number == "" {
		errors = append(errors, "Number is required")
	}
	if card.ExpMonth < 0 || card.ExpMonth > 12 {
		errors = append(errors, fmt.Sprintf("ExpMonth should be between 1 and 12 but %d", card.ExpMonth))
	}
	if card.ExpYear <= 0 {
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

func (t TokenService) Get(id string) (*TokenResponse, error) {
	return parseToken(t.service.get("/tokens/" + id))
}

type TokenResponse struct {
	Card      CardResponse
	CreatedAt time.Time
	ID        string
	LiveMode  bool
	Used      bool
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
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
