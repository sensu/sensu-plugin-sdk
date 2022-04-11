package sensu

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

type handlerValues struct {
	arg1 string
	arg2 uint64
	arg3 bool
}

var (
	defaultHandlerConfig = PluginConfig{
		Name:     "TestHandler",
		Short:    "Short Description",
		Timeout:  10,
		Keyspace: "sensu.io/plugins/segp/config",
	}

	stringOpt = PluginConfigOption[string]{
		Argument:  "string",
		Default:   "Default1",
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
	}

	uint64Opt = PluginConfigOption[uint64]{
		Argument:  "uint64",
		Default:   uint64(33333),
		Env:       "ENV_2",
		Path:      "path2",
		Shorthand: "e",
		Usage:     "Second argument",
	}

	boolOpt = PluginConfigOption[bool]{
		Argument:  "bool",
		Default:   false,
		Env:       "ENV_3",
		Path:      "path3",
		Shorthand: "f",
		Usage:     "Third argument",
	}

	stringSliceOpt = SlicePluginConfigOption[string]{
		Argument:  "stringslice",
		Default:   []string{"Default1"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
	}

	stringMapOpt = MapPluginConfigOption[string]{
		Argument:  "stringslice",
		Default:   map[string]string{"default": "yes"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
	}

	intSliceOpt = SlicePluginConfigOption[int]{
		Argument:  "intslice",
		Default:   []int{0},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
	}

	int64Opt = PluginConfigOption[int64]{
		Argument:  "int64",
		Default:   int64(33333),
		Env:       "ENV_2",
		Path:      "path2",
		Shorthand: "e",
		Usage:     "Second argument",
	}

	defaultCmdLineArgs = []string{"--string", "value-arg1", "--uint64", "7531", "--bool=false"}
)

func TestNewGoHandler(t *testing.T) {
	values := &handlerValues{}
	options := getHandlerOptions(values)
	goHandler := NewGoHandler(&defaultHandlerConfig, options, func(event *corev2.Event) error {
		return nil
	}, func(event *corev2.Event) error {
		return nil
	})

	assert.NotNil(t, goHandler)
	assert.NotNil(t, goHandler.framework.options)
	assert.Equal(t, options, goHandler.framework.options)
	assert.NotNil(t, goHandler.framework.config)
	assert.Equal(t, &defaultHandlerConfig, goHandler.framework.config)
	assert.NotNil(t, goHandler.validationFunction)
	assert.NotNil(t, goHandler.executeFunction)
	assert.Nil(t, goHandler.framework.GetStdinEvent())
	assert.Equal(t, os.Stdin, goHandler.framework.eventReader)
}

func TestNewGoHandler_NoOptionValue(t *testing.T) {
	var exitStatus int
	options := getHandlerOptions(nil)
	handlerConfig := defaultHandlerConfig

	goHandler := NewGoHandler(&handlerConfig, options,
		func(event *corev2.Event) error {
			return nil
		}, func(event *corev2.Event) error {
			return nil
		})

	goHandler.framework.exitFunction = func(i int) {
		exitStatus = i
	}
	goHandler.Execute()
	assert.Equal(t, 1, exitStatus)
}

func goHandlerExecuteUtil(t *testing.T, handlerConfig *PluginConfig, eventFile string, cmdLineArgs []string,
	validationFunction func(*corev2.Event) error, executeFunction func(*corev2.Event) error,
	expectedValue1 string, expectedValue2 uint64, expectedValue3 bool) (int, string) {

	t.Helper()
	values := handlerValues{}
	options := getHandlerOptions(&values)

	goHandler := NewGoHandler(handlerConfig, options, validationFunction, executeFunction)

	// Simulate the command line arguments if necessary
	if len(cmdLineArgs) > 0 {
		goHandler.framework.cmd.SetArgs(cmdLineArgs)
	} else {
		goHandler.framework.cmd.SetArgs([]string{})
	}

	goHandler.framework.cmd.SilenceErrors = true
	goHandler.framework.cmd.SilenceUsage = true

	// Replace stdin reader with file reader and exitFunction with our own so we can know the exit status
	var exitStatus int
	var errorStr = ""
	goHandler.framework.eventReader = getFileReader(eventFile)
	goHandler.framework.exitFunction = func(i int) {
		exitStatus = i
	}
	goHandler.framework.errorLogFunction = func(format string, a ...interface{}) {
		errorStr = fmt.Sprintf(format, a...)
	}
	goHandler.Execute()

	if exitStatus == 0 {
		if expectedValue1 != values.arg1 {
			t.Errorf("%q != %q", expectedValue1, values.arg1)
		}
		if expectedValue2 != values.arg2 {
			t.Errorf("%v != %v", expectedValue2, values.arg2)
		}
		if expectedValue3 != values.arg3 {
			t.Errorf("%v != %v", expectedValue3, values.arg3)
		}
	}

	return exitStatus, errorStr
}

// Test check override
func TestGoHandler_Execute_Check(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-check-override.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-check1", uint64(1357), false)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test check override with invalid value
func TestGoHandler_Execute_CheckInvalidValue(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-check-override-invalid-value.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-check1", uint64(33333), false)
	assert.Equal(t, 1, exitStatus)
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test entity override
func TestGoHandler_Execute_Entity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-entity-override.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-entity1", uint64(2468), true)

	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test entity override - invalid value
func TestGoHandler_Execute_EntityInvalidValue(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-entity-override-invalid-value.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-entity1", uint64(33333), false)

	assert.Equal(t, 1, exitStatus)
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test environment
func TestGoHandler_Execute_Environment(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-env1", uint64(9753), true)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test cmd line arguments
func TestGoHandler_Execute_CmdLineArgs(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test check priority - check override
func TestGoHandler_Execute_PriorityCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-check-entity-override.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-check1", uint64(1357), false)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test next priority - entity override
func TestGoHandler_Execute_PriorityEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-entity-override.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-entity1", uint64(2468), true)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test next priority - cmd line arguments
func TestGoHandler_Execute_PriorityCmdLine(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test validation error
func TestGoHandler_Execute_ValidationError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return fmt.Errorf("validation error")
		}, func(event *corev2.Event) error {
			executeCalled = true
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "error validating input: validation error")
	assert.True(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test execute error
func TestGoHandler_Execute_ExecuteError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return fmt.Errorf("execution error")
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "error executing handler: execution error")
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test invalid event - no timestamp
func TestGoHandler_Execute_EventNoTimestamp(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-timestamp.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "timestamp is missing or must be greater than zero")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - timestamp 0
func TestGoHandler_Execute_EventTimestampZero(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-timestamp-zero.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "timestamp is missing or must be greater than zero")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - no entity
func TestGoHandler_Execute_EventNoEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-entity.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "event must contain an entity")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - invalid entity
func TestGoHandler_Execute_EventInvalidEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-entity.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "entity name must not be empty")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - no check
func TestGoHandler_Execute_EventNoCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-check.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "event must contain a check or metrics")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - invalid check
func TestGoHandler_Execute_EventInvalidCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-check.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "check name must not be empty")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test unmarshalling error
func TestGoHandler_Execute_EventInvalidJson(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-json.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "failed to unmarshal")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test fail to read stdin
func TestGoHandler_Execute_ReaderError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-json.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "failed to unmarshal")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test no keyspace
func TestGoHandler_Execute_NoKeyspace(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	handlerConfig := defaultHandlerConfig
	handlerConfig.Keyspace = ""
	exitStatus, _ := goHandlerExecuteUtil(t, &handlerConfig, "test/event-check-entity-override.json", defaultCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return nil
		},
		"value-arg1", uint64(7531), false)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

func getHandlerOptions(values *handlerValues) []ConfigOption {
	option1 := stringOpt
	option2 := uint64Opt
	option3 := boolOpt
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

func TestNewGoHandlerEnterprise(t *testing.T) {
	var exitStatus int
	values := &handlerValues{}
	options := getHandlerOptions(values)
	goHandler := NewEnterpriseGoHandler(&defaultHandlerConfig, options, func(event *corev2.Event) error {
		return nil
	}, func(event *corev2.Event) error {
		return nil
	})
	assert.True(t, goHandler.enterprise)

	goHandler.framework.exitFunction = func(i int) {
		exitStatus = i
	}
	goHandler.Execute()
	assert.Equal(t, 1, exitStatus)
}
