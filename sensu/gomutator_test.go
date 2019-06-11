package sensu

import (
	"bytes"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestNewGoMutator(t *testing.T) {
	values := &allTestValues{}
	options := getTestOptions(values)
	goPlugin := NewGoMutator(defaultPluginConfig, options, func(event *types.Event) error {
		return nil
	}, func(event *types.Event) (*types.Event, error) {
		return nil, nil
	})

	goMutator := goPlugin.(*goMutator)

	assert.NotNil(t, goMutator)
	assert.NotNil(t, goMutator.options)
	assert.Equal(t, options, goMutator.options)
	assert.NotNil(t, goMutator.config)
	assert.Equal(t, defaultPluginConfig, goMutator.config)
	assert.NotNil(t, goMutator.validationFunction)
	assert.NotNil(t, goMutator.executeFunction)
	assert.Nil(t, goMutator.sensuEvent)
	assert.Equal(t, os.Stdin, goMutator.eventReader)
	assert.NotNil(t, goMutator.pluginWorkflowFunction)
	assert.NotNil(t, goMutator.cmdArgs)
}

func TestNewGoMutator_NoOptionValue(t *testing.T) {
	options := getTestOptions(nil)
	mutatorConfig := *defaultPluginConfig

	goPlugin := NewGoMutator(&mutatorConfig, options,
		func(event *types.Event) error {
			return nil
		}, func(event *types.Event) (*types.Event, error) {
			return nil, nil
		})

	assert.NotNil(t, goPlugin)
	goMutator := goPlugin.(*goMutator)

	var exitStatus = -99
	goMutator.exitFunction = func(i int) {
		exitStatus = i
	}
	goMutator.Execute()

	assert.Equal(t, 1, exitStatus)
}

func goMutatorExecuteUtil(t *testing.T, mutatorConfig *PluginConfig, eventFile string, cmdLineArgs []string,
	expectedValues *allTestValues, writer io.Writer,
	validationFunction func(*types.Event) error, executeFunction func(*types.Event) (*types.Event, error)) (int, string) {
	values := allTestValues{}
	options := getTestOptions(&values)

	goPlugin := NewGoMutator(mutatorConfig, options, validationFunction, executeFunction)
	goMutator := goPlugin.(*goMutator)
	if writer != nil {
		goMutator.out = writer
	}

	if len(cmdLineArgs) > 0 {
		goMutator.cmdArgs.SetArgs(cmdLineArgs)
	} else {
		goMutator.cmdArgs.SetArgs([]string{})
	}

	// Replace stdin reader with file reader
	var exitStatus = -99
	var errorStr = ""
	goMutator.eventReader = getFileReader(eventFile)
	goMutator.exitFunction = func(i int) {
		exitStatus = i
	}
	goMutator.errorLogFunction = func(format string, a ...interface{}) {
		errorStr = fmt.Sprintf(format, a...)
	}
	goMutator.Execute()

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

// Test check override
func TestGoMutator_Execute_Check_NilEvent(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	var writer io.Writer = new(bytes.Buffer)
	exitStatus, err := goMutatorExecuteUtil(t, defaultPluginConfig, "test/event-check-override.json", nil,
		nil, writer,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) (*types.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		})
	assert.Equal(t, "", err)
	assert.Equal(t, 0, exitStatus)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)

	output := writer.(*bytes.Buffer).String()
	fmt.Printf("Output: %s", output)
	assert.Equal(t, output, "{}")
}

// Test validation error
func TestGoMutator_Execute_ValidationError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, defaultPluginConfig, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues, nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return fmt.Errorf("validation error")
		}, func(event *types.Event) (*types.Event, error) {
			executeCalled = true
			return nil, nil
		})
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "error validating input: validation error")
	assert.True(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test execute error
func TestGoMutator_Execute_ExecuteError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, defaultPluginConfig, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues, nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) (*types.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, fmt.Errorf("execution error")
		})
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "error executing mutator: execution error")
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test fail to read stdin
func TestGoMutator_Execute_ReaderError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, err := goMutatorExecuteUtil(t, defaultPluginConfig, "test/event-invalid-json.json", cmdLineArgsDefault,
		&expectedCmdLineValues, nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) (*types.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		})
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, err, "Failed to unmarshal STDIN data: invalid character ':' after object key:value pair")
	assert.False(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test no keyspace
func TestGoMutator_Execute_NoKeyspace(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	mutatorConfig := *defaultPluginConfig
	mutatorConfig.Keyspace = ""
	exitStatus, err := goMutatorExecuteUtil(t, &mutatorConfig, "test/event-check-entity-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues, nil,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) (*types.Event, error) {
			executeCalled = true
			assert.NotNil(t, event)
			return nil, nil
		})
	assert.Equal(t, 0, exitStatus)
	assert.Equal(t, "", err)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}
