package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

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
		Default:   uint64(33333),
		Env:       "ENV_2",
		Path:      "path2",
		Shorthand: "e",
		Usage:     "Second argument",
	}

	defaultOption3 = PluginConfigOption{
		Argument:  "arg3",
		Default:   false,
		Env:       "ENV_3",
		Path:      "path3",
		Shorthand: "f",
		Usage:     "Third argument",
	}

	defaultCmdLineArgs = []string{"--arg1", "value-arg1", "--arg2", "7531", "--arg3=false"}
)

func TestInitHandler(t *testing.T) {
	values := &handlerValues{}
	options := getHandlerOptions(values)
	Handler := InitHandler(&defaultHandlerConfig, options, func(event *types.Event) error {
		return nil
	}, func(event *types.Event) error {
		return nil
	})

	assert.NotNil(t, Handler)
	assert.NotNil(t, Handler.options)
	assert.Equal(t, options, Handler.options)
	assert.NotNil(t, Handler.config)
	assert.Equal(t, &defaultHandlerConfig, Handler.config)
	assert.NotNil(t, Handler.validationFunction)
	assert.NotNil(t, Handler.executeFunction)
	assert.Nil(t, Handler.sensuEvent)
	assert.Equal(t, os.Stdin, Handler.eventReader)
}

func TestInitHandler_NoOptionValue(t *testing.T) {
	var exitStatus int
	options := getHandlerOptions(nil)
	handlerConfig := defaultHandlerConfig

	Handler := InitHandler(&handlerConfig, options,
		func(event *types.Event) error {
			return nil
		}, func(event *types.Event) error {
			return nil
		})

	Handler.exitFunction = func(i int) {
		exitStatus = i
	}
	Handler.Execute()
	assert.Equal(t, 1, exitStatus)
}

func HandlerExecuteUtil(t *testing.T, handlerConfig *PluginConfig, eventFile string, cmdLineArgs []string,
	validationFunction func(*types.Event) error, executeFunction func(*types.Event) error,
	expectedValue1 interface{}, expectedValue2 interface{}, expectedValue3 interface{}) (int, string) {
	values := handlerValues{}
	options := getHandlerOptions(&values)

	Handler := InitHandler(handlerConfig, options, validationFunction, executeFunction)

	// Simulate the command line arguments if necessary
	if len(cmdLineArgs) > 0 {
		Handler.cmdArgs.SetArgs(cmdLineArgs)
	} else {
		Handler.cmdArgs.SetArgs([]string{})
	}

	// Replace stdin reader with file reader and exitFunction with our own so we can know the exit status
	var exitStatus int
	var errorStr = ""
	Handler.eventReader = getFileReader(eventFile)
	Handler.exitFunction = func(i int) {
		exitStatus = i
	}
	Handler.errorLogFunction = func(format string, a ...interface{}) {
		errorStr = fmt.Sprintf(format, a...)
	}
	Handler.Execute()

	assert.Equal(t, expectedValue1, values.arg1)
	assert.Equal(t, expectedValue2, values.arg2)
	assert.Equal(t, expectedValue3, values.arg3)

	return exitStatus, errorStr
}

// Test check override
func TestHandler_Execute_Check(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-check-override.json", nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_CheckInvalidValue(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-check-override-invalid-value.json", nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_Entity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-entity-override.json", nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_EntityInvalidValue(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-entity-override-invalid-value.json", nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_Environment(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_CmdLineArgs(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
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

// Test check priority - check override
func TestHandler_Execute_PriorityCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-check-entity-override.json", defaultCmdLineArgs,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_PriorityEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-entity-override.json", defaultCmdLineArgs,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_PriorityCmdLine(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, _ := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
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

// Test validation error
func TestHandler_Execute_ValidationError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return fmt.Errorf("validation error")
		}, func(event *types.Event) error {
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
func TestHandler_Execute_ExecuteError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-override.json", defaultCmdLineArgs,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
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
func TestHandler_Execute_EventNoTimestamp(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-timestamp.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "timestamp is missing or must be greater than zero")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - timestamp 0
func TestHandler_Execute_EventTimestampZero(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-timestamp-zero.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "timestamp is missing or must be greater than zero")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - no entity
func TestHandler_Execute_EventNoEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-entity.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "event must contain an entity")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - invalid entity
func TestHandler_Execute_EventInvalidEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-entity.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "entity name must not be empty")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - no check
func TestHandler_Execute_EventNoCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-no-check.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "event must contain a check or metrics")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - invalid check
func TestHandler_Execute_EventInvalidCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-check.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "check name must not be empty")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test unmarshalling error
func TestHandler_Execute_EventInvalidJson(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-json.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "Failed to unmarshal STDIN data: invalid character ':' after object key:value pair")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test fail to read stdin
func TestHandler_Execute_ReaderError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := HandlerExecuteUtil(t, &defaultHandlerConfig, "test/event-invalid-json.json", defaultCmdLineArgs,
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
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "Failed to unmarshal STDIN data: invalid character ':' after object key:value pair")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test no keyspace
func TestHandler_Execute_NoKeyspace(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	handlerConfig := defaultHandlerConfig
	handlerConfig.Keyspace = ""
	exitStatus, _ := HandlerExecuteUtil(t, &handlerConfig, "test/event-check-entity-override.json", defaultCmdLineArgs,
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

func getHandlerOptions(values *handlerValues) []*PluginConfigOption {
	option1 := defaultOption1
	option2 := defaultOption2
	option3 := defaultOption3
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
