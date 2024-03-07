package payjp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var errorResponseJSON = []byte(`{"error":` + errorJSONStr + `}`)
var errorStr = "400: Type: type Code: code Message: message, Param: param"

func TestErrorJson(t *testing.T) {
	err := parseError([]byte(`{}`))
	assert.NoError(t, err)
	assert.Nil(t, err)

	err = parseError(errorResponseJSON)
	assert.Error(t, err)
	assert.IsType(t, &Error{}, err)
	assert.Equal(t, errorStr, err.Error())
	payErr := err.(*Error)
	assert.Equal(t, 400, payErr.Status)
	assert.Equal(t, "type", payErr.Type)
	assert.Equal(t, "param", payErr.Param)
	assert.Equal(t, "message", payErr.Message)
	assert.Equal(t, "code", payErr.Code)
	assert.Equal(t, errorStr, err.Error())
	payErr.Param = ""
	assert.Equal(t, "400: Type: type Code: code Message: message", payErr.Error())
}
