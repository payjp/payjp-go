package payjp

import (
	"encoding/json"
	"testing"
	"time"
)

var transferResponseJSON = []byte(`
{
  "amount": 1000,
  "carried_balance": null,
  "charges": {
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
        "object": "charge",
        "paid": true,
        "refund_reason": null,
        "refunded": false,
        "subscription": null
      }
    ],
    "has_more": false,
    "object": "list",
    "url": "/v1/transfers/tr_8f0c0fe2c9f8a47f9d18f03959ba1/charges"
  },
  "created": 1438354800,
  "currency": "jpy",
  "description": null,
  "id": "tr_8f0c0fe2c9f8a47f9d18f03959ba1",
  "livemode": false,
  "object": "transfer",
  "scheduled_date": "2015-09-16",
  "status": "pending",
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
}
`)

var recombinationTransferResponseJSON = []byte(`
{
  "amount": 1000,
  "carried_balance": null,
  "charges": {
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
        "object": "charge",
        "paid": true,
        "refund_reason": null,
        "refunded": false,
        "subscription": null
      }
    ],
    "has_more": false,
    "object": "list",
    "url": "/v1/transfers/tr_8f0c0fe2c9f8a47f9d18f03959ba1/charges"
  },
  "created": 1438354800,
  "currency": "jpy",
  "description": null,
  "id": "tr_8f0c0fe2c9f8a47f9d18f03959ba1",
  "livemode": false,
  "object": "transfer",
  "scheduled_date": "2015-09-16",
  "status": "recombination",
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
}
`)

var transferListResponseJSON = []byte(`
{
  "count": 1,
  "data": [
    {
      "amount": 1000,
      "carried_balance": null,
      "charges": {
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
            "object": "charge",
            "paid": true,
            "refund_reason": null,
            "refunded": false,
            "subscription": null
          }
        ],
        "has_more": false,
        "object": "list",
        "url": "/v1/transfers/tr_8f0c0fe2c9f8a47f9d18f03959ba1/charges"
      },
      "created": 1438354800,
      "currency": "jpy",
      "description": null,
      "id": "tr_8f0c0fe2c9f8a47f9d18f03959ba1",
      "livemode": false,
      "object": "transfer",
      "scheduled_date": "2015-09-16",
      "status": "pending",
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
    }
  ],
  "has_more": false,
  "object": "list",
  "url": "/v1/transfers"
}
`)

var transferChargeListResponseJSON = []byte(`
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
  "url": "/v1/transfers/tr_8f0c0fe2c9f8a47f9d18f03959ba1/charges"
}
`)

func TestParseTransferResponseJSON(t *testing.T) {
	transfer := &TransferResponse{}
	err := json.Unmarshal(transferResponseJSON, transfer)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if transfer.ID != "tr_8f0c0fe2c9f8a47f9d18f03959ba1" {
		t.Errorf("transfer.ID should be 'tr_8f0c0fe2c9f8a47f9d18f03959ba1', but '%s'", transfer.ID)
	}
	if transfer.Charges[0].ID != "ch_60baaf2dc8f3e35684ebe2031a6e0" {
		t.Errorf("Charge.ID should be 'ch_60baaf2dc8f3e35684ebe2031a6e0', but '%s'", transfer.Charges[0].ID)
	}
	if transfer.Status.status() != "pending" {
		t.Errorf("transfer.Status should be 'pending', but '%s'", transfer.Status.status())
	}
}

func TestParseRecombinationTransferResponseJSON(t *testing.T) {
	recombinationTransfer := &TransferResponse{}
	err := json.Unmarshal(recombinationTransferResponseJSON, recombinationTransfer)

	if err != nil {
		t.Errorf("err should be nil, but %v", err)
	}
	if recombinationTransfer.Status.status() != "recombination" {
		t.Errorf("transfer.Status should be 'recombination', but '%s'", recombinationTransfer.Status.status())
	}
}

func TestTransferRetrieve(t *testing.T) {
	mock, transport := newMockClient(200, transferResponseJSON)
	service := New("api-key", mock)
	transfer, err := service.Transfer.Retrieve("tr_8f0c0fe2c9f8a47f9d18f03959ba1")
	if transport.URL != "https://api.pay.jp/v1/transfers/tr_8f0c0fe2c9f8a47f9d18f03959ba1" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	} else if transfer == nil {
		t.Error("transfer should not be nil")
	} else if transfer.Summary.ChargeGross != 1000 {
		t.Errorf("transfer.Summary.ChargeGross should be 1000 but %d", transfer.Summary.ChargeGross)
	}
}

func TestTransferList(t *testing.T) {
	mock, transport := newMockClient(200, transferListResponseJSON)
	service := New("api-key", mock)
	subscriptions, hasMore, err := service.Transfer.List().
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).Do()
	if transport.URL != "https://api.pay.jp/v1/transfers?limit=10&offset=15&since=1455328095&until=1455500895" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if hasMore {
		t.Error("parse error: hasMore")
	}
	if len(subscriptions) != 1 {
		t.Error("parse error: plans")
	}
}

func TestTransferChargeList(t *testing.T) {
	mock, transport := newMockClient(200, transferChargeListResponseJSON)
	service := New("api-key", mock)
	subscriptions, hasMore, err := service.Transfer.ChargeList("tr_8f0c0fe2c9f8a47f9d18f03959ba1").
		Limit(10).
		Offset(15).
		Since(time.Unix(1455328095, 0)).
		Until(time.Unix(1455500895, 0)).
		CustomerID("cus_b92b879e60f62b532d6756ae12af").Do()
	if transport.URL != "https://api.pay.jp/v1/transfers/tr_8f0c0fe2c9f8a47f9d18f03959ba1/charges?customer=cus_b92b879e60f62b532d6756ae12af&limit=10&offset=15&since=1455328095&until=1455500895" {
		t.Errorf("URL is wrong: %s", transport.URL)
	}
	if transport.Method != "GET" {
		t.Errorf("Method should be GET, but %s", transport.Method)
	}
	if err != nil {
		t.Errorf("err should be nil, but %v", err)
		return
	}
	if hasMore {
		t.Error("parse error: hasMore")
	}
	if len(subscriptions) != 1 {
		t.Error("parse error: plans")
	}
}
