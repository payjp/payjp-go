package payjp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Config struct {
	ApiBase string
}

type Service struct {
	Client  *http.Client
	apiKey  string
	apiBase string

	Customer     *customerService
	Plan         *PlanService
	Subscription *subscriptionService
	Account      *AccountService
	Token        *TokenService
}

func New(apiKey string, client *http.Client, config ...Config) *Service {
	if client == nil {
		client = &http.Client{}
	}
	service := &Service{
		apiKey: "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":")),
		Client: client,
	}
	if len(config) > 0 {
		service.apiBase = config[0].ApiBase
	} else {
		service.apiBase = "https://api.pay.jp/v1"
	}

	service.Customer = newCustomerService(service)
	service.Plan = newPlanService(service)
	service.Subscription = newSubscriptionService(service)
	service.Account = newAccountService(service)
	service.Token = newTokenService(service)

	return service
}

func (s Service) ApiBase() string {
	return s.apiBase
}

func (s Service) get(resourceUrl string) ([]byte, error) {
	request, err := http.NewRequest("GET", s.apiBase+resourceUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.apiKey)

	return respToBody(s.Client.Do(request))
}

func (s Service) delete(resourceUrl string) error {
	request, err := http.NewRequest("DELETE", s.apiBase+resourceUrl, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", s.apiKey)

	_, err = parseResponseError(s.Client.Do(request))
	return err
}

func (s Service) queryList(resourcePath string, limit, offset, since, until int) ([]byte, error) {
	if limit < 0 || limit > 100 {
		return nil, fmt.Errorf("List().Limit() should be between 1 and 100, but %d.", limit)
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
	var requestUrl string
	if hasParam {
		requestUrl = s.apiBase + resourcePath + "?" + values.Encode()
	} else {
		requestUrl = s.apiBase + resourcePath
	}
	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", s.apiKey)

	return respToBody(s.Client.Do(request))
}
