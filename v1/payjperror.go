package payjp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PayJpError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param"`
	Status  int    `json:"status"`
	Type    string `json:"type"`
}

type ErrorResponse struct {
	Error PayJpError `json:"error"`
}

func (ce PayJpError) Error() string {
	if ce.Param != "" {
		return fmt.Sprintf("%d: Type: %s Code: %s Message: %s, Param: %s", ce.Status, ce.Type, ce.Code, ce.Message, ce.Param)
	} else {
		return fmt.Sprintf("%d: Type: %s Code: %s Message: %s", ce.Status, ce.Type, ce.Code, ce.Message)
	}
}

func parseResponseError(resp *http.Response, err error) ([]byte, error) {
	body, err := respToBody(resp, err)
	if err != nil {
		return nil, err
	}
	payjpError := &PayJpError{}
	err = json.Unmarshal(body, payjpError)
	if err != nil {
		return nil, err
	}
	if payjpError.Status != 0 {
		return nil, payjpError
	}
	return body, nil
}
