package sensu

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sensu/sensu-go/types"
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

	checkOption1 = defaultOption1
	checkOption2 = defaultOption2
	checkOption3 = defaultOption3
)

func TestNewGoCheck(t *testing.T) {
	values := &checkValues{}
	options := getCheckOptions(values)
	goCheck := NewGoCheck(&defaultCheckConfig, options, func(event *types.Event) error {
		return nil
	}, func(event *types.Event) error {
		return nil
	})

	assert.NotNil(t, goCheck)
	assert.NotNil(t, goCheck.options)
	assert.Equal(t, options, goCheck.options)
	assert.NotNil(t, goCheck.config)
	assert.Equal(t, &defaultHandlerConfig, goCheck.config)
	assert.NotNil(t, goCheck.validationFunction)
	assert.NotNil(t, goCheck.executeFunction)
	assert.Nil(t, goCheck.sensuEvent)
	assert.Equal(t, os.Stdin, goCheck.eventReader)
}

func getCheckOptions(values *checkValues) []*PluginConfigOption {
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
	return []*PluginConfigOption{&option1, &option2, &option3}
}
