package payjp

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type HeaderMap = map[string]string
type RetryConfig struct {
	MaxCount     int
	InitialDelay float64     // sec
	MaxDelay     float64     // sec
	Logger       *log.Logger // nilable
}

func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		0,
		2,
		32,
		nil,
	}
}

// リクエストリトライ時に遅延させる時間を計算する
// equal jitter に基づいて算出
// ref: https://aws.amazon.com/jp/blogs/architecture/exponential-backoff-and-jitter/
func (r RetryConfig) getRetryDelay(retryCount int) float64 {
	delay := math.Min(r.MaxDelay, r.InitialDelay*math.Pow(2.0, float64(retryCount)))
	half := delay / 2.0
	offset := RandUniform(0, half)
	return half + offset
}

// Service 構造体はPAY.JPのすべてのAPIの起点となる構造体です。
// New()を使ってインスタンスを生成します。
type Service struct {
	Client      *http.Client
	apiKey      string
	apiBase     string
	retryConfig RetryConfig

	Charge       *ChargeService       // 支払いに関するAPI
	Customer     *CustomerService     // 顧客情報に関するAPI
	Plan         *PlanService         // プランに関するAPI
	Subscription *SubscriptionService // 定期課金に関するAPI
	Token        *TokenService        // トークンに関するAPI
	Transfer     *TransferService     // 入金に関するAPI
	Event        *EventService        // イベント情報に関するAPI
	Account      *AccountService      // アカウント情報に関するAPI
}

type Option func(*Service)

func OptionApiBase(url string) Option {
	return func(s *Service) {
		s.apiBase = url
	}
}

func OptionRetryConfig(retryConfig RetryConfig) Option {
	return func(s *Service) {
		s.retryConfig = retryConfig
	}
}

// New はPAY.JPのAPIを初期化する関数です。
//
// apiKeyはPAY.JPのウェブサイトで作成したキーを指定します。
//
// clientは特別な設定をしたhttp.Clientを使用する場合に渡します。nilを指定するとデフォルトのもhttp.Clientを指定します。
//
// configは追加の設定が必要な場合に渡します。現状で設定できるのはAPIのエントリーポイントのURLのみです。省略できます。
func New(apiKey string, client *http.Client, options ...Option) *Service {
	if client == nil {
		client = &http.Client{}
	}
	service := &Service{
		apiKey: "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":")),
		Client: client,
	}

	service.apiBase = "https://api.pay.jp/v1"
	service.retryConfig = defaultRetryConfig()
	for _, o := range options {
		o(service)
	}

	service.Charge = newChargeService(service)
	service.Customer = newCustomerService(service)
	service.Plan = newPlanService(service)
	service.Subscription = newSubscriptionService(service)
	service.Account = newAccountService(service)
	service.Token = newTokenService(service)
	service.Transfer = newTransferService(service)
	service.Event = newEventService(service)

	return service
}

// APIBase はPAY.JPのエントリーポイントの基底部分のURLを返します。
func (s Service) APIBase() string {
	return s.apiBase
}

func (s Service) RetryConfig() RetryConfig {
	return s.retryConfig
}

type HttpMethod int

const (
	GET HttpMethod = iota + 1
	POST
	PUT
	DELETE
)

type UnkownHttpMethod struct{}

func (m HttpMethod) String() string {
	switch m {
	case 1:
		return "GET"
	case 2:
		return "POST"
	case 3:
		return "PUT"
	case 4:
		return "DELETE"
	default:
		// though never reach
		panic(UnkownHttpMethod{})
	}
}

func (s Service) buildRequest(method HttpMethod, url string, headers HeaderMap, requestBuilder *requestBuilder) (*http.Request, error) {
	var payload io.Reader = nil
	if requestBuilder != nil {
		payload = requestBuilder.Reader()
	}
	req, err := http.NewRequest(method.String(), url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", s.apiKey)
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
	return req, nil
}

func (s Service) doRequest(request *http.Request) (*http.Response, error) {
	res, err := s.Client.Do(request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

var rateLimitStatusCode = 429

func (s Service) attemptRequest(request *http.Request) (res *http.Response, err error) {
	logger := s.retryConfig.Logger
	// レートリミット時、必要に応じてリトライを試行するリクエストのラッパー
	res, err = s.doRequest(request)
	for currentRetryCount := 0; currentRetryCount < s.retryConfig.MaxCount; currentRetryCount++ {
		if res.StatusCode != rateLimitStatusCode {
			// レートリミット制限ではないのでこれ以上のリトライは不要
			break
		}
		delay := s.retryConfig.getRetryDelay(currentRetryCount)
		delaySec := time.Duration(delay*1000) * time.Millisecond
		if logger != nil {
			logger.Printf("Current Retry Count: %d. Retry after %v seconds", currentRetryCount+1, delaySec)
		}
		time.Sleep(delaySec)
		res, err = s.doRequest(request)
	}
	return res, err
}

func (s Service) request(
	method HttpMethod,
	url string,
	headers HeaderMap,
	requestBuilder *requestBuilder,
) (*http.Response, error) {
	// レスポンスのデコードを含めたHTTPリクエストを行う
	req, err := s.buildRequest(method, url, headers, requestBuilder)
	if err != nil {
		return nil, err
	}
	return s.attemptRequest(req)
}

func (s Service) postRequest(url string, headers HeaderMap, requestBuilder *requestBuilder) (*http.Response, error) {
	return s.request(POST, url, headers, requestBuilder)
}

func (s Service) getRequest(url string, headers HeaderMap, requestBuilder *requestBuilder) (*http.Response, error) {
	return s.request(GET, url, headers, requestBuilder)
}

func (s Service) deleteRequest(url string, headers HeaderMap, requestBuilder *requestBuilder) (*http.Response, error) {
	return s.request(DELETE, url, headers, requestBuilder)
}

func (s Service) retrieve(resourceURL string) ([]byte, error) {
	return respToBody(s.getRequest(s.apiBase+resourceURL, make(HeaderMap), nil))
}

func (s Service) delete(resourceURL string) error {
	_, err := parseResponseError(s.deleteRequest(s.apiBase+resourceURL, make(HeaderMap), nil))
	return err
}

func (s Service) queryList(resourcePath string, limit, offset, since, until int, callbacks ...func(*url.Values) bool) ([]byte, error) {
	if limit < 0 || limit > 100 {
		return nil, fmt.Errorf("method Limit() should be between 1 and 100, but %d", limit)
	}

	values := url.Values{}
	hasParam := false
	if limit != 0 {
		values.Add("limit", strconv.Itoa(limit))
		hasParam = true
	}
	if offset != 0 {
		values.Add("offset", strconv.Itoa(offset))
		hasParam = true
	}
	if since != 0 {
		values.Add("since", strconv.Itoa(since))
		hasParam = true
	}
	if until != 0 {
		values.Add("until", strconv.Itoa(until))
		hasParam = true
	}
	// add extra parameters
	for _, callback := range callbacks {
		if callback(&values) {
			hasParam = true
		}
	}
	var requestURL string
	if hasParam {
		requestURL = s.apiBase + resourcePath + "?" + values.Encode()
	} else {
		requestURL = s.apiBase + resourcePath
	}
	return respToBody(s.getRequest(requestURL, make(HeaderMap), nil))
}
