package payjp

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
)

const Version = "v0.1.0"
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

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
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
	request.Header.Add("User-Agent", "Go-http-client/payjp-"+Version)
	request.Header.Add("X-Payjp-Client-User-Agent", "payjp-go/"+Version+"("+runtime.Version()+",os:"+runtime.GOOS+",arch:"+runtime.GOARCH+")")
	if method == "POST" && body != nil {
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		if path == "/tokens" {
			request.Header.Add("X-Payjp-Direct-Token-Generate", "true")
		}
	}

	resp, err := s.Client.Do(request)
	if err != nil {
		return nil, err
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
