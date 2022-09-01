package payjp

import (
	"encoding/json"
	"testing"
	"github.com/stretchr/testify/assert"
)

var errorJSONStr = `
{
  "code": "code",
  "message": "message",
  "param": "param",
  "status": 400,
  "type": "type"
}
`
var errorJSON = []byte(errorJSONStr)
var errorResponseJSON = []byte(`{"error":`+errorJSONStr+`}`)
var errorStr = "400: Type: type Code: code Message: message, Param: param"

func TestErrorJson(t *testing.T) {
	actual := &errorResponse{}
	err := json.Unmarshal(errorResponseJSON, actual)

	assert.NoError(t, err)
	assert.IsType(t, &errorResponse{}, actual)
	assert.Equal(t, 400, actual.Error.Status)
	assert.Equal(t, "type", actual.Error.Type)
	assert.Equal(t, "param", actual.Error.Param)
	assert.Equal(t, "message", actual.Error.Message)
	assert.Equal(t, "code", actual.Error.Code)

	actual2 := &Error{}
	err = json.Unmarshal(errorJSON, actual2)

	assert.NoError(t, err)
	assert.IsType(t, &Error{}, actual2)
	assert.Equal(t, actual2.Status, actual.Error.Status)
	assert.Equal(t, actual2.Type, actual.Error.Type)
	assert.Equal(t, actual2.Param, actual.Error.Param)
	assert.Equal(t, actual2.Message, actual.Error.Message)
	assert.Equal(t, actual2.Code, actual.Error.Code)

	assert.Equal(t, errorStr, actual2.Error())
	actual2.Param = ""
	assert.Equal(t, "400: Type: type Code: code Message: message", actual2.Error())
}

