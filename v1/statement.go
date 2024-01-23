package payjp

import (
	"encoding/json"
	"time"
)

type StatementService struct {
	service *Service
}

func newStatementService(service *Service) *StatementService {
	return &StatementService{
		service: service,
	}
}

type StatementUrls struct {
	Platformer interface{} // プラットフォーム手数料に関する明細か否か (bool)
}

func (s *StatementResponse) StatementUrls(p StatementUrls) (*StatementUrlResponse, error) {
	qb := newRequestBuilder()
	qb.Add("platformer", p.Platformer)

	body, err := s.service.request("POST", "/statements/"+s.ID+"/statement_urls", qb.Reader())
	if err != nil {
		return nil, err
	}
	result := &StatementUrlResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Retrieve Statement object.
func (s StatementService) Retrieve(id string) (*StatementResponse, error) {
	body, err := s.service.request("GET", "/statements/"+id, nil)
	if err != nil {
		return nil, err
	}
	result := &StatementResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.service = s.service
	return result, nil
}

// StatementResponse
type StatementResponse struct {
	ID        string  `json:"id"`
	LiveMode  bool    `json:"livemode"`
	Object    string  `json:"object"`
	Title     *string `json:"title"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Items     []StatementItem

	Created  *int              `json:"created"`
	Updated  *int              `json:"updated"`
	RawItems []json.RawMessage `json:"items"`

	service *Service
}

// StatementUrlResponse
type StatementUrlResponse struct {
	Expires int    `json:"expires"`
	Object  string `json:"object"`
	URL     string `json:"url"`
}

// StatementItem
type StatementItem struct {
	Amount  int    `json:"amount"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
	TaxRate string `json:"tax_rate"`
}

func (s *StatementResponse) UnmarshalJSON(b []byte) error {
	type statementResponseParser StatementResponse
	var raw statementResponseParser
	err := json.Unmarshal(b, &raw)
	if err == nil && raw.Object == "statement" {
		raw.CreatedAt = time.Unix(IntValue(raw.Created), 0)
		raw.UpdatedAt = time.Unix(IntValue(raw.Updated), 0)
		raw.Items = make([]StatementItem, len(raw.RawItems))
		for i, r := range raw.RawItems {
			json.Unmarshal(r, &(raw.Items[i]))
		}
		raw.service = s.service
		*s = StatementResponse(raw)
		return nil
	}
	return parseError(b)
}

type StatementListParams struct {
	ListParams     `form:"*"`
	Owner          *string `form:"owner"`
	SourceTransfer *string `form:"source_transfer"`
	Tenant         *string `form:"tenant"`

	service *Service `form:"-"`
}

// List Statements
func (s StatementService) List(params ...*StatementListParams) *StatementListParams {
	if len(params) > 0 {
		p := params[0]
		p.service = s.service
		return p
	}
	r := &StatementListParams{}
	r.service = s.service
	return r
}

// Limit はリストの要素数の最大値を設定します
func (c *StatementListParams) Limit(limit int) *StatementListParams {
	c.ListParams.Limit = &limit
	return c
}

// Offset は取得するリストの先頭要素のインデックスのオフセットを設定します
func (c *StatementListParams) Offset(offset int) *StatementListParams {
	c.ListParams.Offset = &offset
	return c
}

// Since はここに指定したタイムスタンプ以降に作成されたデータを取得します
func (c *StatementListParams) Since(p time.Time) *StatementListParams {
	i := int(p.Unix())
	c.ListParams.Since = &i
	return c
}

// Until はここに指定したタイムスタンプ以前に作成されたデータを取得します
func (c *StatementListParams) Until(p time.Time) *StatementListParams {
	i := int(p.Unix())
	c.ListParams.Until = &i
	return c
}

// Do は指定されたクエリーを元にリストを配列で取得します。
func (c *StatementListParams) Do() ([]*StatementResponse, bool, error) {
	body, err := c.service.request("GET", "/statements"+c.service.getQuery(c), nil)
	if err != nil {
		return nil, false, err
	}
	raw := &listResponseParser{}
	err = json.Unmarshal(body, raw)
	if err != nil {
		return nil, false, err
	}
	result := make([]*StatementResponse, len(raw.Data))
	for i, rawStatement := range raw.Data {
		json.Unmarshal(rawStatement, &result[i])
		result[i].service = c.service
	}
	return result, raw.HasMore, nil
}
