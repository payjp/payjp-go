package payjp

import (
	"encoding/json"
	"net/http"
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
	Card        Card              // カード
	Metadata    map[string]string // メタデータ
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

// Create はメールアドレスやIDなどを指定して顧客を作成します。
//
// 作成と同時にカード情報を登録する場合、トークンIDかカードオブジェクトのどちらかを指定します。
//
// 作成した顧客やカード情報はあとから更新・削除することができます。
//
// DefaultCardは更新時のみ設定が可能です
func (c CustomerService) Create(customer Customer) (*CustomerResponse, error) {
	qb := newRequestBuilder()
	if customer.Email != "" {
		qb.Add("email", customer.Email)
	}
	if customer.Description != "" {
		qb.Add("description", customer.Description)
	}
	if customer.ID != "" {
		qb.Add("id", customer.ID)
	}
	if customer.CardToken != "" {
		qb.Add("card", customer.CardToken)
	} else {
		qb.AddCard(customer.Card)
	}
	qb.AddMetadata(customer.Metadata)

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

// Get は生成した顧客情報を取得します。
func (c CustomerService) Retrieve(id string) (*CustomerResponse, error) {
	body, err := c.service.retrieve("/customers/" + id)
	if err != nil {
		return nil, err
	}
	return parseCustomer(c.service, body, &CustomerResponse{})
}

// Update は生成した顧客情報を更新したり、新たなカードを顧客に追加します。
//
// また default_card に保持しているカードIDを指定することで、メイン利用のカードを変更することもできます。
func (c CustomerService) Update(id string, customer Customer) (*CustomerResponse, error) {
	body, err := c.update(id, customer)
	if err != nil {
		return nil, err
	}
	return parseCustomer(c.service, body, &CustomerResponse{})
}

func (c CustomerService) update(id string, customer Customer) ([]byte, error) {
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
	qb.AddMetadata(customer.Metadata)
	request, err := http.NewRequest("POST", c.service.apiBase+"/customers/"+id, qb.Reader())
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", c.service.apiKey)

	return parseResponseError(c.service.Client.Do(request))
}

// Delete は生成した顧客情報を削除します。削除した顧客情報は、もう一度生成することができないためご注意ください。
func (c CustomerService) Delete(id string) error {
	return c.service.delete("/customers/" + id)
}

// List は生成した顧客情報のリストを取得します。リストは、直近で生成された順番に取得されます。
func (c CustomerService) List() *CustomerListCaller {
	return &CustomerListCaller{
		service: c.service,
	}
}

// AddCardToken はトークンIDを指定して、新たにカードを追加します。ただし同じカード番号および同じ有効期限年/月のカードは、重複追加することができません。
func (c CustomerService) AddCardToken(customerID, token string) (*CardResponse, error) {
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

func (c CustomerService) postCard(customerID, resourcePath string, card Card, result *CardResponse) (*CardResponse, error) {
	qb := newRequestBuilder()
	qb.AddCardFlat(card)

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

// AddCard はカード情報のパラメーターを指定して、新たにカードを追加します。ただし同じカード番号および同じ有効期限年/月のカードは、重複追加することができません。
func (c CustomerService) AddCard(customerID string, card Card) (*CardResponse, error) {
	return c.postCard(customerID, "", card, &CardResponse{})
}

// GetCard は顧客の特定のカード情報を取得します。
func (c CustomerService) GetCard(customerID, cardID string) (*CardResponse, error) {
	body, err := c.service.retrieve("/customers/" + customerID + "/cards/" + cardID)
	if err != nil {
		return nil, err
	}
	return parseCard(c.service, body, &CardResponse{}, customerID)
}

// UpdateCard は顧客の特定のカード情報を更新します。
func (c CustomerService) UpdateCard(customerID, cardID string, card Card) (*CardResponse, error) {
	result := &CardResponse{
		customerID: customerID,
		service:    c.service,
	}
	return c.postCard(customerID, "/"+cardID, card, result)
}

// DeleteCard は顧客の特定のカードを削除します。
func (c CustomerService) DeleteCard(customerID, cardID string) error {
	return c.service.delete("/customers/" + customerID + "/cards/" + cardID)
}

// ListCard は顧客の保持しているカードリストを取得します。リストは、直近で生成された順番に取得されます。
func (c CustomerService) ListCard(customerID string) *CustomerCardListCaller {
	return &CustomerCardListCaller{
		service:    c.service,
		customerID: customerID,
	}
}

// GetSubscription は顧客の特定の定期課金情報を取得します。
func (c CustomerService) GetSubscription(customerID, subscriptionID string) (*SubscriptionResponse, error) {
	return c.service.Subscription.Retrieve(customerID, subscriptionID)
}

// ListSubscription は顧客の定期課金リストを取得します。リストは、直近で生成された順番に取得されます。
func (c CustomerService) ListSubscription(customerID string) *SubscriptionListCaller {
	return &SubscriptionListCaller{
		service:    c.service,
		customerID: customerID,
	}
}

// CustomerListCaller はリスト取得に使用する構造体です。
//
// Fluentインタフェースを提供しており、最後にDoを呼ぶことでリストが取得できます:
//
//     pay := payjp.New("api-key", nil)
//     customers, err := pay.Customer.List().Limit(50).Offset(150).Do()
type CustomerListCaller struct {
	service *Service
	limit   int
	offset  int
	since   int
	until   int
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *CustomerListCaller) Limit(limit int) *CustomerListCaller {
	c.limit = limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *CustomerListCaller) Offset(offset int) *CustomerListCaller {
	c.offset = offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *CustomerListCaller) Since(since time.Time) *CustomerListCaller {
	c.since = int(since.Unix())
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *CustomerListCaller) Until(until time.Time) *CustomerListCaller {
	c.until = int(until.Unix())
	return c
}

// Do は指定されたクエリーを元に顧客のリストを配列で取得します。
func (c *CustomerListCaller) Do() ([]*CustomerResponse, bool, error) {
	body, err := c.service.queryList("/customers", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*CustomerResponse, len(raw.Data))

	for i, raw := range raw.Data {
		customer := &CustomerResponse{}
		json.Unmarshal(raw, customer)
		customer.service = c.service
		result[i] = customer
	}
	return result, raw.HasMore, nil
}

// CustomerCardListCaller はカードのリスト取得に使用する構造体です。
//
// Fluentインタフェースを提供しており、最後にDoを呼ぶことでリストが取得できます:
//
//     pay := payjp.New("api-key", nil)
//     cards, err := pay.Customer.ListCard("userID").Limit(50).Offset(150).Do()
type CustomerCardListCaller struct {
	service    *Service
	customerID string
	limit      int
	offset     int
	since      int
	until      int
}

// Limit はリストの要素数の最大値を設定します(1-100)
func (c *CustomerCardListCaller) Limit(limit int) *CustomerCardListCaller {
	c.limit = limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *CustomerCardListCaller) Offset(offset int) *CustomerCardListCaller {
	c.offset = offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *CustomerCardListCaller) Since(since time.Time) *CustomerCardListCaller {
	c.since = int(since.Unix())
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *CustomerCardListCaller) Until(until time.Time) *CustomerCardListCaller {
	c.until = int(until.Unix())
	return c
}

// Do は指定されたクエリーを元に支払いのリストを配列で取得します。
func (c *CustomerCardListCaller) Do() ([]*CardResponse, bool, error) {
	body, err := c.service.queryList("/customers/"+c.customerID+"/cards", c.limit, c.offset, c.since, c.until)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*CardResponse, len(raw.Data))
	for i, rawCustomer := range raw.Data {
		card := &CardResponse{}
		json.Unmarshal(rawCustomer, card)
		result[i] = card
	}
	return result, raw.HasMore, nil
}

// CustomerResponse はCustomerService.GetやCustomerService.Listで返される顧客を表す構造体です
type CustomerResponse struct {
	ID            string                  // 一意なオブジェクトを示す文字列
	LiveMode      bool                    // 本番環境かどうか
	CreatedAt     time.Time               // この顧客作成時のタイムスタンプ
	DefaultCard   string                  // 支払いに使用されるカードのcar_から始まるID
	Cards         []*CardResponse         // この顧客に紐づけられているカードのリスト
	Email         string                  // メールアドレス
	Description   string                  // 概要
	Subscriptions []*SubscriptionResponse // この顧客が購読している定期課金のリスト
	Metadata      map[string]string       // メタデータ

	service *Service
}

type customerResponseParser struct {
	Cards         listResponseParser `json:"cards"`
	CreatedEpoch  int                `json:"created"`
	DefaultCard   string             `json:"default_card"`
	Description   string             `json:"description"`
	Email         string             `json:"email"`
	ID            string             `json:"id"`
	LiveMode      bool               `json:"livemode"`
	Object        string             `json:"object"`
	Subscriptions listResponseParser `json:"subscriptions"`
	Metadata      map[string]string  `json:"metadata"`
}

// Update は生成した顧客情報を更新したり、新たなカードを顧客に追加することができます。
//
// また default_card に保持しているカードIDを指定することで、メイン利用のカードを変更することもできます。
func (c *CustomerResponse) Update(customer Customer) error {
	body, err := c.service.Customer.update(c.ID, customer)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, c)
}

// Delete は生成した顧客情報を削除します。削除した顧客情報は、もう一度生成することができないためご注意ください。
func (c *CustomerResponse) Delete() error {
	return c.service.Customer.Delete(c.ID)
}

// AddCard はカード情報のパラメーターを指定して、新たにカードを追加します。ただし同じカード番号および同じ有効期限年/月のカードは、重複追加することができません。
func (c *CustomerResponse) AddCard(card Card) (*CardResponse, error) {
	return c.service.Customer.AddCard(c.ID, card)
}

// AddCardToken はトークンIDを指定して、新たにカードを追加します。ただし同じカード番号および同じ有効期限年/月のカードは、重複追加することができません。
func (c *CustomerResponse) AddCardToken(token string) (*CardResponse, error) {
	return c.service.Customer.AddCardToken(c.ID, token)
}

// GetCard は顧客の特定のカード情報を取得します。
func (c *CustomerResponse) GetCard(cardID string) (*CardResponse, error) {
	return c.service.Customer.GetCard(c.ID, cardID)
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

// ListSubscription は顧客の定期課金リストを取得します。リストは、直近で生成された順番に取得されます。
func (c *CustomerResponse) ListSubscription() *SubscriptionListCaller {
	return c.service.Customer.ListSubscription(c.ID)
}

// UnmarshalJSON はJSONパース用の内部APIです。
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
		c.Metadata = raw.Metadata
		return nil
	}
	rawError := errorResponse{}
	err = json.Unmarshal(b, &rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}

	return nil
}
