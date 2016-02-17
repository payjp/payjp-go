package payjp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type tokenService struct {
	service *Service
}

func newTokenService(service *Service) *tokenService {
	return &tokenService{
		service: service,
	}
}

func (t *tokenService) Create(card Card) (*Token, error) {
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

	resp, err := t.service.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	result := &Token{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *tokenService) Get(id string) (*Token, error) {
	body, err := t.service.get("/tokens/" + id)
	if err != nil {
		return nil, err
	}
	result := &Token{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type Token struct {
	Card         CardResponse `json:"card"`
	CreatedEpoch int          `json:"created"`
	ID           string       `json:"id"`
	LiveMode     bool         `json:"livemode"`
	Object       string       `json:"object"`
	Used         bool         `json:"used"`
	CreatedAt    time.Time
}

type token Token

func (t *Token) UnmarshalJSON(b []byte) error {
	raw := token{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "token" {
		*t = Token(raw)
		t.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
