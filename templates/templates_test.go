package templates

import (
	"encoding/json"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	templateOk          = "Check: {{ .Check.Name }} Entity: {{ .Entity.Name }} !"
	templateOkUnixTime  = "Check: {{ .Check.Name }} Entity: {{ .Entity.Name }} Executed: {{(UnixTime .Check.Executed).Format \"2 Jan 2006 15:04:05\"}} !"
	templateVarNotFound = "Check: {{ .Check.NameZZZ }} Entity: {{ .Entity.Name }} !"
	templateInvalid     = "Check: {{ .Check.Name Entity: {{ .Entity.Name }} !"
)

// Valid test
func TestEvalTemplate_Valid(t *testing.T) {
	event := &types.Event{}
	_ = json.Unmarshal(testEventBytes, event)

	result, err := EvalTemplate("templOk", templateOk, event)
	assert.Nil(t, err)
	assert.Equal(t, "Check: check-nginx Entity: webserver01 !", result)
}

// Valid test - Time Check
func TestEvalTemplateUnixTime_Valid(t *testing.T) {
	event := &types.Event{}
	_ = json.Unmarshal(testEventBytes, event)

	result, err := EvalTemplate("templOk", templateOkUnixTime, event)
	assert.Nil(t, err)
	assert.Equal(t, "Check: check-nginx Entity: webserver01 Executed: 10 Dec 2018 20:55:19 !", result)
}

// Variable not found
func TestEvalTemplate_VarNotFound(t *testing.T) {
	event := &types.Event{}
	_ = json.Unmarshal(testEventBytes, event)

	result, err := EvalTemplate("templOk", templateVarNotFound, event)
	assert.Equal(t, "", result)
	assert.NotNil(t, err)
}

// Nil template source
func TestEvalTemplate_NilSource(t *testing.T) {
	result, err := EvalTemplate("templOk", templateVarNotFound, nil)
	assert.Equal(t, "", result)
	assert.NotNil(t, err)
}

// Empty template
func TestEvalTemplate_NilTemplate(t *testing.T) {
	event := &types.Event{}
	_ = json.Unmarshal(testEventBytes, event)

	result, err := EvalTemplate("templOk", "", event)
	assert.Equal(t, "", result)
	assert.NotNil(t, err)
}

// Invalid template
func TestEvalTemplate_InvalidTemplate(t *testing.T) {
	event := &types.Event{}
	_ = json.Unmarshal(testEventBytes, event)

	result, err := EvalTemplate("templOk", templateInvalid, event)
	assert.Equal(t, "", result)
	assert.NotNil(t, err)
}
