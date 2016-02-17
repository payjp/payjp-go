package payjp

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
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

func (qb *requestBuilder) Add(key, value string) {
	if qb.hasValue {
		qb.buffer.WriteByte(qb.delimiter)
	}
	qb.hasValue = true
	qb.buffer.WriteString(key)
	qb.buffer.WriteByte('=')
	if value != EmptyString {
		qb.buffer.WriteString(url.QueryEscape(value))
	}
}

func (qb *requestBuilder) AddCard(card Card) {
	if card.Number != "" {
		qb.Add("card[number]", card.Number)
	}
	if card.ExpMonth != 0 {
		qb.Add("card[exp_month]", strconv.Itoa(card.ExpMonth))
	}
	if card.ExpYear != 0 {
		qb.Add("card[exp_year]", strconv.Itoa(card.ExpYear))
	}
	if card.CVC != 0 {
		qb.Add("card[cvc]", strconv.Itoa(card.CVC))
	}
	if card.AddressState != "" {
		qb.Add("card[address_state]", card.AddressState)
	}
	if card.AddressCity != "" {
		qb.Add("card[address_city]", card.AddressCity)
	}
	if card.AddressLine1 != "" {
		qb.Add("card[address_line1]", card.AddressLine1)
	}
	if card.AddressLine2 != "" {
		qb.Add("card[address_line2]", card.AddressLine2)
	}
	if card.AddressZip != "" {
		qb.Add("card[address_zip]", card.AddressZip)
	}
	if card.Country != "" {
		qb.Add("card[country]", card.Country)
	}
	if card.Name != "" {
		qb.Add("card[name]", card.Name)
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
