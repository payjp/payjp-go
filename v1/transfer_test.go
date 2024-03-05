package payjp

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var chargeListJSONStr = `
{
  "count": 1,
  "data": [
    {
      "amount": 1000,
      "amount_refunded": 0,
      "captured": true,
      "captured_at": 1441706750,
      "card": {
        "address_city": null,
        "address_line1": null,
        "address_line2": null,
        "address_state": null,
        "address_zip": null,
        "address_zip_check": "unchecked",
        "brand": "Visa",
        "country": null,
        "created": 1441706750,
        "customer": "cus_b92b879e60f62b532d6756ae12af",
        "cvc_check": "unchecked",
        "exp_month": 5,
        "exp_year": 2018,
        "fingerprint": "e1d8225886e3a7211127df751c86787f",
        "id": "car_93e59e9a9714134ef639865e2b9e",
        "last4": "4242",
        "name": null,
        "object": "card"
      },
      "created": 1441706750,
      "currency": "jpy",
      "customer": "cus_b92b879e60f62b532d6756ae12af",
      "description": null,
      "expired_at": null,
      "failure_code": null,
      "failure_message": null,
      "id": "ch_60baaf2dc8f3e35684ebe2031a6e0",
      "livemode": false,
      "object": "charge",
      "paid": true,
      "refund_reason": null,
      "refunded": false,
      "subscription": null
    }
  ],
  "has_more": false,
  "object": "list",
  "url": "/v1/transfers/tr_xxx/charges"
}`
var transferChargeListResponseJSON = []byte(chargeListJSONStr)

func makeTransferJSONStr(t TransferStatus) string {
	return `
{
  "amount": 1000,
  "carried_balance": null,
  "charges": ` + chargeListJSONStr + `,
  "created": 1438354800,
  "currency": "jpy",
  "description": "test",
  "id": "tr_xxx",
  "livemode": false,
  "object": "transfer",
  "scheduled_date": "2015-09-16",
  "status": "` + string(t) + `",
  "summary": {
    "charge_count": 1,
    "charge_fee": 0,
    "charge_gross": 1000,
    "net": 1000,
    "refund_amount": 0,
    "refund_count": 0,
    "dispute_amount": 0,
    "dispute_count": 0
  },
  "term_end": 1439650800,
  "term_start": 1438354800,
  "transfer_amount": null,
  "transfer_date": null
}`
}

var transferJSONStr = makeTransferJSONStr(TransferPending)

var transferResponseJSON = []byte(transferJSONStr)

var transferListJSONStr = `
{
  "count": 1,
  "data": [` + transferJSONStr +
	`],
  "has_more": false,
  "object": "list",
  "url": "/v1/transfers"
}`
var transferListResponseJSON = []byte(transferListJSONStr)

func TestParseTransferResponseJSON(t *testing.T) {
	service := &Service{}
	s := &TransferResponse{
		service: service,
	}
	err := json.Unmarshal(transferResponseJSON, s)

	assert.NoError(t, err)
	assert.Equal(t, "tr_xxx", s.ID)
	assert.False(t, s.LiveMode)
	assert.Equal(t, "transfer", s.Object)
	assert.Equal(t, 1438354800, *s.Created)
	assert.IsType(t, time.Unix(0, 0), s.CreatedAt)
	assert.Equal(t, 1000, s.Amount)
	assert.Nil(t, s.RawCarriedBalance)
	assert.Equal(t, 0, s.CarriedBalance)
	assert.Equal(t, "jpy", s.Currency)
	assert.Equal(t, TransferPending, s.Status)
	assert.False(t, s.RawCharges.HasMore)
	assert.Equal(t, "ch_60baaf2dc8f3e35684ebe2031a6e0", s.Charges[0].ID)
	assert.Equal(t, "2015-09-16", s.ScheduledDate)
	assert.Equal(t, "test", s.Description)
	assert.Equal(t, 1438354800, *s.TermStart)
	assert.IsType(t, time.Unix(0, 0), s.TermStartAt)
	assert.Equal(t, 1439650800, *s.TermEnd)
	assert.IsType(t, time.Unix(0, 0), s.TermEndAt)
	assert.Nil(t, s.RawTransferAmount)
	assert.Nil(t, s.RawTransferDate)
	assert.Equal(t, 0, s.TransferAmount)
	assert.Equal(t, "", s.TransferDate)
	assert.Equal(t, service, s.service)
}

func TestParseTransferStatusResponseJSON(t *testing.T) {
	transfer := &TransferResponse{}

	response := []byte(makeTransferJSONStr(TransferRecombination))
	err := json.Unmarshal(response, transfer)
	assert.NoError(t, err)
	assert.Equal(t, TransferRecombination, transfer.Status)
	assert.Equal(t, "recombination", transfer.Status.status())

	response = []byte(makeTransferJSONStr(TransferCarriedOver))
	err = json.Unmarshal(response, transfer)
	assert.NoError(t, err)
	assert.Equal(t, "carried_over", transfer.Status.status())

	response = []byte(makeTransferJSONStr(TransferStop))
	err = json.Unmarshal(response, transfer)
	assert.NoError(t, err)
	assert.Equal(t, "stop", transfer.Status.status())
}

func TestTransferList(t *testing.T) {
	mock, transport := newMockClient(200, transferListResponseJSON)
	transport.AddResponse(200, transferListResponseJSON)
	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	status := TransferPending
	// deprecated だが後方互換で残すリクエスト方法
	res, hasMore, err := service.Transfer.List().
		Limit(10).
		Offset(15).
		SinceSheduledDate(time.Unix(1455328095, 0)).
		UntilSheduledDate(time.Unix(1455500895, 0)).
		Status(status).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/transfers?limit=10&offset=15&since=1455328095&since_scheduled_date=1455328095&status=pending&until=1455500895&until_scheduled_date=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.False(t, hasMore)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, "tr_xxx", res[0].ID)
	assert.Equal(t, service, res[0].service)

	params := &TransferListParams{
		ListParams: ListParams{
			Limit:  Int(10),
			Offset: Int(0),
		},
		SinceSheduledDate: Int(1455328095),
		Status:            &status,
	}
	res, hasMore, err = service.Transfer.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/transfers?limit=10&offset=0&since_scheduled_date=1455328095&status=pending", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.False(t, hasMore)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, "tr_xxx", res[0].ID)
	assert.Equal(t, service, res[0].service)

	_, hasMore, err = service.Transfer.All()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}

func TestTransferChargeList(t *testing.T) {
	mock, transport := newMockClient(200, transferResponseJSON)
	transport.AddResponse(200, transferChargeListResponseJSON)

	transport.AddResponse(200, transferResponseJSON)

	transport.AddResponse(200, transferChargeListResponseJSON)

	transport.AddResponse(400, errorResponseJSON)
	service := New("api-key", mock)

	// deprecated だが後方互換で残すリクエスト方法
	res, hasMore, err := service.Transfer.ChargeList("tr_xxx").
		Limit(10).
		Offset(15).
		CustomerID("cus_xxx").
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/transfers/tr_xxx/charges?customer=cus_xxx&limit=10&offset=15&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.False(t, hasMore)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, "ch_60baaf2dc8f3e35684ebe2031a6e0", res[0].ID)
	assert.Equal(t, service, res[0].service)

	transfer, err := service.Transfer.Retrieve("tr_xxx")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/transfers/tr_xxx", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.NotNil(t, transfer)
	assert.Equal(t, 1000, transfer.Summary.ChargeGross)

	params := &TransferChargeListParams{
		ListParams: ListParams{
			Limit:  Int(10),
			Offset: Int(15),
			Since:  Int(1455328095),
			Until:  Int(1455500895),
		},
		Customer: String("cus_xxx"),
	}
	res, hasMore, err = transfer.All(params)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/transfers/tr_xxx/charges?customer=cus_xxx&limit=10&offset=15&since=1455328095&until=1455500895", transport.URL)
	assert.Equal(t, "GET", transport.Method)
	assert.False(t, hasMore)
	assert.Equal(t, len(res), 1)
	assert.Equal(t, "ch_60baaf2dc8f3e35684ebe2031a6e0", res[0].ID)
	assert.Equal(t, service, res[0].service)

	_, hasMore, err = service.Transfer.All()
	assert.False(t, hasMore)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
}
