package payjp

import (
	"net/http"
	"testing"
)

func TestNewClient(t *testing.T) {
	// default constructor
	service := New("sk_test_37dba67cf2cb5932eb4859af", nil)

	if service == nil {
		t.Error("service should be valid")
	}
	if service.ApiBase() != "https://api.pay.jp/v1" {
		t.Errorf(`ApiBase should be "https://api.pay.jp/v1", but "%s"`, service.ApiBase())
	}
}

func TestNewClientWithClient(t *testing.T) {
	// init with http.Client (to support proxy, etc)
	client := &http.Client{}
	service := New("sk_test_37dba67cf2cb5932eb4859af", client)

	if service == nil {
		t.Error("service should be valid")
	} else if service.Client != client {
		t.Error("service.Client should have passed client")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// init with http.Client (to support proxy, etc)
	client := &http.Client{}
	service := New("sk_test_37dba67cf2cb5932eb4859af", client, Config{
		ApiBase: "https://api.pay.jp/v2",
	})

	if service == nil {
		t.Error("service should be valid")
	} else if service.Client != client {
		t.Error("service.Client should have passed client")
	}
	if service.ApiBase() != "https://api.pay.jp/v2" {
		t.Errorf(`ApiBase should be "https://api.pay.jp/v2", but "%s"`, service.ApiBase())
	}
}
