package sensu

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
)

type mutatorValues struct {
	arg1 string
	arg2 uint64
	arg3 bool
}

var (
	defaultMutatorConfig = PluginConfig{
		Name:     "TestMutator",
		Short:    "Short Description",
		Timeout:  10,
		Keyspace: "sensu.io/plugins/segp/config",
	}

	mutatorOption1 = PluginConfigOption[string]{
		Argument:  "arg1",
		Default:   "Default1",
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
	}

	mutatorOption2 = PluginConfigOption[uint64]{
		Argument:  "arg2",
		Default:   uint64(33333),
		Env:       "ENV_2",
		Path:      "path2",
		Shorthand: "e",
		Usage:     "Second argument",
	}

	mutatorOption3 = PluginConfigOption[bool]{
		Argument:  "arg3",
		Default:   false,
		Env:       "ENV_3",
		Path:      "path3",
		Shorthand: "f",
		Usage:     "Third argument",
	}

	mutatorCmdLineArgs = []string{"--arg1", "value-arg1", "--arg2", "7531", "--arg3=false"}
)

func TestNewGoMutator(t *testing.T) {
	values := &mutatorValues{}
	options := getMutatorVales(values)
	goMutator := NewGoMutator(&defaultMutatorConfig, options, func(event *corev2.Event) error {
		return nil
	}, func(event *corev2.Event) (*corev2.Event, error) {
		return nil, nil
	})

	assert.NotNil(t, goMutator)
	assert.NotNil(t, goMutator.framework.options)
	assert.Equal(t, options, goMutator.framework.options)
	assert.NotNil(t, goMutator.framework.config)
	assert.Equal(t, &defaultMutatorConfig, goMutator.framework.config)
	assert.NotNil(t, goMutator.validationFunction)
	assert.NotNil(t, goMutator.executeFunction)
	assert.Nil(t, goMutator.framework.GetStdinEvent())
	assert.Equal(t, os.Stdin, goMutator.framework.eventReader)
	assert.NotNil(t, goMutator.framework.pluginWorkflowFunction)
	assert.NotNil(t, goMutator.framework.cmd)
}

func TestNewGoMutator_NoOptionValue(t *testing.T) {
	options := getMutatorVales(nil)
	mutatorConfig := defaultMutatorConfig

	goMutator := NewGoMutator(&mutatorConfig, options,
		func(event *corev2.Event) error {
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			return nil, nil
		})

	assert.NotNil(t, goMutator)

	var exitStatus = -99
	goMutator.framework.exitFunction = func(i int) {
		exitStatus = i
	}
	goMutator.Execute()

	assert.Equal(t, 1, exitStatus)
}

func goMutatorExecuteUtil(t *testing.T, mutatorConfig *PluginConfig, eventFile string, cmdLineArgs []string,
	validationFunction func(*corev2.Event) error, executeFunction func(*corev2.Event) (*corev2.Event, error),
	expectedValue1 string, expectedValue2 uint64, expectedValue3 bool, writer io.Writer) (int, string) {
	values := mutatorValues{}
	options := getMutatorVales(&values)

	goMutator := NewGoMutator(mutatorConfig, options, validationFunction, executeFunction)
	if writer != nil {
		goMutator.out = writer
	}

	if len(cmdLineArgs) > 0 {
		goMutator.framework.cmd.SetArgs(cmdLineArgs)
	} else {
		goMutator.framework.cmd.SetArgs([]string{})
	}

	goMutator.framework.cmd.SilenceErrors = true
	goMutator.framework.cmd.SilenceUsage = true

	// Replace stdin reader with file reader
	var exitStatus = -99
	var errorStr = ""
	goMutator.framework.eventReader = getFileReader(eventFile)
	goMutator.framework.exitFunction = func(i int) {
		exitStatus = i
	}
	goMutator.framework.errorLogFunction = func(format string, a ...interface{}) {
		errorStr = fmt.Sprintf(format, a...)
	}
	goMutator.Execute()

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
func TestGoMutator_Execute_Check(t *testing.T) {
	var validateCalled, executeCalled bool
	const newName = "Modified Name"
	clearEnvironment()
	var writer io.Writer = new(bytes.Buffer)
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-check-override.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			resultEvent := *event
			resultEvent.Check.Name = newName
			return &resultEvent, nil
		},
		"value-check1", uint64(1357), false, writer)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)

	output := writer.(*bytes.Buffer).String()
	assert.True(t, len(output) > 5)
	assert.True(t, strings.Contains(output, newName))
}

// Test check override
func TestGoMutator_Execute_Check_NilEvent(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	var writer io.Writer = new(bytes.Buffer)
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-check-override.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-check1", uint64(1357), false, writer)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)

	output := writer.(*bytes.Buffer).String()
	assert.Equal(t, output, "{}")
}

// Test check override with invalid value
func TestGoMutator_Execute_CheckInvalidValue(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-check-override-invalid-value.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-check1", uint64(33333), false, nil)
	assert.NotEqual(t, "", err)
	assert.Equal(t, 1, exitStatus)
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test entity override
func TestGoMutator_Execute_Entity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-entity-override.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-entity1", uint64(2468), true, nil)

	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test entity override - invalid value
func TestGoMutator_Execute_EntityInvalidValue(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-entity-override-invalid-value.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-entity1", uint64(33333), false, nil)

	assert.NotEqual(t, "", err)
	assert.Equal(t, 1, exitStatus)
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test environment
func TestGoMutator_Execute_Environment(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-override.json", nil,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-env1", uint64(9753), true, nil)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test cmd line arguments
func TestGoMutator_Execute_CmdLineArgs(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-override.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test check priority - check override
func TestGoMutator_Execute_PriorityCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-check-entity-override.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-check1", uint64(1357), false, nil)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test next priority - entity override
func TestGoMutator_Execute_PriorityEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-entity-override.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-entity1", uint64(2468), true, nil)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test next priority - cmd line arguments
func TestGoMutator_Execute_PriorityCmdLine(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	_ = os.Setenv("ENV_1", "value-env1")
	_ = os.Setenv("ENV_2", "9753")
	_ = os.Setenv("ENV_3", "true")
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-override.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test validation error
func TestGoMutator_Execute_ValidationError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-override.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return fmt.Errorf("validation error")
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "error validating input: validation error")
	assert.True(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test execute error
func TestGoMutator_Execute_ExecuteError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-override.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, fmt.Errorf("execution error")
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "error executing mutator: execution error")
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test invalid event - no timestamp
func TestGoMutator_Execute_EventNoTimestamp(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-timestamp.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "timestamp is missing or must be greater than zero")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - timestamp 0
func TestGoMutator_Execute_EventTimestampZero(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-timestamp-zero.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "timestamp is missing or must be greater than zero")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - no entity
func TestGoMutator_Execute_EventNoEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-entity.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "event must contain an entity")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - invalid entity
func TestGoMutator_Execute_EventInvalidEntity(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-invalid-entity.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "entity name must not be empty")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - no check
func TestGoMutator_Execute_EventNoCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-no-check.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "event must contain a check or metrics")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test invalid event - invalid check
func TestGoMutator_Execute_EventInvalidCheck(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-invalid-check.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "check name must not be empty")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test unmarshalling error
func TestGoMutator_Execute_EventInvalidJson(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-invalid-json.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "failed to unmarshal")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test fail to read stdin
func TestGoMutator_Execute_ReaderError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, &defaultMutatorConfig, "test/event-invalid-json.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "failed to unmarshal")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test no keyspace
func TestGoMutator_Execute_NoKeyspace(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	mutatorConfig := defaultMutatorConfig
	mutatorConfig.Keyspace = ""
	exitStatus, err := goMutatorExecuteUtil(t, &mutatorConfig, "test/event-check-entity-override.json", mutatorCmdLineArgs,
		func(event *corev2.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *corev2.Event) (*corev2.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		},
		"value-arg1", uint64(7531), false, nil)
	assert.Equal(t, 0, exitStatus)
	assert.Equal(t, "", err)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

func getMutatorVales(values *mutatorValues) []ConfigOption {
	option1 := mutatorOption1
	option2 := mutatorOption2
	option3 := mutatorOption3
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
