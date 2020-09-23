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
	if service.APIBase() != "https://api.pay.jp/v1" {
		t.Errorf(`ApiBase should be "https://api.pay.jp/v1", but "%s"`, service.APIBase())
	}
	defaultRetryConfig := RetryConfig{0, 2, 32}
	if service.RetryConfig() != defaultRetryConfig {
		t.Errorf(`RetryConfig should be %v, but %v`, defaultRetryConfig, service.RetryConfig())
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

func TestNewClientWithOptions(t *testing.T) {
	// init with http.Client (to support proxy, etc)
	client := &http.Client{}
	retryConfig := RetryConfig{1, 2, 30}
	service := New("sk_test_37dba67cf2cb5932eb4859af", client, OptionApiBase("https://api.pay.jp/v2"), OptionRetryConfig(retryConfig))

	if service == nil {
		t.Error("service should be valid")
	} else if service.Client != client {
		t.Error("service.Client should have passed client")
	}
	if service.APIBase() != "https://api.pay.jp/v2" {
		t.Errorf(`ApiBase should be "https://api.pay.jp/v2", but "%s"`, service.APIBase())
	}
	if service.RetryConfig() != retryConfig {
		t.Errorf(`RetryConfig should be %v, but %v`, retryConfig, service.RetryConfig())
	}

	retryConfig2 := RetryConfig{3, 4, 50}
	service2 := New("sk_test_37dba67cf2cb5932eb4859af", client, OptionRetryConfig(retryConfig2))
	if service2.APIBase() != "https://api.pay.jp/v1" {
		t.Errorf(`ApiBase should be "https://api.pay.jp/v1", but "%s"`, service2.APIBase())
	}
	if service2.RetryConfig() != retryConfig2 {
		t.Errorf(`RetryConfig should be %v, but %v`, retryConfig2, service2.RetryConfig())
	}
}
