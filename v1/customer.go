package payjp

import (
	"encoding/json"
	"net/http"
	"time"
)

type customerService struct {
	service *Service
}

func newCustomerService(service *Service) *customerService {
	return &customerService{
		service: service,
	}
}

type Customer struct {
	Email       interface{}
	Description interface{}
	DefaultCard interface{}
	ID          interface{}
	CardToken   interface{}
	Card        Card
}

func parseCustomer(service *Service, body []byte, result *CustomerResponse) (*CustomerResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	for _, card := range result.Cards {
		card.service = service
		card.customerID = result.ID
	}
	return result, nil
}

func (c customerService) Create(customer Customer) (*CustomerResponse, error) {
	qb := newRequestBuilder()
	if customer.Email != "" {
		qb.Add("email", customer.Email)
	}
	if customer.Description != "" {
		qb.Add("description", customer.Description)
	}
	if customer.DefaultCard != "" {
		qb.Add("default_card", customer.DefaultCard)
	}
	if customer.ID != "" {
		qb.Add("id", customer.ID)
	}
	if customer.CardToken != "" {
		qb.Add("card", customer.CardToken)
	} else {
		qb.AddCard(customer.Card)
	}

	request, err := http.NewRequest("POST", c.service.apiBase+"/customers", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	body, err := respToBody(c.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseCustomer(c.service, body, &CustomerResponse{})
}

func (c customerService) Get(id string) (*CustomerResponse, error) {
	body, err := c.service.get("/customers/" + id)
	if err != nil {
		return nil, err
	}
	return parseCustomer(c.service, body, &CustomerResponse{})
}

func (c customerService) Update(id string, customer Customer) (*CustomerResponse, error) {
	body, err := c.update(id, customer)
	if err != nil {
		return nil, err
	}
	return parseCustomer(c.service, body, &CustomerResponse{})
}

func (c customerService) update(id string, customer Customer) ([]byte, error) {
	qb := newRequestBuilder()
	if customer.Email != "" {
		qb.Add("email", customer.Email)
	}
	if customer.Description != "" {
		qb.Add("description", customer.Description)
	}
	if customer.DefaultCard != "" {
		qb.Add("default_card", customer.DefaultCard)
	}
	if customer.CardToken != "" {
		qb.Add("card", customer.CardToken)
	} else {
		qb.AddCard(customer.Card)
	}
	request, err := http.NewRequest("POST", c.service.apiBase+"/customers/"+id, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

func (c customerService) Delete(id string) error {
	return c.service.delete("/customers/" + id)
}

func (c customerService) List() *customerListCaller {
	return &customerListCaller{
		service: c.service,
	}
}

func (c customerService) AddCardToken(customerID, token string) (*CardResponse, error) {
	qb := newRequestBuilder()
	qb.Add("card", token)

	request, err := http.NewRequest("POST", c.service.apiBase+"/customers/"+customerID+"/cards", qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	body, err := respToBody(c.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseCard(c.service, body, &CardResponse{}, customerID)
}

func (c customerService) postCard(customerID, resourcePath string, card Card, result *CardResponse) (*CardResponse, error) {
	qb := newRequestBuilder()
	qb.AddCard(card)

	request, err := http.NewRequest("POST", c.service.apiBase+"/customers/"+customerID+"/cards"+resourcePath, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	body, err := respToBody(c.service.Client.Do(request))
	if err != nil {
		return nil, err
	}
	return parseCard(c.service, body, result, customerID)
}

func (c customerService) AddCard(id string, card Card) (*CardResponse, error) {
	return c.postCard(id, "", card, &CardResponse{})
}

func (c customerService) GetCard(customerID, cardID string) (*CardResponse, error) {
	body, err := c.service.get("/customers/" + customerID + "/cards/" + cardID)
	if err != nil {
		return nil, err
	}
	return parseCard(c.service, body, &CardResponse{}, customerID)
}

func (c customerService) UpdateCard(iD, cardID string, card Card) (*CardResponse, error) {
	result := &CardResponse{
		customerID: iD,
		service:    c.service,
	}
	return c.postCard(iD, "/"+cardID, card, result)
}

func (c customerService) DeleteCard(ID, cardID string) error {
	return c.service.delete("/customers/" + ID + "/cards/" + cardID)
}

func (c customerService) ListCard(id string) *customerCardListCaller {
	return &customerCardListCaller{
		service:    c.service,
		customerID: id,
	}
}

func (c customerService) GetSubscription(customerID, subscriptionID string) (*SubscriptionResponse, error) {
	return c.service.Subscription.Get(customerID, subscriptionID)
}

func (c customerService) ListSubscription(ID string) *subscriptionListCaller {
	return &subscriptionListCaller{
		service:    c.service,
		customerID: ID,
	}
}

type customerListCaller struct {
	service *Service
	limit   int
	offset  int
	since   int
	until   int
}

func (c *customerListCaller) Limit(limit int) *customerListCaller {
	c.limit = limit
	return c
}

func (c *customerListCaller) Offset(offset int) *customerListCaller {
	c.offset = offset
	return c
}

func (c *customerListCaller) Since(since time.Time) *customerListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *customerListCaller) Until(until time.Time) *customerListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *customerListCaller) Do() ([]*CustomerResponse, bool, error) {
	body, err := c.service.queryList("/customers", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	temp := &CustomerListResponse{}
	err = json.Unmarshal(body, temp)
	if err != nil {
		return nil, false, err
	}
	result := make([]*CustomerResponse, len(temp.Data))

	for i, raw := range temp.Data {
		customer := &CustomerResponse{}
		json.Unmarshal(raw, customer)
		customer.service = c.service
		result[i] = customer
	}
	return result, temp.HasMore, nil
}

type customerCardListCaller struct {
	service    *Service
	customerID string
	limit      int
	offset     int
	since      int
	until      int
}

func (c *customerCardListCaller) Limit(limit int) *customerCardListCaller {
	c.limit = limit
	return c
}

func (c *customerCardListCaller) Offset(offset int) *customerCardListCaller {
	c.offset = offset
	return c
}

func (c *customerCardListCaller) Since(since time.Time) *customerCardListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *customerCardListCaller) Until(until time.Time) *customerCardListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *customerCardListCaller) Do() ([]*CardResponse, bool, error) {
	body, err := c.service.queryList("/customers/"+c.customerID+"/cards", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	rawResult := &CardList{}
	err = json.Unmarshal(body, rawResult)
	if err != nil {
		return nil, false, err
	}
	result := make([]*CardResponse, len(rawResult.Data))
	for i, rawCustomer := range rawResult.Data {
		card := &CardResponse{}
		json.Unmarshal(rawCustomer, card)
		result[i] = card
	}
	return result, rawResult.HasMore, nil
}

type CustomerResponse struct {
	Cards         []*CardResponse
	CreatedAt     time.Time
	DefaultCard   string
	Description   string
	Email         string
	ID            string
	LiveMode      bool
	Subscriptions []*SubscriptionResponse

	service *Service
}

type customerResponseParser struct {
	Cards         CardList         `json:"cards"`
	CreatedEpoch  int              `json:"created"`
	DefaultCard   string           `json:"default_card"`
	Description   string           `json:"description"`
	Email         string           `json:"email"`
	ID            string           `json:"id"`
	LiveMode      bool             `json:"livemode"`
	Object        string           `json:"object"`
	Subscriptions SubscriptionList `json:"subscriptions"`

	CreatedAt time.Time

	service *Service
}

func (c *CustomerResponse) Update(customer Customer) error {
	body, err := c.service.Customer.update(c.ID, customer)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, c)
}

func (c *CustomerResponse) Delete() error {
	return c.service.Customer.Delete(c.ID)
}

func (c *CustomerResponse) AddCard(card Card) (*CardResponse, error) {
	return c.service.Customer.AddCard(c.ID, card)
}

func (c *CustomerResponse) AddCardToken(token string) (*CardResponse, error) {
	return c.service.Customer.AddCardToken(c.ID, token)
}

func (c *CustomerResponse) GetCard(cardId string) (*CardResponse, error) {
	return c.service.Customer.GetCard(c.ID, cardId)
}

func (c *CustomerResponse) ListCard() *customerCardListCaller {
	return c.service.Customer.ListCard(c.ID)
}

func (c *CustomerResponse) GetSubscription(subscriptionID string) (*SubscriptionResponse, error) {
	return c.service.Customer.GetSubscription(c.ID, subscriptionID)
}

func (c *CustomerResponse) ListSubscription() *subscriptionListCaller {
	return c.service.Customer.ListSubscription(c.ID)
}

func (c *CustomerResponse) UnmarshalJSON(b []byte) error {
	raw := customerResponseParser{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "customer" {
		c.Cards = make([]*CardResponse, len(raw.Cards.Data))
		for i, rawCard := range raw.Cards.Data {
			card := &CardResponse{}
			json.Unmarshal(rawCard, card)
			c.Cards[i] = card
		}
		c.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
		c.DefaultCard = raw.DefaultCard
		c.Description = raw.Description
		c.Email = raw.Email
		c.ID = raw.ID
		c.LiveMode = raw.LiveMode
		c.Subscriptions = make([]*SubscriptionResponse, len(raw.Subscriptions.Data))
		for i, rawSubscription := range raw.Subscriptions.Data {
			subscription := &SubscriptionResponse{}
			json.Unmarshal(rawSubscription, subscription)
			c.Subscriptions[i] = subscription
		}
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}

type CustomerListResponse struct {
	Count   int               `json:"count"`
	Data    []json.RawMessage `json:"data"`
	HasMore bool              `json:"has_more"`
	Object  string            `json:"object"`
	URL     string            `json:"url"`
}

type customerList CustomerListResponse

func (p *CustomerListResponse) UnmarshalJSON(b []byte) error {
	raw := customerList{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "list" {
		*p = CustomerListResponse(raw)
		return nil
	}
	rawError := ErrorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
