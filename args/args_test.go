package args

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

type argumentValues struct {
	stringArg  string
	uInt64Arg  uint64
	booleanArg bool
}

const (
	stringArg        = "string argument"
	uint64Arg uint64 = 123456789
	boolArg          = true

	stringEnvVar = "ENV_STR"
	uint64EnvVar = "ENV_UINT64"
	boolEnvVar   = "ENV_BOOL"

	defaultStringArg        = "default str"
	defaultUint64Arg uint64 = 343466773
	defaultBoolArg          = true
)

// TestNewArgs makes sure the args object is initialized correctly
func TestArgs_NewArgs(t *testing.T) {
	arguments := NewArgs("use", "short", func(_ []string) error {
		return nil
	})
	assert.NotNil(t, arguments)
	assert.NotNil(t, arguments.runE)
	assert.NotNil(t, arguments.cmd)
	assert.Equal(t, "use", arguments.cmd.Use)
	assert.Equal(t, "short", arguments.cmd.Short)
}

// Test short command-line arguments
func TestArgs_ExecuteShort(t *testing.T) {
	functionExecuted := false
	argValues := &argumentValues{}
	ClearEnvironment()

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{
		"-s", stringArg,
		"-i", strconv.FormatUint(uint64Arg, 10),
		"-b", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, boolArg, argValues.booleanArg)
	assert.Nil(t, err)
	assert.True(t, functionExecuted)
}

// Tests full command-line arguments
func TestArgs_ExecuteFull(t *testing.T) {
	functionExecuted := false
	argValues := &argumentValues{}
	ClearEnvironment()

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{
		"--str", stringArg,
		"--uint64", strconv.FormatUint(uint64Arg, 10),
		"--bool", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, boolArg, argValues.booleanArg)
	assert.Nil(t, err)
	assert.True(t, functionExecuted)
}

// Test environment variables
func TestArgs_ExecuteEnvironment(t *testing.T) {
	functionExecuted := false
	argValues := &argumentValues{}
	ClearEnvironment()

	_ = os.Setenv(stringEnvVar, stringArg)
	_ = os.Setenv(uint64EnvVar, strconv.FormatUint(uint64Arg, 10))
	_ = os.Setenv(boolEnvVar, strconv.FormatBool(boolArg))

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, boolArg, argValues.booleanArg)
	assert.Nil(t, err)
	assert.True(t, functionExecuted)
}

// Test environment variables
func TestArgs_ExecuteDefaultValues(t *testing.T) {
	functionExecuted := false
	argValues := &argumentValues{}
	ClearEnvironment()

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{})

	err := arguments.Execute()

	assert.Equal(t, defaultStringArg, argValues.stringArg)
	assert.Equal(t, defaultUint64Arg, argValues.uInt64Arg)
	assert.Equal(t, defaultBoolArg, argValues.booleanArg)
	assert.Nil(t, err)
	assert.True(t, functionExecuted)
}

// Test environment variables and command line arguments, making sure the command
// line arguments have priority
func TestArgs_ExecuteArgsAndEnvironment(t *testing.T) {
	functionExecuted := false
	argValues := &argumentValues{}
	ClearEnvironment()

	_ = os.Setenv(stringEnvVar, "env"+stringArg)
	_ = os.Setenv(uint64EnvVar, "env"+strconv.FormatUint(uint64Arg, 10))
	_ = os.Setenv(boolEnvVar, "env"+strconv.FormatBool(boolArg))

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{
		"--str", stringArg,
		"--uint64", strconv.FormatUint(uint64Arg, 10),
		"--bool", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, boolArg, argValues.booleanArg)
	assert.Nil(t, err)
	assert.True(t, functionExecuted)
}

// Test error
func TestArgs_ExecuteError(t *testing.T) {
	functionExecuted := false
	argValues := &argumentValues{}
	ClearEnvironment()

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return fmt.Errorf("test error")
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{
		"-s", stringArg,
		"-i", strconv.FormatUint(uint64Arg, 10),
		"-b", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, boolArg, argValues.booleanArg)
	assert.True(t, functionExecuted)
	assert.NotNil(t, err)
	assert.Equal(t, "test error", err.Error())
}

// TestHelp makes sure the help command-line argument is set
func TestArgs_Help(t *testing.T) {
	argValues := &argumentValues{}
	ClearEnvironment()

	arguments := NewArgs("use", "short", func(strings []string) error {
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{
		"-s", stringArg,
		"-i", strconv.FormatUint(uint64Arg, 10),
		"-b", strconv.FormatBool(boolArg),
	})

	err := arguments.Help()

	assert.Nil(t, err)
}

func setupArgs(arguments *Args, argValues *argumentValues) {
	arguments.StringVarP(&argValues.stringArg, "str", "s", "ENV_STR", defaultStringArg, "Use str")
	arguments.Uint64VarP(&argValues.uInt64Arg, "uint64", "i", "ENV_UINT64", defaultUint64Arg, "Use uint64")
	arguments.BoolVarP(&argValues.booleanArg, "bool", "b", "ENV_BOOL", defaultBoolArg, "Use bool")
}

func ClearEnvironment() {
	_ = os.Unsetenv(stringEnvVar)
	_ = os.Unsetenv(uint64EnvVar)
	_ = os.Unsetenv(boolEnvVar)
}
