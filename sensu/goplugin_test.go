package sensu

import (
	"fmt"
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
		Argument:  "strArg",
		Default:   "Default1",
		Env:       "STR_ENV",
		Path:      "strPath",
		Shorthand: "a",
		Usage:     "string argument",
	}

	defaultOptionUint64 = PluginConfigOption{
		Argument:  "uint64Arg",
		Default:   uint64(22222),
		Env:       "UINT64_ENV",
		Path:      "uint64Path",
		Shorthand: "b",
		Usage:     "uint64 argument",
	}

	defaultOptionUint32 = PluginConfigOption{
		Argument:  "uint32Arg",
		Default:   uint32(33333),
		Env:       "UINT32_ENV",
		Path:      "uint32Path",
		Shorthand: "c",
		Usage:     "uint32 argument",
	}

	defaultOptionUint16 = PluginConfigOption{
		Argument:  "uint16Arg",
		Default:   uint16(44444),
		Env:       "UINT16_ENV",
		Path:      "uint16Path",
		Shorthand: "d",
		Usage:     "uint16 argument",
	}

	defaultOptionInt64 = PluginConfigOption{
		Argument:  "int64Arg",
		Default:   int64(-11111),
		Env:       "INT64_ENV",
		Path:      "int64Path",
		Shorthand: "e",
		Usage:     "int64 argument",
	}

	defaultOptionInt32 = PluginConfigOption{
		Argument:  "int32Arg",
		Default:   int32(-33333),
		Env:       "INT32_ENV",
		Path:      "int32Path",
		Shorthand: "f",
		Usage:     "int32 argument",
	}

	defaultOptionInt16 = PluginConfigOption{
		Argument:  "int16Arg",
		Default:   int16(-4444),
		Env:       "INT16_ENV",
		Path:      "int16Path",
		Shorthand: "g",
		Usage:     "int16 argument",
	}

	defaultOptionBool = PluginConfigOption{
		Argument:  "boolArg",
		Default:   false,
		Env:       "BOOL_ENV",
		Path:      "boolPath",
		Shorthand: "i",
		Usage:     "bool argument",
	}

	defaultPluginConfig = &PluginConfig{
		Name:     "TestPluginConfig",
		Short:    "test-plugin-config",
		Timeout:  10,
		Keyspace: "sensu.io/plugins/test/config",
	}

	cmdLineArgsDefault = []string{"--uint64Arg", "99999", "--uint32Arg", "88888", "--uint16Arg", "7777",
		"--int64Arg", "-99999", "--int32Arg", "-88888", "--int16Arg", "-7777", "--strArg", "str", "--boolArg", "true"}

	cmdLineArgsShort = []string{"-a", "shortstr", "-b", "9999", "-c", "8888", "-d", "777", "-e", "-9999",
		"-f", "-8888", "-g", "-777", "-i", "true"}

	cmdLineArgsNoArgs = make([]string, 0)

	expectedCmdLineValues = allTestValues{
		uint64Value: 99999,
		uint32Value: 88888,
		uint16Value: 7777,
		int64Value:  -99999,
		int32Value:  -88888,
		int16Value:  -7777,
		strValue:    "str",
		boolValue:   true,
	}

	expectedShortCmdLineValues = allTestValues{
		uint64Value: 9999,
		uint32Value: 8888,
		uint16Value: 777,
		int64Value:  -9999,
		int32Value:  -8888,
		int16Value:  -777,
		strValue:    "shortstr",
		boolValue:   true,
	}

	expectedEntityOverrideValues = allTestValues{
		uint64Value: 98765,
		uint32Value: 87654,
		uint16Value: 65432,
		int64Value:  -98765,
		int32Value:  -87654,
		int16Value:  -7654,
		strValue:    "entity override",
		boolValue:   true,
	}

	expectedCheckOverrideValues = allTestValues{
		uint64Value: 198765,
		uint32Value: 187654,
		uint16Value: 5432,
		int64Value:  -198765,
		int32Value:  -187654,
		int16Value:  -654,
		strValue:    "check override",
		boolValue:   true,
	}

	expectedEnvironmentValues = allTestValues{
		uint64Value: 98989,
		uint32Value: 87878,
		uint16Value: 12121,
		int64Value:  -98989,
		int32Value:  -87878,
		int16Value:  -2121,
		strValue:    "env str",
		boolValue:   true,
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
	workflowExecuted := false
	clearEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_CmdLineArgs_InvalidValue(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	cmdLineArgs := append([]string(nil), cmdLineArgsDefault...)
	cmdLineArgs[1] = "not an int"

	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-override.json", cmdLineArgs, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Contains(t, errStr, "invalid argument")
	assert.Equal(t, 1, exitStatus)
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_Environment(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-override.json", cmdLineArgsNoArgs,
		&expectedEnvironmentValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_ShortCmdLineArgs(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-override.json", cmdLineArgsShort,
		&expectedShortCmdLineValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_ShortCmdLineArgs_InvalidValue(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	cmdLineArgs := append([]string(nil), cmdLineArgsShort...)
	cmdLineArgs[3] = "not an int"

	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-override.json", cmdLineArgs, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Contains(t, errStr, "invalid argument")
	assert.Equal(t, 1, exitStatus)
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_EntityOverride(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-entity-override.json", cmdLineArgsDefault,
		&expectedEntityOverrideValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_EntityOverride_InvalidValue(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()

	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-entity-override-invalid-value.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Contains(t, errStr, "Error parsing")
	assert.Equal(t, 1, exitStatus)
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_CheckOverride(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-check-override.json", cmdLineArgsDefault,
		&expectedCheckOverrideValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_CheckOverride_InvalidValue(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()

	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-check-override-invalid-value.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Contains(t, errStr, "Error parsing")
	assert.Equal(t, 1, exitStatus)
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_PriorityCheck(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-check-entity-override.json", cmdLineArgsDefault,
		&expectedCheckOverrideValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_PriorityEntity(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-entity-override.json", cmdLineArgsDefault,
		&expectedEntityOverrideValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_PriorityCmdLine(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 0, nil
		})

	assert.Equal(t, "", errStr)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_WorkflowError(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Contains(t, errStr, "something went wrong")
	assert.Equal(t, 2, exitStatus)
	assert.True(t, workflowExecuted)
}

func TestGoPlugin_Execute_EventNoTimestamp(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-timestamp.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errStr, "timestamp is missing or must be greater than zero")
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_EventZeroTimestamp(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-timestamp-zero.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errStr, "timestamp is missing or must be greater than zero")
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_EventNoEntity(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-entity.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errStr, "event must contain an entity")
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_EventInvalidEntity(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-invalid-entity.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errStr, "entity name must not be empty")
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_EventNoCheck(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-no-check.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errStr, "event must contain a check")
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_EventInvalidCheck(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-invalid-check.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errStr, "check name must not be empty")
	assert.False(t, workflowExecuted)
}

func TestGoPlugin_Execute_EventInvalidJson(t *testing.T) {
	workflowExecuted := false
	clearEnvironment()
	setEnvironment()
	exitStatus, errStr := goPluginExecuteUtil(t, "test/event-invalid-json.json", cmdLineArgsDefault, nil,
		func(args []string) (int, error) {
			workflowExecuted = true
			return 2, fmt.Errorf("something went wrong")
		})

	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errStr, "Failed to unmarshal STDIN data")
	assert.False(t, workflowExecuted)
}

func getFileReader(file string) io.Reader {
	reader, _ := os.Open(file)
	return reader
}

func clearEnvironment() {
	_ = os.Unsetenv("STR_ENV")
	_ = os.Unsetenv("UINT64_ENV")
	_ = os.Unsetenv("UINT32_ENV")
	_ = os.Unsetenv("UINT32_ENV")
	_ = os.Unsetenv("UINT16_ENV")
	_ = os.Unsetenv("INT64_ENV")
	_ = os.Unsetenv("INT32_ENV")
	_ = os.Unsetenv("INT16_ENV")
	_ = os.Unsetenv("BOOL_ENV")
}

func setEnvironment() {
	_ = os.Setenv("UINT64_ENV", "98989")
	_ = os.Setenv("UINT32_ENV", "87878")
	_ = os.Setenv("UINT16_ENV", "12121")
	_ = os.Setenv("INT64_ENV", "-98989")
	_ = os.Setenv("INT32_ENV", "-87878")
	_ = os.Setenv("INT16_ENV", "-2121")
	_ = os.Setenv("STR_ENV", "env str")
	_ = os.Setenv("BOOL_ENV", "true")
}

func goPluginExecuteUtil(t *testing.T, eventFile string, cmdLineArgs []string, expectedValues *allTestValues,
	workflowFunction func([]string) (int, error)) (int, string) {
	values := allTestValues{}
	options := getTestOptions(&values)
	var eventReader io.Reader
	if len(eventFile) > 0 {
		eventReader = getFileReader(eventFile)
	}

	basePlugin := newTestGoPlugin(options, eventReader, workflowFunction)

	// Simulate the command line arguments if necessary
	if len(cmdLineArgs) > 0 {
		basePlugin.cmdArgs.SetArgs(cmdLineArgs)
	} else {
		basePlugin.cmdArgs.SetArgs([]string{})
	}

	// Replace stdin reader with file reader and exitFunction with our own so we can know the exit status
	var exitStatus int
	var errorStr = ""
	basePlugin.eventReader = getFileReader(eventFile)
	basePlugin.exitFunction = func(i int) {
		exitStatus = i
	}
	basePlugin.errorLogFunction = func(format string, a ...interface{}) {
		errorStr = fmt.Sprintf(format, a...)
	}
	basePlugin.Execute()

	if expectedValues != nil {
		assert.Equal(t, expectedValues.boolValue, values.boolValue)
		assert.Equal(t, expectedValues.uint64Value, values.uint64Value)
		assert.Equal(t, expectedValues.uint32Value, values.uint32Value)
		assert.Equal(t, expectedValues.uint16Value, values.uint16Value)
		assert.Equal(t, expectedValues.int64Value, values.int64Value)
		assert.Equal(t, expectedValues.int32Value, values.int32Value)
		assert.Equal(t, expectedValues.int16Value, values.int16Value)
		assert.Equal(t, expectedValues.strValue, values.strValue)
	}

	return exitStatus, errorStr
}

func newTestGoPlugin(options []*PluginConfigOption, eventReader io.Reader, workflowFunction func([]string) (int, error)) *basePlugin {
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
	goPlugin.initPlugin()

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
