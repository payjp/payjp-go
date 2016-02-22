package payjp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error はPAY.JP固有のエラーを表す構造体です
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param"`
	Status  int    `json:"status"`
	Type    string `json:"type"`
}

func (ce Error) Error() string {
	if ce.Param != "" {
		return fmt.Sprintf("%d: Type: %s Code: %s Message: %s, Param: %s", ce.Status, ce.Type, ce.Code, ce.Message, ce.Param)
	}
	return fmt.Sprintf("%d: Type: %s Code: %s Message: %s", ce.Status, ce.Type, ce.Code, ce.Message)
}

type errorResponse struct {
	Error Error `json:"error"`
}

func parseResponseError(resp *http.Response, err error) ([]byte, error) {
	body, err := respToBody(resp, err)
	if err != nil {
		return nil, err
	}
	payjpError := &Error{}
	err = json.Unmarshal(body, payjpError)
	if err != nil {
		// ignore JSON parsing error.
		// Subscription JSON has same name property 'status' but it is string.
		// it would be error, but it can be omitted.
		return body, nil
	}
	if payjpError.Status != 0 {
		return nil, payjpError
	}
	return body, nil
}
