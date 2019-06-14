package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewGoHandler(t *testing.T) {
	values := &allTestValues{}
	options := getTestOptions(values)
	goPlugin := NewGoHandler(defaultPluginConfig, options, func(event *types.Event) error {
		return nil
	}, func(event *types.Event) error {
		return nil
	})

	goHandler := goPlugin.(*goHandler)

	assert.NotNil(t, goHandler)
	assert.NotNil(t, goHandler.options)
	assert.Equal(t, options, goHandler.options)
	assert.NotNil(t, goHandler.config)
	assert.Equal(t, defaultPluginConfig, goHandler.config)
	assert.NotNil(t, goHandler.validationFunction)
	assert.NotNil(t, goHandler.executeFunction)
	assert.Nil(t, goHandler.sensuEvent)
	assert.Equal(t, os.Stdin, goHandler.eventReader)
}

func TestNewGoHandler_NoOptionValue(t *testing.T) {
	var exitStatus int
	options := getTestOptions(nil)
	handlerConfig := defaultPluginConfig

	goPlugin := NewGoHandler(handlerConfig, options,
		func(event *types.Event) error {
			return nil
		}, func(event *types.Event) error {
			return nil
		})

	assert.NotNil(t, goPlugin)

	goHandler := goPlugin.(*goHandler)
	goHandler.exitFunction = func(i int) {
		exitStatus = i
	}
	goHandler.Execute()
	assert.Equal(t, 1, exitStatus)
}

func goHandlerExecuteUtil(t *testing.T, handlerConfig *PluginConfig, eventFile string, cmdLineArgs []string,
	expectedValues *allTestValues, validationFunction func(*types.Event) error,
	executeFunction func(*types.Event) error) (int, string) {
	values := allTestValues{}
	options := getTestOptions(&values)

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

func TestGoHandler_Execute_Ok(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errStr := goHandlerExecuteUtil(t, defaultPluginConfig, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
			executeCalled = true
			return nil
		})
	assert.Equal(t, 0, exitStatus)
	assert.Equal(t, "", errStr)
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}

// Test validation error
func TestGoHandler_Execute_ValidationError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, defaultPluginConfig, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return fmt.Errorf("validation error")
		}, func(event *types.Event) error {
			executeCalled = true
			return nil
		})
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "error validating input: validation error")
	assert.True(t, validateCalled)
	assert.False(t, executeCalled)
}

// Test execute error
func TestGoHandler_Execute_ExecuteError(t *testing.T) {
	var validateCalled, executeCalled bool
	clearEnvironment()
	exitStatus, errorStr := goHandlerExecuteUtil(t, defaultPluginConfig, "test/event-no-override.json", cmdLineArgsDefault,
		&expectedCmdLineValues,
		func(event *types.Event) error {
			validateCalled = true
			assert.NotNil(t, event)
			return nil
		}, func(event *types.Event) error {
			executeCalled = true
			assert.NotNil(t, event)
			return fmt.Errorf("execution error")
		})
	assert.Equal(t, 1, exitStatus)
	assert.Contains(t, errorStr, "error executing handler: execution error")
	assert.True(t, validateCalled)
	assert.True(t, executeCalled)
}
