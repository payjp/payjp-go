package payjp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
  "math/rand"
)

type requestBuilder struct {
	buffer    *bytes.Buffer
	delimiter byte
	hasValue  bool
}

func newRequestBuilder() *requestBuilder {
	return &requestBuilder{
		buffer:    &bytes.Buffer{},
		delimiter: '&',
	}
}

func (qb *requestBuilder) Add(key string, value interface{}) {
	if value == nil {
		return
	}
	var valueString string
	s, ok := value.(string)
	if ok {
		valueString = url.QueryEscape(s)
	} else {
		b, ok := value.(bool)
		if ok {
			if b {
				valueString = "true"
			} else {
				valueString = "false"
			}
		} else {
			valueString = strconv.Itoa(value.(int))
		}
	}
	if qb.hasValue {
		qb.buffer.WriteByte(qb.delimiter)
	}
	qb.hasValue = true
	qb.buffer.WriteString(key)
	qb.buffer.WriteByte('=')
	qb.buffer.WriteString(valueString)
}

func (qb *requestBuilder) AddCard(card Card) {
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
	qb.AddMetadata(card.Metadata)
}

func (qb *requestBuilder) AddMetadata(metadata map[string]string) {
	for key, value := range metadata {
		qb.Add("metadata["+key+"]", value)
	}
}

func (qb *requestBuilder) Reader() io.Reader {
	return qb.buffer
}

func respToBody(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type listResponseParser struct {
	Count   int               `json:"count"`
	Data    []json.RawMessage `json:"data"`
	HasMore bool              `json:"has_more"`
	Object  string            `json:"object"`
	URL     string            `json:"url"`
}

type listParser listResponseParser

func (p *listResponseParser) UnmarshalJSON(b []byte) error {
	raw := listParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "list" {
		*p = listResponseParser(raw)
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

func RandUniform(min, max float64) float64 {
  return (rand.Float64() * (max - min)) + min
}
