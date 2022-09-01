package payjp

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

const tagName = "form"

// Config 構造体はNewに渡すパラメータを設定するのに使用します。
type Config struct {
	APIBase string // APIのエンドポイントのURL(省略時は'https://api.pay.jp/v1')
}

// Service 構造体はPAY.JPのすべてのAPIの起点となる構造体です。
// New()を使ってインスタンスを生成します。
type Service struct {
	Client  *http.Client
	apiKey  string
	apiBase string

	Charge       *ChargeService       // 支払いに関するAPI
	Customer     *CustomerService     // 顧客情報に関するAPI
	Plan         *PlanService         // プランに関するAPI
	Subscription *SubscriptionService // 定期課金に関するAPI
	Token        *TokenService        // トークンに関するAPI
	Transfer     *TransferService     // 入金に関するAPI
	Event        *EventService        // イベント情報に関するAPI
	Account      *AccountService      // アカウント情報に関するAPI
}

// New はPAY.JPのAPIを初期化する関数です。
//
// apiKeyはPAY.JPのウェブサイトで作成したキーを指定します。
//
// clientは特別な設定をしたhttp.Clientを使用する場合に渡します。nilを指定するとデフォルトのもhttp.Clientを指定します。
//
// configは追加の設定が必要な場合に渡します。現状で設定できるのはAPIのエントリーポイントのURLのみです。省略できます。
func New(apiKey string, client *http.Client, config ...Config) *Service {
	if client == nil {
		client = &http.Client{}
	}
	service := &Service{
		apiKey: "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":")),
		Client: client,
	}
	if len(config) > 0 {
		service.apiBase = config[0].APIBase
	} else {
		service.apiBase = "https://api.pay.jp/v1"
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

func (s Service) request(method, path string, body io.Reader) ([]byte, error) {
	request, err := http.NewRequest(method, s.apiBase+path, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.apiKey)
	if method == "POST" && body != nil {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := s.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (s Service) retrieve(path string) ([]byte, error) {
	return s.request("GET", path, nil)
}

func (s Service) delete(path string) error {
	body, err := s.request("DELETE", path, nil)
	if err != nil {
		return err
	}
	_, err = parseError(body)
	return err
}

func (s Service) queryList(resourcePath string, limit, offset, since, until int, callbacks ...func(*url.Values) bool) ([]byte, error) {
	return s.queryListAll(resourcePath, limit, offset, since, until, 0, 0, callbacks...)
}

func (s Service) queryTransferList(resourcePath string, limit, offset, since, until, sinceSheduledDate, untilSheduledDate int, callbacks ...func(*url.Values) bool) ([]byte, error) {
	return s.queryListAll(resourcePath, limit, offset, since, until, sinceSheduledDate, untilSheduledDate, callbacks...)
}

func (s Service) queryListAll(resourcePath string, limit, offset, since, until, sinceSheduledDate, untilSheduledDate int, callbacks ...func(*url.Values) bool) ([]byte, error) {
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
	if sinceSheduledDate != 0 {
		values.Add("since_sheduled_date", strconv.Itoa(sinceSheduledDate))
		hasParam = true
	}
	if untilSheduledDate != 0 {
		values.Add("until_sheduled_date", strconv.Itoa(untilSheduledDate))
		hasParam = true
	}
	// add extra parameters
	for _, callback := range callbacks {
		if callback(&values) {
			hasParam = true
		}
	}
	requestURL := resourcePath
	if hasParam {
		requestURL = requestURL + "?" + values.Encode()
	}
	return s.retrieve(requestURL)
}

// 第二引数の構造体はListParamsを含む必要があり、かつメンバは全てポインタ型である必要があります
func (s Service) getList(path string, c interface{}) ([]byte, error) {
	values := url.Values{}
	hasParam := false
	rv := reflect.ValueOf(c)
	if c != nil && rv.Kind() == reflect.Ptr && !rv.IsNil() {
		v := rv.Elem()
		t := v.Type()

		for i := 0; i < t.NumField(); i++ {
			rf := t.Field(i)
			formName := rf.Tag.Get(tagName)

			if formName == "-" {
				continue
			}
			fieldV := v.Field(i)
			if fieldV.IsZero() {
				continue
			}

			fieldV = fieldV.Elem()
			var value string
			switch rf.Type.Elem().Kind() {
		    case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				value = strconv.FormatInt(fieldV.Int(), 10)
			default:
				value = fieldV.String()
			}
			values.Add(formName, value)
			hasParam = true
		}
	}
	if hasParam {
		return s.retrieve(path + "?" + values.Encode())
	}
	return s.retrieve(path)
}
