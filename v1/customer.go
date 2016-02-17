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
	Email       string
	Description string
	DefaultCard string
	ID          string
	CardToken   string
	Card        Card
}

func parseCustomer(service *Service, body []byte, result *CustomerResponse) (*CustomerResponse, error) {
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = service
	for _, card := range result.Cards.Data {
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

func (c customerService) UpdateCard(id, cardId string, card Card) (*CardResponse, error) {
	result := &CardResponse{
		customerID: id,
		service:    c.service,
	}
	return c.postCard(id, "/"+cardId, card, result)
}

func (c customerService) DeleteCard(id, cardId string) error {
	return c.service.delete("/customers/" + id + "/cards/" + cardId)
}

func (c customerService) ListCard(id string) *customerCardListCaller {
	return &customerCardListCaller{
		service:    c.service,
		customerId: id,
	}
}

func (c customerService) GetSubscription(customerId, cardId string) (*SubscriptionResponse, error) {
	body, err := c.service.get("/customers/" + customerId + "/subscriptions/" + cardId)
	if err != nil {
		return nil, err
	}
	result := &SubscriptionResponse{
		service: c.service,
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c customerService) ListSubscription(id string) *customerSubscriptionListCaller {
	return &customerSubscriptionListCaller{
		service:    c.service,
		customerId: id,
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
	result := &CustomerListResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, false, err
	}
	for _, customer := range result.Data {
		customer.service = c.service
	}
	return result.Data, result.HasMore, nil
}

type customerCardListCaller struct {
	service    *Service
	customerId string
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
	body, err := c.service.queryList("/customers/"+c.customerId+"/cards", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	result := &CardList{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, false, err
	}
	for _, customer := range result.Data {
		customer.service = c.service
	}
	return result.Data, result.HasMore, nil
}

type customerSubscriptionListCaller struct {
	service    *Service
	customerId string
	limit      int
	offset     int
	since      int
	until      int
}

func (c *customerSubscriptionListCaller) Limit(limit int) *customerSubscriptionListCaller {
	c.limit = limit
	return c
}

func (c *customerSubscriptionListCaller) Offset(offset int) *customerSubscriptionListCaller {
	c.offset = offset
	return c
}

func (c *customerSubscriptionListCaller) Since(since time.Time) *customerSubscriptionListCaller {
	c.since = int(since.Unix())
	return c
}

func (c *customerSubscriptionListCaller) Until(until time.Time) *customerSubscriptionListCaller {
	c.until = int(until.Unix())
	return c
}

func (c *customerSubscriptionListCaller) Do() ([]*SubscriptionResponse, bool, error) {
	body, err := c.service.queryList("/customers/"+c.customerId+"/subscriptions", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	result := &SubscriptionList{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, false, err
	}
	for _, customer := range result.Data {
		customer.service = c.service
	}
	return result.Data, result.HasMore, nil
}

type CustomerResponse struct {
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

func (c *CustomerResponse) ListSubscription() *customerSubscriptionListCaller {
	return c.service.Customer.ListSubscription(c.ID)
}

type customer CustomerResponse

func (c *CustomerResponse) UnmarshalJSON(b []byte) error {
	raw := customer{}
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "customer" {
		*c = CustomerResponse(raw)
		c.CreatedAt = time.Unix(int64(raw.CreatedEpoch), 0)
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
	Count   int                 `json:"count"`
	Data    []*CustomerResponse `json:"data"`
	HasMore bool                `json:"has_more"`
	Object  string              `json:"object"`
	URL     string              `json:"url"`
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
