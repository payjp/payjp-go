package payjp

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	defaultRetryConfig := RetryConfig{0, 2, 32, nil}
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
	retryConfig := RetryConfig{1, 2, 30, nil}
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

	retryConfig2 := RetryConfig{3, 4, 50, nil}
	service2 := New("sk_test_37dba67cf2cb5932eb4859af", client, OptionRetryConfig(retryConfig2))
	if service2.APIBase() != "https://api.pay.jp/v1" {
		t.Errorf(`ApiBase should be "https://api.pay.jp/v1", but "%s"`, service2.APIBase())
	}
	if service2.RetryConfig() != retryConfig2 {
		t.Errorf(`RetryConfig should be %v, but %v`, retryConfig2, service2.RetryConfig())
	}
}

func TestRetryConfigLogging(t *testing.T) {
	var buf bytes.Buffer
	prefix := "TestRetryConfigLoggin_"
	logger := log.New(&buf, prefix, log.Ldate)
	config := RetryConfig{0, 2, 32, logger}
	msg := "this-is-log-message"
	config.Logger.Printf(msg)
	result := buf.String()
	if !strings.Contains(result, msg) {
		t.Errorf("Failed to log")
	}
}

func TestGetRetryDelay(t *testing.T) {
	config := RetryConfig{0, 2, 32, nil}
	first := config.getRetryDelay(0)
	if !(first >= 1.0 && first <= 2.0) {
		t.Errorf("first: Not allowed delay value %f", first)
		return
	}
	second := config.getRetryDelay(1)
	if !(second >= 2.0 && second <= 4.0) {
		t.Errorf("second: Not allowed delay value %f", second)
		return
	}
	third := config.getRetryDelay(2)
	if !(third >= 4.0 && third <= 8.0) {
		t.Errorf("third: Not allowed delay value %f", third)
		return
	}

	upperLimit := config.getRetryDelay(4)
	if !(upperLimit >= 16.0 && upperLimit <= 32.0) {
		t.Errorf("upperLimit: Not allowed delay value %f", upperLimit)
		return
	}

	overLimit := config.getRetryDelay(10)
	if !(overLimit >= 16.0 && overLimit <= 32.0) {
		t.Errorf("overLimit: Not allowed delay value %f", overLimit)
		return
	}
}

func TestAttemptRequestWithoutRetrySetting(t *testing.T) {
	// リトライなし設定におけるリクエスト試行をテスト
	// RetryConfig.Logger によるログ記録がないことで処理完了を検証
	var body []byte
	statusCode := 200
	client, _ := NewMockClient(statusCode, body)
	var buf bytes.Buffer
	logger := log.New(&buf, "pre", log.Ldate)
	noRetry := RetryConfig{0, 2, 32, logger} // リトライなしであることを明示
	s := New("sk_test_xxxx", client, OptionRetryConfig(noRetry))
	req, _ := s.buildRequest(POST, "https://te.st/somewhere/endpoint", make(HeaderMap), newRequestBuilder())
	resp, _ := s.attemptRequest(req)
	if buf.Len() != 0 {
		t.Error("Unexpected logging fired")
	}
	if resp.StatusCode != statusCode {
		t.Error("Expected 429")
	}
}

var rateLimitResponseBody = []byte(`{
  "error": {
    "code": "over_capacity",
    "message": "The service is over capacity. Please try again later.",
    "status": 429,
    "type": "client_error"
  }
}`)

func TestAttemptRequestReachedRetryLimit(t *testing.T) {
	// リトライ回数上限に到達する場合のテスト
	client, transport := NewMockClient(rateLimitStatusCode, rateLimitResponseBody)
	var buf bytes.Buffer
	logger := log.New(&buf, "TestAttemptRequestReachedRateLimit_", log.Ldate)
	retryConfig := RetryConfig{2, 0.1, 10, logger}
	s := New("sk_test_xxxx", client, OptionRetryConfig(retryConfig))
	transport.AddResponse(rateLimitStatusCode, rateLimitResponseBody)
	req, _ := s.buildRequest(POST, "https://te.st/somewhere/endpoint", make(HeaderMap), newRequestBuilder())
	resp, err := s.attemptRequest(req)
	if err != nil {
		t.Error("Expected normal response")
		return
	}
	if resp.StatusCode != rateLimitStatusCode {
		t.Error("Expected reached the rate limit")
		return
	}
	logResults := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	if len(logResults) != retryConfig.MaxCount {
		t.Error("Expected logging retry statuses")
		return
	}
	for i := 0; i < len(logResults); i++ {
		if !strings.Contains(logResults[i], fmt.Sprintf("Retry Count: %d", i+1)) {
			t.Error("Failed to log retry count correctly")
			return
		}
	}
}

func TestAttemptRequestNotReachedRetryLimit(t *testing.T) {
	// リトライ回数上限到達前にリクエストを完了する場合のテスト
	client, transport := NewMockClient(rateLimitStatusCode, rateLimitResponseBody)
	var buf bytes.Buffer
	logger := log.New(&buf, "TestAttemptRequestNotReachedRetryLimit", log.Ldate)
	retryConfig := RetryConfig{3, 0.1, 10, logger}
	s := New("sk_test_xxx", client, OptionRetryConfig(retryConfig))
	notRateLimitStatusCode := 200
	transport.AddResponse(rateLimitStatusCode, rateLimitResponseBody)
	transport.AddResponse(notRateLimitStatusCode, []byte(``))
	req, _ := s.buildRequest(POST, "https://te.st/somewhere/endpoint", make(HeaderMap), newRequestBuilder())
	resp, err := s.attemptRequest(req)
	if err != nil {
		t.Error("Expected normal response")
		return
	}
	if resp.StatusCode != notRateLimitStatusCode {
		t.Errorf("Expected not rate limit status code(%d), but %d", notRateLimitStatusCode, resp.StatusCode)
		return
	}
	logResults := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	if len(logResults) != 2 {
		t.Error("Incorrect retry logging")
		return
	}
}
