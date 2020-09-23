package payjp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type RetryConfig struct {
	MaxCount     int
	InitialDelay int
	MaxDelay     int
}

func defaultRetryConfig() RetryConfig {
	return RetryConfig{0, 2, 32}
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

func (s Service) buildRequest(method HttpMethod, url string, requestBuilder *requestBuilder) (*http.Request, error) {
	req, err := http.NewRequest(method.String(), url, requestBuilder.Reader())
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", s.apiKey)
	return req, nil
}

func (s Service) retrieve(resourceURL string) ([]byte, error) {
	request, err := http.NewRequest("GET", s.apiBase+resourceURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.apiKey)

	return respToBody(s.Client.Do(request))
}

func (s Service) post(url string, requestBuilder requestBuilder) ([]byte, error) {
	request, err := http.NewRequest("POST", url, requestBuilder.Reader())
	if err != nil {
		return nil, err
	}
	return respToBody(s.Client.Do(request))
}

func (s Service) delete(resourceURL string) error {
	request, err := http.NewRequest("DELETE", s.apiBase+resourceURL, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.apiKey)

	_, err = parseResponseError(s.Client.Do(request))
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
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.apiKey)

	return respToBody(s.Client.Do(request))
}
