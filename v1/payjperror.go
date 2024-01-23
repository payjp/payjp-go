package payjp

import (
	"encoding/json"
	"fmt"
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

func parseError(body []byte) error {
	rawError := &errorResponse{}
	err := json.Unmarshal(body, rawError)
	if err == nil && rawError.Error.Status != 0 {
		return &rawError.Error
	}
	// ignore JSON parsing error.
	// Subscription JSON has same name property 'status' but it is string.
	// it would be error, but it can be omitted.
	return nil
}
