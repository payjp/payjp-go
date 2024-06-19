package payjp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeBalanceJSONStr(bank string, d string) string {
	return `{
  "net": 1000,
  "due_date": ` + d + `,
  "statements": ` + statementListResponseJSONStr + `,
  "created": 1438354800,
  "type": "collecting",
  "id": "ba_xxx",
  "livemode": false,
  "closed": false,
  "object": "balance",
  "bank_info": ` + bank + `}`
}

var balanceJSONStr = makeBalanceJSONStr("null", "null")

var balanceResponseJSON = []byte(balanceJSONStr)

var balanceListJSONStr = `
{
  "count": 1,
  "data": [` + balanceJSONStr +
	`],
  "has_more": false,
  "object": "list",
  "url": "/v1/balances"
}`
var balanceListResponseJSON = []byte(balanceListJSONStr)

func TestParseBalance(t *testing.T) {
	service := &Service{}
	s := &BalanceResponse{
		service: service,
	}
	err := json.Unmarshal(balanceResponseJSON, s)

	assert.NoError(t, err)
	assert.Equal(t, "ba_xxx", s.ID)
	assert.False(t, s.LiveMode)
	assert.Equal(t, "balance", s.Object)
	assert.Equal(t, 1438354800, *s.Created)
	assert.IsType(t, time.Unix(0, 0), s.CreatedAt)
	assert.EqualValues(t, 1000, s.Net)
	assert.True(t, s.RawStatements.HasMore)
	assert.Equal(t, "st_xxx", s.Statements[0].ID)
	assert.Equal(t, (*int)(nil), s.RawDueDate)
	assert.Equal(t, time.Unix(0, 0), s.DueDate)
	assert.Equal(t, "collecting", s.Type)
	assert.False(t, s.Closed)
	assert.Nil(t, s.BankInfo)
	assert.Equal(t, service, s.service)

	balanceJSONStr2 := makeBalanceJSONStr(`{
	"bank_code": "0001",
	"bank_branch_code": "123",
	"bank_account_type": "普通",
	"bank_account_number": "1234567",
	"bank_account_holder_name": "ペイ　タロウ",
	"bank_account_status": "pending"
}`, `1711897200`)
	err = json.Unmarshal([]byte(balanceJSONStr2), s)
	assert.NoError(t, err)
	assert.Equal(t, 1711897200, *s.RawDueDate)
	assert.Equal(t, 1711897200, (int)(s.DueDate.Unix()))
	assert.Equal(t, "0001", s.BankInfo.BankCode)
	assert.Equal(t, "123", s.BankInfo.BankBranchCode)
	assert.Equal(t, "普通", s.BankInfo.BankAccountType)
	assert.Equal(t, "1234567", s.BankInfo.BankAccountNumber)
	assert.Equal(t, "ペイ　タロウ", s.BankInfo.BankAccountHolderName)
	assert.Equal(t, "pending", s.BankInfo.BankAccountStatus)
}

func TestBalanceList(t *testing.T) {
	mock, transport := newMockClient(200, balanceListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	limit := Int(10)
	params := &BalanceListParams{
		ListParams: ListParams{
			Offset: Int(0),
		},
		SinceDueDate: Int(1455328095),
		UntilDueDate: Int(1455328095),
		State:        String("collecting"),
		Closed:       Bool(true),
		Owner:        String("tenant"),
		Tenant:       String("ten_xxx"),
	}
	params.Limit = limit
	res, hasMore, err := service.Balance.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/balances?closed=true&limit=10&offset=0&owner=tenant&since_due_date=1455328095&state=collecting&tenant=ten_xxx&until_due_date=1455328095", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.False(t, hasMore)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, "ba_xxx", res[0].ID)
	assert.Equal(t, service, res[0].service)

	_, hasMore, err = service.Balance.All()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestBalance(t *testing.T) {
	mock, transport := newMockClient(200, balanceResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	v, err := service.Balance.Retrieve("ba_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/balances/ba_xxx", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, v)
	assert.Equal(t, "ba_xxx", v.ID)

	_, err = service.Balance.Retrieve("ba_xxx")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestBalanceStatementUrls(t *testing.T) {
	mock, transport := newMockClient(200, balanceResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	transport.AddResponse(200, statementUrlsResponseJSON)
	transport.AddResponse(200, statementUrlsResponseJSON)
	transport.AddResponse(200, statementUrlsResponseJSON)
	service := New("api-key", mock)

	balance, err := service.Balance.Retrieve("ba_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/balances/ba_xxx", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, balance)

	_, err = service.Balance.Retrieve("ba_xxx")
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())

	url, err := balance.StatementUrls(StatementUrls{
		Platformer: Bool(true),
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/balances/ba_xxx/statement_urls", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, "platformer=true", *transport.Body)
	assert.NotNil(t, url)
	assert.Equal(t, "url", url.URL)

	_, err = balance.StatementUrls()
	assert.NoError(t, err)
	assert.Equal(t, "", *transport.Body)

	_, err = balance.StatementUrls(StatementUrls{})
	assert.NoError(t, err)
	assert.Equal(t, "", *transport.Body)
}
