package payjp

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

type TestListParams struct {
	Param *string `form:"param"`
}

var errorJSONStr = `
{
  "code": "code",
  "message": "message",
  "param": "param",
  "status": 400,
  "type": "type"
}
`
var rateLimitResponseBody = []byte(`{
  "error": {
    "code": "over_capacity",
    "message": "The service is over capacity. Please try again later.",
    "status": 429,
    "type": "client_error"
  }
}`)
var rateLimitStr = "429: Type: client_error Code: over_capacity Message: The service is over capacity. Please try again later."

var errorJSON = []byte(errorJSONStr)

func TestNew(t *testing.T) {
	// New(): default constructor
	service := New("api-key", nil)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Client)

	assert.NotNil(t, service.Charge)
	assert.NotNil(t, service.Customer)
	assert.NotNil(t, service.Plan)
	assert.NotNil(t, service.Subscription)
	assert.NotNil(t, service.Account)
	assert.NotNil(t, service.Token)
	assert.NotNil(t, service.Transfer)
	assert.NotNil(t, service.Event)
	assert.Equal(t, "https://api.pay.jp/v1", service.apiBase)
	assert.Regexp(t, "^Basic .*", service.apiKey)
	assert.EqualValues(t, 0, service.MaxCount)
	assert.EqualValues(t, 2, service.InitialDelay)
	assert.EqualValues(t, 32, service.MaxDelay)
	assert.EqualValues(t, service.Logger, DefaultLogger)

	client := &http.Client{}
	assert.NotSame(t, client, service.Client)

	service = New("api-key", client)
	assert.Same(t, client, service.Client)

	service = New("api-key", nil, WithAPIBase("https://api.pay.jp/v2"))
	assert.Equal(t, "https://api.pay.jp/v2", service.APIBase())
}

func TestRequests(t *testing.T) {
	mock, transport := newMockClient(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	service := New("api-key", mock)
	qb := newRequestBuilder()

	body, err := service.request("POST", "/test", qb.Reader())
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/test", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, "Basic YXBpLWtleTo=", transport.Header.Get("Authorization"))
	assert.Regexp(t, "^Go-http-client/payjp-v[0-9.]+$", transport.Header.Get("User-Agent"))
	assert.Regexp(t, "^payjp-go/v[0-9.]+\\(go[0-9.]+,os\\:.+,arch\\:.+\\)$", transport.Header.Get("X-Payjp-Client-User-Agent"))
	assert.Equal(t, "application/x-www-form-urlencoded", transport.Header.Get("Content-Type"))
	assert.Equal(t, errorJSONStr, string(body))
}

func TestGetList(t *testing.T) {
	mock, transport := newMockClient(400, errorJSON)
	transport.AddResponse(400, errorJSON)
	service := New("api-key", mock)

	l := &TestListParams{}
	q := service.getQuery(l)
	assert.Equal(t, "", q)

	str := "str"
	l2 := &TestListParams{
		Param: &str,
	}
	q = service.getQuery(l2)
	assert.Equal(t, "?param=str", q)
}

func TestGetRetryDelay(t *testing.T) {
	config := New("sk_test_xxxx", nil)
	first := config.getRetryDelay(0)
	assert.GreaterOrEqual(t, first, 1.0)
	assert.LessOrEqual(t, first, 2.0)
	second := config.getRetryDelay(1)
	assert.GreaterOrEqual(t, second, 2.0)
	assert.LessOrEqual(t, second, 4.0)
	third := config.getRetryDelay(2)
	assert.GreaterOrEqual(t, third, 4.0)
	assert.LessOrEqual(t, third, 8.0)
	upperLimit := config.getRetryDelay(4)
	assert.GreaterOrEqual(t, upperLimit, 16.0)
	assert.LessOrEqual(t, upperLimit, 32.0)
	overLimit := config.getRetryDelay(10)
	assert.GreaterOrEqual(t, overLimit, 16.0)
	assert.LessOrEqual(t, overLimit, 32.0)
}

func TestAttemptRequestWithoutRetrySetting(t *testing.T) {
	// リトライなし設定におけるリクエスト試行をテスト
	// RetryConfig.Logger によるログ記録がないことで処理完了を検証
	client, transport := newMockClient(rateLimitStatusCode, rateLimitResponseBody)
	var buf bytes.Buffer
	logger := &PayjpLogger{logLevel: LogLevelDebug, stdoutOverride: &buf}
	var noRetry = []serviceConfig{
		WithMaxCount(0),
		WithInitialDelay(2),
		WithMaxDelay(32),
		WithLogger(logger),
	} // リトライなしであることを明示
	s := New("sk_test_xxxx", client, noRetry...)
	resp, err := s.request("POST", "/endpoint", newRequestBuilder().Reader())
	assert.NoError(t, err)
	assert.Equal(t, "https://api.pay.jp/v1/endpoint", transport.URL)
	assert.Equal(t, "POST", transport.Method)
	assert.Equal(t, rateLimitResponseBody, resp)

	result := buf.String()
	assert.Equal(t, result, "")
	assert.Equal(t, 0, buf.Len())
}

func TestAttemptRequestReachedRetryLimit(t *testing.T) {
	// リトライ回数上限に到達する場合のテスト
	client, transport := newMockClient(rateLimitStatusCode, rateLimitResponseBody)
	transport.AddResponse(rateLimitStatusCode, rateLimitResponseBody)
	var buf bytes.Buffer
	logger := &PayjpLogger{logLevel: LogLevelDebug, stdoutOverride: &buf}
	s := New("sk_test_xxxx", client,
		WithMaxCount(2),
		WithInitialDelay(0.1),
		WithMaxDelay(10),
		WithLogger(logger),
	)
	_, err := s.Account.Retrieve()
	assert.Equal(t, rateLimitStr, err.Error())
	logResults := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	assert.Equal(t, len(logResults), s.MaxCount)
	for i := 0; i < len(logResults); i++ {
		assert.True(t, strings.Contains(logResults[i], fmt.Sprintf("Retry Count: %d", i+1)))
	}
}

func TestAttemptRequestNotReachedRetryLimit(t *testing.T) {
	// リトライ回数上限到達前にリクエストを完了する場合のテスト
	client, transport := newMockClient(rateLimitStatusCode, rateLimitResponseBody)
	notRateLimitStatusCode := 200
	transport.AddResponse(rateLimitStatusCode, rateLimitResponseBody)
	transport.AddResponse(notRateLimitStatusCode, accountResponseJSON)
	var buf bytes.Buffer
	logger := &PayjpLogger{logLevel: LogLevelDebug, stdoutOverride: &buf}
	s := New("sk_test_xxxx", client,
		WithMaxCount(3),
		WithInitialDelay(0.1),
		WithMaxDelay(10),
		WithLogger(logger),
	)
	_, err := s.Account.Retrieve()
	assert.NoError(t, err)
	logResults := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	assert.Equal(t, len(logResults), 2)
	for i := 0; i < len(logResults); i++ {
		assert.True(t, strings.Contains(logResults[i], fmt.Sprintf("Retry Count: %d", i+1)))
	}
}

// test for RandUniform
type randUniformArgSource struct {
	valueX float64
	valueY float64
}

func (s randUniformArgSource) maxValue() float64 {
	return math.Max(s.valueX, s.valueY)
}

func (s randUniformArgSource) minValue() float64 {
	return math.Min(s.valueX, s.valueY)
}

func (randUniformArgSource) Generate(r *rand.Rand, size int) reflect.Value {
	s := randUniformArgSource{
		float64(r.Int63()),
		float64(r.Int63()),
	}
	return reflect.ValueOf(s)
}

func TestRandUniform(t *testing.T) {
	testFunc := func(arg randUniformArgSource) bool {
		min := arg.minValue()
		max := arg.maxValue()
		result := RandUniform(min, max)
		return result >= min || result < max
	}
	err := quick.Check(testFunc, nil)
	assert.NoError(t, err)
}
