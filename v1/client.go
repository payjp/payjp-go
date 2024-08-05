package payjp

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

const Version = "v0.2.4"
const tagName = "form"
const rateLimitStatusCode = 429

type serviceConfig func(*Service)

// WithAPIBase はAPIのエントリーポイントを変更するために使用します。
func WithAPIBase(apiBase string) serviceConfig {
	return func(s *Service) {
		s.apiBase = apiBase
	}
}

// WithMaxCount はリクエストのリトライ回数を変更するために使用します。
func WithMaxCount(maxCount int) serviceConfig {
	return func(s *Service) {
		s.MaxCount = maxCount
	}
}

// WithInitialDelay はリクエストリトライ時の初期遅延時間を変更するために使用します。
func WithInitialDelay(initialDelay float64) serviceConfig {
	return func(s *Service) {
		s.InitialDelay = initialDelay
	}
}

// WithMaxDelay はリクエストリトライ時の最大遅延時間を変更するために使用します。
func WithMaxDelay(maxDelay float64) serviceConfig {
	return func(s *Service) {
		s.MaxDelay = maxDelay
	}
}

// WithLogger はログ出力を変更するために使用します。
//
// デフォルトではログは出力されません。
// LoggerInterfaceを実装した構造体を渡すことでログ出力を変更できます。
func WithLogger(logger LoggerInterface) serviceConfig {
	return func(s *Service) {
		s.Logger = logger
	}
}

const defaultMaxCount = 0
const defaultInitialDelay = 2
const defaultMaxDelay = 32

func RandUniform(min, max float64) float64 {
	// [min, max) の小数を返す
	return (rand.Float64() * (max - min)) + min
}

// リクエストリトライ時に遅延させる時間を計算する
// equal jitter に基づいて算出
// ref: https://aws.amazon.com/jp/blogs/architecture/exponential-backoff-and-jitter/
func (r Service) getRetryDelay(retryCount int) float64 {
	delay := math.Min(r.MaxDelay, r.InitialDelay*math.Pow(2.0, float64(retryCount)))
	half := delay / 2.0
	offset := RandUniform(0, half)
	return half + offset
}

// Service 構造体はPAY.JPのすべてのAPIの起点となる構造体です。
// New()を使ってインスタンスを生成します。
type Service struct {
	Client       *http.Client
	apiKey       string
	apiBase      string
	MaxCount     int
	InitialDelay float64
	MaxDelay     float64
	Logger       LoggerInterface

	Charge       *ChargeService
	Customer     *CustomerService
	Plan         *PlanService
	Subscription *SubscriptionService
	Token        *TokenService
	Transfer     *TransferService
	Statement    *StatementService
	Term         *TermService
	Balance      *BalanceService
	Event        *EventService
	Account      *AccountService
}

// New はPAY.JPのAPIを初期化する関数です。
//
// apiKeyはPAY.JPのウェブサイトで作成したキーを指定します。
//
// clientは特別な設定をしたhttp.Clientを使用する場合に渡します。nilを指定するとデフォルトのもhttp.Clientを指定します。
//
// configは追加の設定が必要な場合に渡します。
func New(apiKey string, client *http.Client, configs ...serviceConfig) *Service {
	if client == nil {
		client = &http.Client{}
	}
	service := &Service{
		apiKey: "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":")),
		Client: client,
	}
	service.apiBase = "https://api.pay.jp/v1"
	service.MaxCount = defaultMaxCount
	service.InitialDelay = defaultInitialDelay
	service.MaxDelay = defaultMaxDelay
	service.Logger = NullLogger

	for _, c := range configs {
		c(service)
	}

	service.Charge = newChargeService(service)
	service.Customer = newCustomerService(service)
	service.Plan = newPlanService(service)
	service.Subscription = newSubscriptionService(service)
	service.Account = newAccountService(service)
	service.Token = newTokenService(service)
	service.Transfer = newTransferService(service)
	service.Statement = newStatementService(service)
	service.Balance = newBalanceService(service)
	service.Term = newTermService(service)
	service.Event = newEventService(service)

	return service
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// StringValue returns the value of the string pointer passed in or
// "" if the pointer is nil.
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// Int returns a pointer to the int64 value passed in.
func Int(v int) *int {
	return &v
}

// IntValue returns the value of the int64 pointer passed in or
// 0 if the pointer is nil.
func IntValue(v *int) int64 {
	if v != nil {
		return int64(*v)
	}
	return 0
}

// Float returns a pointer to the float64 value passed in.
func Float(v float64) *float64 {
	return &v
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// APIBase はPAY.JPのエントリーポイントの基底部分のURLを返します。
func (s Service) APIBase() string {
	return s.apiBase
}

func (s Service) request(method, path string, body io.Reader, headerMap ...map[string]string) ([]byte, error) {
	request, err := http.NewRequest(method, s.apiBase+path, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.apiKey)
	request.Header.Add("User-Agent", "Go-http-client/payjp-"+Version)
	request.Header.Add("X-Payjp-Client-User-Agent", "payjp-go/"+Version+"("+runtime.Version()+",os:"+runtime.GOOS+",arch:"+runtime.GOARCH+")")
	if method == "POST" && body != nil {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, headers := range headerMap {
		for k, v := range headers {
			request.Header.Add(k, v)
		}
	}

	resp, err := s.Client.Do(request)
	if err != nil {
		return nil, err
	}
	logger := s.Logger
	for currentRetryCount := 0; currentRetryCount < s.MaxCount; currentRetryCount++ {
		if resp.StatusCode != rateLimitStatusCode {
			// レートリミット制限ではないのでこれ以上のリトライは不要
			break
		}
		delay := time.Duration(s.getRetryDelay(currentRetryCount)*1000) * time.Millisecond
		if logger != nil {
			logger.Infof("Current Retry Count: %d. Retry after %v", currentRetryCount+1, delay)
		}
		time.Sleep(delay)
		resp, err = s.Client.Do(request)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (s Service) delete(path string) error {
	body, _ := s.request("DELETE", path, nil)
	var res DeleteResponse
	err := json.Unmarshal(body, &res)
	if err == nil && res.Deleted {
		return nil
	}
	return parseError(body)
}

type ListParams struct {
	Limit  *int `form:"limit"`
	Offset *int `form:"offset"`
	Since  *int `form:"since"`
	Until  *int `form:"until"`
}

// 第一引数の構造体を第二引数のURLパラメータにパースします(keyはメタ情報の値、valueは再帰しつつプリミティブに変換)
func (s Service) makeEncoder(v reflect.Value, values url.Values) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		rf := t.Field(i)
		formName := rf.Tag.Get(tagName)
		if formName == "" || formName == "-" {
			continue
		}
		fieldV := v.Field(i)
		if fieldV.Kind() != reflect.Ptr || fieldV.IsNil() { // todo golang >= 1.13 wanna use fieldV.IsZero()
			if fieldV.Kind() == reflect.Struct {
				s.makeEncoder(fieldV, values)
			}
			continue
		}
		fieldV = fieldV.Elem()
		var value string
		switch rf.Type.Elem().Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			value = strconv.FormatInt(fieldV.Int(), 10)
		case reflect.Bool:
			value = strconv.FormatBool(fieldV.Bool())
		default:
			value = fieldV.String()
		}
		values.Add(formName, value)
	}
}

// 引数の構造体からURLパラメータを生成します(メンバが全てポインタ型でメタ情報を持っている必要があります)
func (s Service) getQuery(c interface{}) string {
	rv := reflect.ValueOf(c)
	if c != nil && rv.Kind() == reflect.Ptr && !rv.IsNil() {
		v := rv.Elem()
		values := url.Values{}
		s.makeEncoder(v, values)
		query := values.Encode()
		if query != "" {
			return "?" + query
		}
	}
	return ""
}
