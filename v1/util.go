package payjp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
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
	var valueString string
	switch v := value.(type) {
	case nil:
		return
	case int:
		valueString = strconv.Itoa(v)
	case *int:
		valueString = strconv.Itoa(*v)
	case string:
		valueString = url.QueryEscape(v)
	case *string:
		valueString = url.QueryEscape(*v)
	case bool:
		if v {
			valueString = "true"
		} else {
			valueString = "false"
		}
	default:
		panic(`invalid parameter type of '` + key + `'`)
	}

	if qb.hasValue {
		qb.buffer.WriteByte(qb.delimiter)
	}
	qb.hasValue = true
	qb.buffer.WriteString(key)
	qb.buffer.WriteByte('=')
	qb.buffer.WriteString(valueString)
}

func (qb *requestBuilder) AddMetadata(metadata map[string]string) {
	for key, value := range metadata {
		qb.Add("metadata["+key+"]", value)
	}
}

func (qb *requestBuilder) Reader() io.Reader {
	return qb.buffer
}

// DeleteResponse
type DeleteResponse struct {
	Deleted  bool   `json:"deleted"`
	ID       string `json:"id"`
	LiveMode bool   `json:"livemode"`
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
	return parseError(b)
}

type payjpResponse struct {
	Object  *string `json:"object"`
	Deleted *bool   `json:"deleted"`
	Parser  interface{}
}

func (e *payjpResponse) UnmarshalJSON(b []byte) error {
	type payjpParser payjpResponse
	var raw payjpParser
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return parseError(b)
	} else if raw.Deleted != nil && *(raw.Deleted) {
		e.Parser = DeleteResponse{}
		return nil
	}
	switch *(raw.Object) {
	case "token":
		e.Parser = TokenResponse{}
	case "charge":
		e.Parser = ChargeResponse{}
	case "customer":
		e.Parser = CustomerResponse{}
	case "card":
		e.Parser = CardResponse{}
	case "plan":
		e.Parser = PlanResponse{}
	case "subscription":
		e.Parser = SubscriptionResponse{}
	case "transfer":
		e.Parser = TransferResponse{}
	case "statement":
		e.Parser = StatementResponse{}
	default:
		return nil
	}
	return nil
}
