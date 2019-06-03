package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

type allValues struct {
	value1 string
	value2 uint64
	value3 uint32
	value4 uint16
	value5 int64
	value6 int32
	value7 int16
	value8 bool
}

var (
	defaultOption1 = PluginConfigOption{
		Argument:  "arg1",
		Default:   "Default1",
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
	}

	defaultOption2 = PluginConfigOption{
		Argument:  "arg2",
		Default:   uint64(22222),
		Env:       "ENV_2",
		Path:      "path2",
		Shorthand: "e",
		Usage:     "Second argument",
	}

	defaultOption3 = PluginConfigOption{
		Argument:  "arg3",
		Default:   uint32(33333),
		Env:       "ENV_3",
		Path:      "path3",
		Shorthand: "g",
		Usage:     "Third argument",
	}

	defaultOption4 = PluginConfigOption{
		Argument:  "arg4",
		Default:   uint16(44444),
		Env:       "ENV_4",
		Path:      "path4",
		Shorthand: "h",
		Usage:     "Fourth argument",
	}

	defaultOption5 = PluginConfigOption{
		Argument:  "arg5",
		Default:   int64(55555),
		Env:       "ENV_5",
		Path:      "path5",
		Shorthand: "i",
		Usage:     "Fifth argument",
	}

	defaultOption6 = PluginConfigOption{
		Argument:  "arg6",
		Default:   int32(666666),
		Env:       "ENV_6",
		Path:      "path6",
		Shorthand: "j",
		Usage:     "Sixth argument",
	}

	defaultOption7 = PluginConfigOption{
		Argument:  "arg7",
		Default:   int16(7777),
		Env:       "ENV_7",
		Path:      "path7",
		Shorthand: "k",
		Usage:     "Seventh argument",
	}

	defaultOption8 = PluginConfigOption{
		Argument:  "arg8",
		Default:   false,
		Env:       "ENV_8",
		Path:      "path8",
		Shorthand: "f",
		Usage:     "Eighth argument",
	}

	allOptions = []*PluginConfigOption{
		&defaultOption1,
		&defaultOption2,
		&defaultOption3,
		&defaultOption4,
		&defaultOption5,
		&defaultOption6,
		&defaultOption7,
		&defaultOption8,
	}
)

func TestSetOptionValue_String(t *testing.T) {
	finalValue := ""
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.Nil(t, err)
	assert.Equal(t, "abc", finalValue)
}

func TestSetOptionValue_EmptyString(t *testing.T) {
	finalValue := ""
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "")
	assert.Nil(t, err)
	assert.Equal(t, "", finalValue)
}

func TestSetOptionValue_ValidUint64(t *testing.T) {
	var finalValue uint64
	option := defaultOption2
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), finalValue)
}

func TestSetOptionValue_InvalidUint64(t *testing.T) {
	var finalValue uint64
	option := defaultOption2
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), finalValue)
}

func TestSetOptionValue_ValidUint32(t *testing.T) {
	var finalValue uint32
	option := defaultOption3
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, uint32(123), finalValue)
}

func TestSetOptionValue_InvalidUint32(t *testing.T) {
	var finalValue uint32
	option := defaultOption3
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint32(0), finalValue)
}

func TestSetOptionValue_ValidUint16(t *testing.T) {
	var finalValue uint16
	option := defaultOption4
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, uint16(123), finalValue)
}

func TestSetOptionValue_InvalidUint16(t *testing.T) {
	var finalValue uint16
	option := defaultOption4
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint16(0), finalValue)
}

func TestSetOptionValue_ValidInt64(t *testing.T) {
	var finalValue int64
	option := defaultOption5
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, int64(123), finalValue)
}

func TestSetOptionValue_InvalidInt64(t *testing.T) {
	var finalValue int64
	option := defaultOption5
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), finalValue)
}

func TestSetOptionValue_ValidInt32(t *testing.T) {
	var finalValue int32
	option := defaultOption6
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, int32(123), finalValue)
}

func TestSetOptionValue_InvalidInt32(t *testing.T) {
	var finalValue int32
	option := defaultOption6
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, int32(0), finalValue)
}

func TestSetOptionValue_ValidInt16(t *testing.T) {
	var finalValue int16
	option := defaultOption7
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, int16(123), finalValue)
}

func TestSetOptionValue_InvalidInt16(t *testing.T) {
	var finalValue int16
	option := defaultOption7
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, int16(0), finalValue)
}

func TestSetOptionValue_TrueBool(t *testing.T) {
	var finalValue bool
	option := defaultOption8
	option.Value = &finalValue
	err := setOptionValue(&option, "true")
	assert.Nil(t, err)
	assert.Equal(t, true, finalValue)
}

func TestSetOptionValue_FalseBool(t *testing.T) {
	finalValue := true
	option := defaultOption8
	option.Value = &finalValue
	err := setOptionValue(&option, "false")
	assert.Nil(t, err)
	assert.Equal(t, false, finalValue)
}

func TestSetOptionValue_InvalidBool(t *testing.T) {
	var finalValue bool
	option := defaultOption8
	option.Value = &finalValue
	err := setOptionValue(&option, "yes")
	assert.NotNil(t, err)
	assert.Equal(t, false, finalValue)
}

// Test cmd line arguments
func TestGoPlugin_Execute_CmdLineArgs(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

func getFileReader(file string) io.Reader {
	reader, _ := os.Open(file)
	return reader
}

func clearEnvironment() {
	_ = os.Unsetenv("ENV_1")
	_ = os.Unsetenv("ENV_2")
	_ = os.Unsetenv("ENV_3")
}

func goPluginExecuteUtil(t *testing.T, handlerConfig *PluginConfig, eventFile string, cmdLineArgs []string,
	validationFunction func(*types.Event) error, executeFunction func(*types.Event) error,
	expectedValue1 interface{}, expectedValue2 interface{}, expectedValue3 interface{}) (int, string) {
	values := handlerValues{}
	options := getHandlerOptions(&values)

	goPlugin := NewGoHandler(handlerConfig, options, validationFunction, executeFunction)
	goHandler := goPlugin.(*goHandler)

	// Simulate the command line arguments if necessary
	if len(cmdLineArgs) > 0 {
		goHandler.cmdArgs.SetArgs(cmdLineArgs)
	} else {
		goHandler.cmdArgs.SetArgs([]string{})
	}

	// Replace stdin reader with file reader and exitFunction with our own so we can know the exit status
	var exitStatus int
	var errorStr = ""
	goHandler.eventReader = getFileReader(eventFile)
	goHandler.exitFunction = func(i int) {
		exitStatus = i
	}
	goHandler.errorLogFunction = func(format string, a ...interface{}) {
		errorStr = fmt.Sprintf(format, a...)
	}
	goHandler.Execute()

	assert.Equal(t, expectedValue1, values.arg1)
	assert.Equal(t, expectedValue2, values.arg2)
	assert.Equal(t, expectedValue3, values.arg3)

	return exitStatus, errorStr
}

func newTestGoPlugin(values *allValues, eventReader io.Reader, workflowFunction func([]string) (int, error)) GoPlugin {
	setupValues(allOptions, values)

	goPlugin := &basePlugin{
		config: &PluginConfig{
			Name:     "TestHandler",
			Short:    "Short Description",
			Timeout:  10,
			Keyspace: "sensu.io/plugins/segp/config",
		},
		options:                allOptions,
		sensuEvent:             nil,
		eventReader:            eventReader,
		readEvent:              true,
		eventMandatory:         true,
		configurationOverrides: true,
		errorExitStatus:        1,
		pluginWorkflowFunction: workflowFunction,
	}

	return goPlugin
}

func setupValues(options []*PluginConfigOption, values *allValues) {
	options[0].Value = &values.value1
	options[1].Value = &values.value2
	options[2].Value = &values.value3
	options[3].Value = &values.value4
	options[4].Value = &values.value5
	options[5].Value = &values.value6
	options[6].Value = &values.value7
	options[7].Value = &values.value8
}
