package sensu

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type checkValues struct {
	arg1 string
	arg2 uint64
	arg3 bool
}

var (
	defaultCheckConfig = PluginConfig{
		Name:     "TestHandler",
		Short:    "Short Description",
		Timeout:  10,
		Keyspace: "sensu.io/plugins/segp/config",
	}

	checkOption1 = stringOpt
	checkOption2 = uint64Opt
	checkOption3 = boolOpt
)

func TestNewGoCheck(t *testing.T) {
	values := &checkValues{}
	options := getCheckOptions(values)
	goCheck := NewGoCheck(&defaultCheckConfig, options, func(_ *corev2.Event) (int, error) {
		return 0, nil
	}, func(_ *corev2.Event) (int, error) {
		return 0, nil
	}, false)

	assert.NotNil(t, goCheck)
	assert.NotNil(t, goCheck.framework.options)
	assert.Equal(t, options, goCheck.framework.options)
	assert.NotNil(t, goCheck.framework.config)
	assert.Equal(t, &defaultHandlerConfig, goCheck.framework.config)
	assert.NotNil(t, goCheck.validationFunction)
	assert.NotNil(t, goCheck.executeFunction)
	assert.Nil(t, goCheck.framework.GetStdinEvent())
	assert.Equal(t, os.Stdin, goCheck.framework.eventReader)
}

func getCheckOptions(values *checkValues) []ConfigOption {
	option1 := checkOption1
	option2 := checkOption2
	option3 := checkOption3
	if values != nil {
		option1.Value = &values.arg1
		option2.Value = &values.arg2
		option3.Value = &values.arg3
	} else {
		option1.Value = nil
		option2.Value = nil
		option3.Value = nil
	}
	return []ConfigOption{&option1, &option2, &option3}
}
