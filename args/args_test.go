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
	uInt32Arg  uint32
	uInt16Arg  uint16
	int64Arg   int64
	int32Arg   int32
	int16Arg   int16
	booleanArg bool
}

const (
	stringArg        = "string argument"
	uint64Arg uint64 = 18446744073709551615
	uint32Arg uint32 = 4294967295
	uint16Arg uint16 = 65535
	int64Arg  int64  = -9223372036854775808
	int32Arg  int32  = -2147483648
	int16Arg  int16  = -32768
	boolArg          = true

	stringEnvVar = "ENV_STR"
	uint64EnvVar = "ENV_UINT64"
	uint32EnvVar = "ENV_UINT32"
	uint16EnvVar = "ENV_UINT16"
	int64EnvVar  = "ENV_INT64"
	int32EnvVar  = "ENV_INT32"
	int16EnvVar  = "ENV_INT16"
	boolEnvVar   = "ENV_BOOL"

	defaultStringArg string = "default str"
	defaultUint64Arg uint64 = 6744073709551615
	defaultUint32Arg uint32 = 42949672
	defaultUint16Arg uint16 = 6553
	defaultInt64Arg  int64  = 6744073709551615
	defaultInt32Arg  int32  = 42949672
	defaultInt16Arg  int16  = 6553
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
		"-j", strconv.FormatUint(uint64(uint32Arg), 10),
		"-k", strconv.FormatUint(uint64(uint16Arg), 10),
		"-l", strconv.FormatInt(int64Arg, 10),
		"-m", strconv.FormatInt(int64(int32Arg), 10),
		"-n", strconv.FormatInt(int64(int16Arg), 10),
		"-b", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
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
		"--uint32", strconv.FormatUint(uint64(uint32Arg), 10),
		"--uint16", strconv.FormatUint(uint64(uint16Arg), 10),
		"--int64", strconv.FormatInt(int64Arg, 10),
		"--int32", strconv.FormatInt(int64(int32Arg), 10),
		"--int16", strconv.FormatInt(int64(int16Arg), 10),
		"--bool", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
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
	_ = os.Setenv(uint32EnvVar, strconv.FormatUint(uint64(uint32Arg), 10))
	_ = os.Setenv(uint16EnvVar, strconv.FormatUint(uint64(uint16Arg), 10))
	_ = os.Setenv(int64EnvVar, strconv.FormatInt(int64Arg, 10))
	_ = os.Setenv(int32EnvVar, strconv.FormatInt(int64(int32Arg), 10))
	_ = os.Setenv(int16EnvVar, strconv.FormatInt(int64(int16Arg), 10))
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
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
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
	assert.Equal(t, defaultUint32Arg, argValues.uInt32Arg)
	assert.Equal(t, defaultUint16Arg, argValues.uInt16Arg)
	assert.Equal(t, defaultInt64Arg, argValues.int64Arg)
	assert.Equal(t, defaultInt32Arg, argValues.int32Arg)
	assert.Equal(t, defaultInt16Arg, argValues.int16Arg)
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
	_ = os.Setenv(uint32EnvVar, "env"+strconv.FormatUint(uint64(uint32Arg), 10))
	_ = os.Setenv(uint16EnvVar, "env"+strconv.FormatUint(uint64(uint16Arg), 10))
	_ = os.Setenv(int64EnvVar, "env"+strconv.FormatInt(int64Arg, 10))
	_ = os.Setenv(int32EnvVar, "env"+strconv.FormatInt(int64(int32Arg), 10))
	_ = os.Setenv(int16EnvVar, "env"+strconv.FormatInt(int64(int16Arg), 10))
	_ = os.Setenv(boolEnvVar, "env"+strconv.FormatBool(boolArg))

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{
		"--str", stringArg,
		"--uint64", strconv.FormatUint(uint64Arg, 10),
		"--uint32", strconv.FormatUint(uint64(uint32Arg), 10),
		"--uint16", strconv.FormatUint(uint64(uint16Arg), 10),
		"--int64", strconv.FormatInt(int64Arg, 10),
		"--int32", strconv.FormatInt(int64(int32Arg), 10),
		"--int16", strconv.FormatInt(int64(int16Arg), 10),
		"--bool", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
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
		"-j", strconv.FormatUint(uint64(uint32Arg), 10),
		"-k", strconv.FormatUint(uint64(uint16Arg), 10),
		"-l", strconv.FormatInt(int64Arg, 10),
		"-m", strconv.FormatInt(int64(int32Arg), 10),
		"-n", strconv.FormatInt(int64(int16Arg), 10),
		"-b", strconv.FormatBool(boolArg),
	})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
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
		"-j", strconv.FormatUint(uint64(uint32Arg), 10),
		"-k", strconv.FormatUint(uint64(uint16Arg), 10),
		"-l", strconv.FormatInt(int64Arg, 10),
		"-m", strconv.FormatInt(int64(uint32Arg), 10),
		"-n", strconv.FormatInt(int64(uint16Arg), 10),
		"-b", strconv.FormatBool(boolArg),
	})

	err := arguments.Help()

	assert.Nil(t, err)
}

func setupArgs(arguments *Args, argValues *argumentValues) {
	arguments.StringVarP(&argValues.stringArg, "str", "s", "ENV_STR", defaultStringArg, "Use str")
	arguments.Uint64VarP(&argValues.uInt64Arg, "uint64", "i", uint64EnvVar, defaultUint64Arg, "Use uint64")
	arguments.Uint32VarP(&argValues.uInt32Arg, "uint32", "j", uint32EnvVar, defaultUint32Arg, "Use uint32")
	arguments.Uint16VarP(&argValues.uInt16Arg, "uint16", "k", uint16EnvVar, defaultUint16Arg, "Use uint16")
	arguments.Int64VarP(&argValues.int64Arg, "int64", "l", int64EnvVar, defaultInt64Arg, "Use int64")
	arguments.Int32VarP(&argValues.int32Arg, "int32", "m", int32EnvVar, defaultInt32Arg, "Use int32")
	arguments.Int16VarP(&argValues.int16Arg, "int16", "n", int16EnvVar, defaultInt16Arg, "Use int16")
	arguments.BoolVarP(&argValues.booleanArg, "bool", "b", "ENV_BOOL", defaultBoolArg, "Use bool")
}

func ClearEnvironment() {
	_ = os.Unsetenv(stringEnvVar)
	_ = os.Unsetenv(uint64EnvVar)
	_ = os.Unsetenv(uint32EnvVar)
	_ = os.Unsetenv(uint16EnvVar)
	_ = os.Unsetenv(int64EnvVar)
	_ = os.Unsetenv(int32EnvVar)
	_ = os.Unsetenv(int16EnvVar)
	_ = os.Unsetenv(boolEnvVar)
}
