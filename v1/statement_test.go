package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var statementResponseJSONStr = `
{
  "created": 1695892351,
  "id": "st_xxx",
  "items": [
    {
      "amount": 3125,
      "name": "売上",
      "subject": "gross_sales",
      "tax_rate": "0.00"
    },
    {
      "amount": -75,
      "name": "決済手数料",
      "subject": "fee",
      "tax_rate": "0.00"
    }
  ],
  "livemode": true,
  "object": "statement",
  "title": null,
  "updated": 1695892351
}
`
var statementResponseJSON = []byte(statementResponseJSONStr)

var statementListResponseJSON = []byte(`
{
  "count": 1,
  "data": [` + statementResponseJSONStr +
	`],
  "has_more": true,
  "object": "list",
  "url": "/v1/statements"
}
`)

var statementUrlsResponseJSON = []byte(`
{
  "expires": 1695903280,
  "object": "statement_url",
  "url": "url"
}
`)

var statementTitleJSON = []byte(`{
  "title": "title",
  "object": "statement"
}
`)

func TestParseStatementResponseJSON(t *testing.T) {
	service := &Service{}
	s := &StatementResponse{
		service: service,
	}
	err := json.Unmarshal(statementResponseJSON, s)

	assert.NoError(t, err)
	assert.Equal(t, "st_xxx", s.ID)
	assert.True(t, s.LiveMode)
	assert.Equal(t, "statement", s.Object)
	assert.Nil(t, s.Title)
	assert.Equal(t, 1695892351, *s.Created)
	assert.IsType(t, s.Updated, s.Created)
	assert.IsType(t, time.Unix(0, 0), s.UpdatedAt)
	assert.IsType(t, s.CreatedAt, s.UpdatedAt)
	assert.Equal(t, 3125, s.Items[0].Amount)
	assert.Equal(t, "売上", s.Items[0].Name)
	assert.Equal(t, "fee", s.Items[1].Subject)
	assert.Equal(t, "0.00", s.Items[1].TaxRate)
	assert.Equal(t, "0.00", s.Items[1].TaxRate)
	assert.Equal(t, service, s.service)
}

func TestParseStatementTitle(t *testing.T) {
	s := &StatementResponse{}
	err := json.Unmarshal(statementTitleJSON, s)
	assert.NoError(t, err)
	assert.Equal(t, "title", *s.Title)
}

func TestStatementUrls(t *testing.T) {
	mock, transport := newMockClient(200, statementResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, statementUrlsResponseJSON)
	service := New("api-key", mock)

	statement, err := service.Statement.Retrieve("st_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/statements/st_xxx", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, statement)
	assert.Equal(t, "gross_sales", statement.Items[0].Subject)

	_, err = service.Statement.Retrieve("st_xxx")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())

	url, err := statement.StatementUrls(StatementUrls{
		Platformer: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/statements/st_xxx/statement_urls", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "platformer=true", *transport.Body)
	assert.NotNil(t, url)
	assert.Equal(t, "url", url.URL)
}

func TestListStatement(t *testing.T) {
	mock, transport := newMockClient(200, statementListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	params := &StatementListParams{
		ListParams: ListParams{
			Limit: Int(1),
		},
		Owner:          String("tenant"),
		SourceTransfer: String("ten_tr_xxx"),
		Tenant:         String("test"),
	}
	statements, hasMore, err := service.Statement.List(params).
		Since(time.Unix(1, 0)).Do()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/statements?limit=1&owner=tenant&since=1&source_transfer=ten_tr_xxx&tenant=test", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.True(t, hasMore)
	assert.Equal(t, len(statements), 1)
	assert.Equal(t, "st_xxx", statements[0].ID)
	assert.Equal(t, service, statements[0].service)

	_, hasMore, err = service.Statement.List().Do()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
