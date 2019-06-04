package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

type allTestValues struct {
	strValue    string
	uint64Value uint64
	uint32Value uint32
	uint16Value uint16
	int64Value  int64
	int32Value  int32
	int16Value  int16
	boolValue   bool
}

var (
	defaultOptionStr = PluginConfigOption{
		Argument:  "arg1",
		Default:   "Default1",
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
	}

	defaultOptionUint64 = PluginConfigOption{
		Argument:  "arg2",
		Default:   uint64(22222),
		Env:       "ENV_2",
		Path:      "path2",
		Shorthand: "e",
		Usage:     "Second argument",
	}

	defaultOptionUint32 = PluginConfigOption{
		Argument:  "arg3",
		Default:   uint32(33333),
		Env:       "ENV_3",
		Path:      "path3",
		Shorthand: "g",
		Usage:     "Third argument",
	}

	defaultOptionUint16 = PluginConfigOption{
		Argument:  "arg4",
		Default:   uint16(44444),
		Env:       "ENV_4",
		Path:      "path4",
		Shorthand: "h",
		Usage:     "Fourth argument",
	}

	defaultOptionInt64 = PluginConfigOption{
		Argument:  "arg5",
		Default:   int64(55555),
		Env:       "ENV_5",
		Path:      "path5",
		Shorthand: "i",
		Usage:     "Fifth argument",
	}

	defaultOptionInt32 = PluginConfigOption{
		Argument:  "arg6",
		Default:   int32(666666),
		Env:       "ENV_6",
		Path:      "path6",
		Shorthand: "j",
		Usage:     "Sixth argument",
	}

	defaultOptionInt16 = PluginConfigOption{
		Argument:  "arg7",
		Default:   int16(7777),
		Env:       "ENV_7",
		Path:      "path7",
		Shorthand: "k",
		Usage:     "Seventh argument",
	}

	defaultOptionBool = PluginConfigOption{
		Argument:  "arg8",
		Default:   false,
		Env:       "ENV_8",
		Path:      "path8",
		Shorthand: "f",
		Usage:     "Eighth argument",
	}

	allOptions = []*PluginConfigOption{
		&defaultOptionStr,
		&defaultOptionUint64,
		&defaultOptionUint32,
		&defaultOptionUint16,
		&defaultOptionInt64,
		&defaultOptionInt32,
		&defaultOptionInt16,
		&defaultOptionBool,
	}

	defaultPluginConfig = &PluginConfig{
		Name:     "TestPluginConfig",
		Short:    "test-plugin-config",
		Timeout:  10,
		Keyspace: "sensu.io/plugins/test/config",
	}
)

func TestSetOptionValue_String(t *testing.T) {
	finalValue := ""
	option := defaultOptionStr
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.Nil(t, err)
	assert.Equal(t, "abc", finalValue)
}

func TestSetOptionValue_EmptyString(t *testing.T) {
	finalValue := ""
	option := defaultOptionStr
	option.Value = &finalValue
	err := setOptionValue(&option, "")
	assert.Nil(t, err)
	assert.Equal(t, "", finalValue)
}

func TestSetOptionValue_ValidUint64(t *testing.T) {
	var finalValue uint64
	option := defaultOptionUint64
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), finalValue)
}

func TestSetOptionValue_InvalidUint64(t *testing.T) {
	var finalValue uint64
	option := defaultOptionUint64
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), finalValue)
}

func TestSetOptionValue_ValidUint32(t *testing.T) {
	var finalValue uint32
	option := defaultOptionUint32
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, uint32(123), finalValue)
}

func TestSetOptionValue_InvalidUint32(t *testing.T) {
	var finalValue uint32
	option := defaultOptionUint32
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint32(0), finalValue)
}

func TestSetOptionValue_ValidUint16(t *testing.T) {
	var finalValue uint16
	option := defaultOptionUint16
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, uint16(123), finalValue)
}

func TestSetOptionValue_InvalidUint16(t *testing.T) {
	var finalValue uint16
	option := defaultOptionUint16
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint16(0), finalValue)
}

func TestSetOptionValue_ValidInt64(t *testing.T) {
	var finalValue int64
	option := defaultOptionInt64
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, int64(123), finalValue)
}

func TestSetOptionValue_InvalidInt64(t *testing.T) {
	var finalValue int64
	option := defaultOptionInt64
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), finalValue)
}

func TestSetOptionValue_ValidInt32(t *testing.T) {
	var finalValue int32
	option := defaultOptionInt32
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, int32(123), finalValue)
}

func TestSetOptionValue_InvalidInt32(t *testing.T) {
	var finalValue int32
	option := defaultOptionInt32
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, int32(0), finalValue)
}

func TestSetOptionValue_ValidInt16(t *testing.T) {
	var finalValue int16
	option := defaultOptionInt16
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, int16(123), finalValue)
}

func TestSetOptionValue_InvalidInt16(t *testing.T) {
	var finalValue int16
	option := defaultOptionInt16
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, int16(0), finalValue)
}

func TestSetOptionValue_TrueBool(t *testing.T) {
	var finalValue bool
	option := defaultOptionBool
	option.Value = &finalValue
	err := setOptionValue(&option, "true")
	assert.Nil(t, err)
	assert.Equal(t, true, finalValue)
}

func TestSetOptionValue_FalseBool(t *testing.T) {
	finalValue := true
	option := defaultOptionBool
	option.Value = &finalValue
	err := setOptionValue(&option, "false")
	assert.Nil(t, err)
	assert.Equal(t, false, finalValue)
}

func TestSetOptionValue_InvalidBool(t *testing.T) {
	var finalValue bool
	option := defaultOptionBool
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

func goPluginExecuteUtil(t *testing.T, eventFile string, cmdLineArgs []string,
	workflowFunction func([]string) (int, error),
	expectedValue1 interface{}, expectedValue2 interface{}, expectedValue3 interface{}) (int, string) {
	values := allTestValues{}
	options := getTestOptions(&values)
	var eventReader io.Reader
	if len(eventFile) > 0 {
		eventReader = getFileReader(eventFile)
	}

	goPlugin := newTestGoPlugin(options, eventReader, workflowFunction)
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

	assert.Equal(t, expectedValue1, values.strValue)
	assert.Equal(t, expectedValue2, values.uint64Value)
	assert.Equal(t, expectedValue3, values.boolValue)

	return exitStatus, errorStr
}

func newTestGoPlugin(options []*PluginConfigOption, eventReader io.Reader, workflowFunction func([]string) (int, error)) GoPlugin {
	goPlugin := &basePlugin{
		config:                 defaultPluginConfig,
		options:                options,
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

func getTestOptions(values *allTestValues) []*PluginConfigOption {
	strOption := defaultOptionStr
	uint64Option := defaultOptionUint64
	uint32Option := defaultOptionUint32
	uint16Option := defaultOptionUint16
	int64Option := defaultOptionInt64
	int32Option := defaultOptionInt32
	int16Option := defaultOptionInt16
	boolOption := defaultOptionBool

	if values != nil {
		strOption.Value = &values.strValue
		uint64Option.Value = &values.uint64Value
		uint32Option.Value = &values.uint32Value
		uint16Option.Value = &values.uint16Value
		int64Option.Value = &values.int64Value
		int32Option.Value = &values.int32Value
		int16Option.Value = &values.int16Value
		boolOption.Value = &values.boolValue
	} else {
		strOption.Value = nil
		uint64Option.Value = nil
		uint32Option.Value = nil
		uint16Option.Value = nil
		int64Option.Value = nil
		int32Option.Value = nil
		int16Option.Value = nil
		boolOption.Value = nil
	}

	return []*PluginConfigOption{
		&strOption,
		&uint64Option,
		&uint32Option,
		&uint16Option,
		&int64Option,
		&int32Option,
		&int16Option,
		&boolOption}
}
