package payjp

import (
	"encoding/json"
	"time"
)

// CustomerService は顧客を管理する機能を提供します。
//
// 顧客における都度の支払いや定期購入、複数カードの管理など、さまざまなことができます。
// 作成した顧客は、あとからカードを追加・更新・削除したり、顧客自体を削除することができます。
type CustomerService struct {
	service *Service
}

func newCustomerService(service *Service) *CustomerService {
	return &CustomerService{
		service: service,
	}
}

// Customer は顧客の登録や更新時に使用する構造体です
type Customer struct {
	Email       interface{}       // メールアドレス
	Description interface{}       // 概要
	ID          interface{}       // 一意の顧客ID
	CardToken   interface{}       // トークンID
	DefaultCard interface{}       // デフォルトカード
	Metadata    map[string]string // メタデータ
}

// Create はメールアドレスやIDなどを指定して顧客を作成します。
// https://pay.jp/docs/api/?go#顧客を作成
func (c CustomerService) Create(customer Customer) (*CustomerResponse, error) {
	qb := newRequestBuilder()
	qb.Add("email", customer.Email)
	qb.Add("description", customer.Description)
	qb.Add("id", customer.ID)
	qb.Add("card", customer.CardToken)
	qb.AddMetadata(customer.Metadata)

	body, err := c.service.request("POST", "/customers", qb.Reader())
	if err != nil {
		return nil, err
	}
	result := &CustomerResponse{service: c.service}
	err = json.Unmarshal(body, result)
	return result, err
}

// Retrieve customer object. 顧客情報を取得します。
// https://pay.jp/docs/api/?go#顧客情報を取得
func (c CustomerService) Retrieve(id string) (*CustomerResponse, error) {
	body, err := c.service.request("GET", "/customers/"+id, nil)
	if err != nil {
		return nil, err
	}
	result := &CustomerResponse{service: c.service}
	err = json.Unmarshal(body, result)
	return result, err
}

// Update は生成した顧客情報を更新したり、新たなカードを顧客に追加します。
// https://pay.jp/docs/api/?go#顧客情報を更新
func (c CustomerService) Update(id string, customer Customer) (*CustomerResponse, error) {
	qb := newRequestBuilder()
	qb.Add("email", customer.Email)
	qb.Add("description", customer.Description)
	qb.Add("default_card", customer.DefaultCard)
	qb.Add("card", customer.CardToken)
	qb.AddMetadata(customer.Metadata)

	body, err := c.service.request("POST", "/customers/"+id, qb.Reader())
	if err != nil {
		return nil, err
	}
	result := &CustomerResponse{service: c.service}
	err = json.Unmarshal(body, result)
	return result, err
}

// Delete は生成した顧客情報を削除します。削除した顧客情報は、もう一度生成することができないためご注意ください。
func (c CustomerService) Delete(id string) error {
	return c.service.delete("/customers/" + id)
}

// Deprecated
func (c CustomerService) List() *CustomerListCaller {
	return &CustomerListCaller{
		service: c,
	}
}

// AddCardToken はトークンIDを指定して、新たにカードを追加します。
func (c CustomerService) AddCardToken(id string, token string, options ...Customer) (*CardResponse, error) {
	qb := newRequestBuilder()
	qb.Add("card", token)
	if len(options) > 0 {
		qb.Add("default", options[0].DefaultCard)
		qb.AddMetadata(options[0].Metadata)
	}
	body, err := c.service.request("POST", "/customers/"+id+"/cards", qb.Reader())
	if err != nil {
		return nil, err
	}
	result := &CardResponse{
		service:    c.service,
		customerID: id,
	}
	err = json.Unmarshal(body, result)
	return result, err
}

// GetCard は顧客の特定のカード情報を取得します。
func (c CustomerService) GetCard(id, cardID string) (*CardResponse, error) {
	body, err := c.service.request("GET", "/customers/"+id+"/cards/"+cardID, nil)
	if err != nil {
		return nil, err
	}
	result := &CardResponse{
		service:    c.service,
		customerID: id,
	}
	err = json.Unmarshal(body, result)
	return result, err
}

// UpdateCard は顧客の特定のカード情報を更新します。
func (c CustomerService) UpdateCard(id string, cardID string, card Card) (*CardResponse, error) {
	qb := newRequestBuilder()
	qb.Add("card[address_state]", card.AddressState)
	qb.Add("card[address_city]", card.AddressCity)
	qb.Add("card[address_line1]", card.AddressLine1)
	qb.Add("card[address_line2]", card.AddressLine2)
	qb.Add("card[address_zip]", card.AddressZip)
	qb.Add("card[country]", card.Country)
	qb.Add("card[name]", card.Name)
	qb.AddMetadata(card.Metadata)

	body, err := c.service.request("POST", "/customers/"+id+"/cards/"+cardID, qb.Reader())
	if err != nil {
		return nil, err
	}
	result := &CardResponse{
		service:    c.service,
		customerID: id,
	}
	err = json.Unmarshal(body, result)
	return result, err
}

// DeleteCard は顧客の特定のカードを削除します。
func (c CustomerService) DeleteCard(customerID, cardID string) error {
	return c.service.delete("/customers/" + customerID + "/cards/" + cardID)
}

// deprecated
func (c CustomerService) ListCard(id string) *CustomerCardListCaller {
	cus, err := c.Retrieve(id)
	if err != nil {
		return nil
	}

	return &CustomerCardListCaller{
		caller: cus,
	}
}

// GetSubscription は顧客の特定の定期課金情報を取得します。
func (c CustomerService) GetSubscription(customerID, subscriptionID string) (*SubscriptionResponse, error) {
	return c.service.Subscription.Retrieve(customerID, subscriptionID)
}

// ListSubscription は顧客の定期課金リストを取得します。リストは、直近で生成された順番に取得されます。
func (c CustomerService) ListSubscription(customerID string) *subscriptionListCaller {
	return &subscriptionListCaller{
		service: *c.service.Subscription,
		SubscriptionListParams: SubscriptionListParams{
			Customer: &customerID,
		},
	}
}

type CustomerListParams struct {
	ListParams `form:"*"`
}

type CustomerListCaller struct {
	service CustomerService
	CustomerListParams
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *CustomerListCaller) Limit(limit int) *CustomerListCaller {
	c.CustomerListParams.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *CustomerListCaller) Offset(offset int) *CustomerListCaller {
	c.CustomerListParams.ListParams.Offset = &offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *CustomerListCaller) Since(since time.Time) *CustomerListCaller {
	p := int(since.Unix())
	c.CustomerListParams.ListParams.Since = &p
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *CustomerListCaller) Until(until time.Time) *CustomerListCaller {
	p := int(until.Unix())
	c.CustomerListParams.ListParams.Until = &p
	return c
}

// Do は指定されたクエリーを元に顧客のリストを配列で取得します。
func (c *CustomerListCaller) Do() ([]*CustomerResponse, bool, error) {
	return c.service.All(&c.CustomerListParams)
}

func (c CustomerService) All(params ...*CustomerListParams) ([]*CustomerResponse, bool, error) {
	p := &CustomerListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/customers"+c.service.getQuery(p), nil)

	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*CustomerResponse, raw.Count)

	for i, raw := range raw.Data {
		json.Unmarshal(raw, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}

type CardListParams struct {
	ListParams `form:"*"`
}

// CustomerCardListCaller はカードのリスト取得に使用する構造体です。
type CustomerCardListCaller struct {
	caller *CustomerResponse
	CardListParams
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *CustomerCardListCaller) Limit(limit int) *CustomerCardListCaller {
	c.CardListParams.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *CustomerCardListCaller) Offset(offset int) *CustomerCardListCaller {
	c.CardListParams.ListParams.Offset = &offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *CustomerCardListCaller) Since(since time.Time) *CustomerCardListCaller {
	p := int(since.Unix())
	c.CardListParams.ListParams.Since = &p
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *CustomerCardListCaller) Until(until time.Time) *CustomerCardListCaller {
	p := int(until.Unix())
	c.CardListParams.ListParams.Until = &p
	return c
}

// Do は指定されたクエリーを元に支払いのリストを配列で取得します。
func (c *CustomerCardListCaller) Do() ([]*CardResponse, bool, error) {
	return c.caller.All(&c.CardListParams)
}

func (c *CustomerResponse) All(params ...*CardListParams) ([]*CardResponse, bool, error) {
	p := &CardListParams{}
	if len(params) > 0 {
		p = params[0]
	}
	body, err := c.service.request("GET", "/customers/"+c.ID+"/cards"+c.service.getQuery(p), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*CardResponse, raw.Count)
	for i, rawCustomer := range raw.Data {
		json.Unmarshal(rawCustomer, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}

// CustomerResponse はCustomerService.GetやCustomerService.Listで返される顧客を表す構造体です
type CustomerResponse struct {
	RawCards         listResponseParser `json:"cards"`
	Cards            []*CardResponse
	DefaultCard      string             `json:"default_card"`
	Description      string             `json:"description"`
	Email            string             `json:"email"`
	ID               string             `json:"id"`
	LiveMode         bool               `json:"livemode"`
	Object           string             `json:"object"`
	RawSubscriptions listResponseParser `json:"subscriptions"`
	Subscriptions    []*SubscriptionResponse
	Metadata         map[string]string `json:"metadata"`
	Created          *int              `json:"created"`
	CreatedAt        time.Time

	service *Service
}

// Update は生成した顧客情報を更新したり、新たなカードを顧客に追加することができます。
//
// また default_card に保持しているカードIDを指定することで、メイン利用のカードを変更することもできます。
func (c *CustomerResponse) Update(customer Customer) error {
	r, err := c.service.Customer.Update(c.ID, customer)
	if err != nil {
		return err
	}
	*c = *r
	return nil
}

// Delete は生成した顧客情報を削除します。削除した顧客情報は、もう一度生成することができないためご注意ください。
func (c *CustomerResponse) Delete() error {
	return c.service.Customer.Delete(c.ID)
}

// GetCard は顧客の特定のカード情報を取得します。
func (c *CustomerResponse) GetCard(cardID string) (*CardResponse, error) {
	return c.service.Customer.GetCard(c.ID, cardID)
}

// AddCardToken をCustomerResponseから実行します。
func (c *CustomerResponse) AddCardToken(token string, options ...Customer) (*CardResponse, error) {
	return c.service.Customer.AddCardToken(c.ID, token, options...)
}

// UpdateCard は顧客の特定のカード情報を更新します。
func (c CustomerResponse) UpdateCard(cardID string, card Card) (*CardResponse, error) {
	return c.service.Customer.UpdateCard(c.ID, cardID, card)
}

// DeleteCard は顧客の特定のカードを削除します。
func (c CustomerResponse) DeleteCard(cardID string) error {
	return c.service.Customer.DeleteCard(c.ID, cardID)
}

// ListCard は顧客の保持しているカードリストを取得します。リストは、直近で生成された順番に取得されます。
func (c *CustomerResponse) ListCard() *CustomerCardListCaller {
	return c.service.Customer.ListCard(c.ID)
}

// GetSubscription は顧客の特定の定期課金情報を取得します。
func (c *CustomerResponse) GetSubscription(subscriptionID string) (*SubscriptionResponse, error) {
	return c.service.Customer.GetSubscription(c.ID, subscriptionID)
}

// deprecated
func (c *CustomerResponse) ListSubscription() *subscriptionListCaller {
	return c.service.Customer.ListSubscription(c.ID)
}

func (c *CustomerResponse) UnmarshalJSON(b []byte) error {
	type customerResponseParser CustomerResponse
	var raw customerResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "customer" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.Cards = make([]*CardResponse, raw.RawCards.Count)
		raw.Subscriptions = make([]*SubscriptionResponse, raw.RawSubscriptions.Count)
		for i, rawCard := range raw.RawCards.Data {
			json.Unmarshal(rawCard, &raw.Cards[i])
			raw.Cards[i].service = c.service
			raw.Cards[i].customerID = raw.ID
		}
		for i, rawSubscription := range raw.RawSubscriptions.Data {
			json.Unmarshal(rawSubscription, &raw.Subscriptions[i])
			raw.Subscriptions[i].service = c.service
		}
		raw.service = c.service
		*c = CustomerResponse(raw)
		return nil
	}
	return parseError(b)
}
