package args

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"reflect"
	"strconv"
	"testing"
)

type argumentValues struct {
	stringArg  string
	uintArg    uint
	uInt64Arg  uint64
	uInt32Arg  uint32
	uInt16Arg  uint16
	uInt8Arg   uint8
	intArg     int
	int64Arg   int64
	int32Arg   int32
	int16Arg   int16
	int8Arg    int8
	booleanArg bool
}

type readEnvTestData struct {
	envKind       reflect.Kind
	envValue      string
	expectedValue interface{}
	expectedError bool
}

var (
	envTestDataRecords = []readEnvTestData{
		{reflect.Int8, "0", int8(0), false},
		{reflect.Int8, "127", int8(127), false},
		{reflect.Int8, "-128", int8(-128), false},
		{reflect.Int8, "-44", int8(-44), false},
		{reflect.Int8, "44", int8(44), false},
		{reflect.Int8, "-129", nil, true},
		{reflect.Int8, "128", nil, true},
		{reflect.Int8, "-21474836499999", nil, true},
		{reflect.Int8, "21474836489999", nil, true},
		{reflect.Int8, "", nil, true},
		{reflect.Int8, "abcde", nil, true},

		{reflect.Int16, "0", int16(0), false},
		{reflect.Int16, "12345", int16(12345), false},
		{reflect.Int16, "-12345", int16(-12345), false},
		{reflect.Int16, "-32768", int16(-32768), false},
		{reflect.Int16, "32767", int16(32767), false},
		{reflect.Int16, "-32769", nil, true},
		{reflect.Int16, "32768", nil, true},
		{reflect.Int16, "-21474836499999", nil, true},
		{reflect.Int16, "21474836489999", nil, true},
		{reflect.Int16, "", nil, true},
		{reflect.Int16, "abcde", nil, true},

		{reflect.Int32, "0", int32(0), false},
		{reflect.Int32, "12345", int32(12345), false},
		{reflect.Int32, "-12345", int32(-12345), false},
		{reflect.Int32, "-2147483648", int32(-2147483648), false},
		{reflect.Int32, "2147483647", int32(2147483647), false},
		{reflect.Int32, "-2147483649", nil, true},
		{reflect.Int32, "2147483648", nil, true},
		{reflect.Int32, "-21474836499999", nil, true},
		{reflect.Int32, "21474836489999", nil, true},
		{reflect.Int32, "", nil, true},
		{reflect.Int32, "abcde", nil, true},

		{reflect.Int64, "0", int64(0), false},
		{reflect.Int64, "12345", int64(12345), false},
		{reflect.Int64, "-12345", int64(-12345), false},
		{reflect.Int64, "-9223372036854775808", int64(-9223372036854775808), false},
		{reflect.Int64, "9223372036854775807", int64(9223372036854775807), false},
		{reflect.Int64, "-9223372036854775809", nil, true},
		{reflect.Int64, "9223372036854775808", nil, true},
		{reflect.Int64, "-21474839999996499999", nil, true},
		{reflect.Int64, "214748364999999989999", nil, true},
		{reflect.Int64, "", nil, true},
		{reflect.Int64, "abcde", nil, true},

		{reflect.Int, "0", 0, false},
		{reflect.Int, "12345", 12345, false},
		{reflect.Int, "-12345", -12345, false},
		{reflect.Int, "-9223372036854775808", -9223372036854775808, false},
		{reflect.Int, "9223372036854775807", 9223372036854775807, false},
		{reflect.Int, "-9223372036854775809", nil, true},
		{reflect.Int, "9223372036854775808", nil, true},
		{reflect.Int, "-21474839999996499999", nil, true},
		{reflect.Int, "214748364999999989999", nil, true},
		{reflect.Int, "", nil, true},
		{reflect.Int, "abcde", nil, true},

		{reflect.Uint8, "0", uint8(0), false},
		{reflect.Uint8, "255", uint8(255), false},
		{reflect.Uint8, "-128", nil, true},
		{reflect.Uint8, "-44", nil, true},
		{reflect.Uint8, "44", uint8(44), false},
		{reflect.Uint8, "-256", nil, true},
		{reflect.Uint8, "256", nil, true},
		{reflect.Uint8, "-21474836499999", nil, true},
		{reflect.Uint8, "21474836489999", nil, true},
		{reflect.Uint8, "", nil, true},
		{reflect.Uint8, "abcde", nil, true},

		{reflect.Uint16, "0", uint16(0), false},
		{reflect.Uint16, "12345", uint16(12345), false},
		{reflect.Uint16, "-12345", nil, true},
		{reflect.Uint16, "-32768", nil, true},
		{reflect.Uint16, "65535", uint16(65535), false},
		{reflect.Uint16, "-32769", nil, true},
		{reflect.Uint16, "65536", nil, true},
		{reflect.Uint16, "-21474836499999", nil, true},
		{reflect.Uint16, "21474836489999", nil, true},
		{reflect.Uint16, "", nil, true},
		{reflect.Uint16, "abcde", nil, true},

		{reflect.Uint32, "0", uint32(0), false},
		{reflect.Uint32, "12345", uint32(12345), false},
		{reflect.Uint32, "-12345", nil, true},
		{reflect.Uint32, "-2147483648", nil, true},
		{reflect.Uint32, "4294967295", uint32(4294967295), false},
		{reflect.Uint32, "-2147483649", nil, true},
		{reflect.Uint32, "4294967296", nil, true},
		{reflect.Uint32, "-21474836499999", nil, true},
		{reflect.Uint32, "21474836489999", nil, true},
		{reflect.Uint32, "", nil, true},
		{reflect.Uint32, "abcde", nil, true},

		{reflect.Uint64, "0", uint64(0), false},
		{reflect.Uint64, "12345", uint64(12345), false},
		{reflect.Uint64, "-12345", nil, true},
		{reflect.Uint64, "-9223372036854775808", nil, true},
		{reflect.Uint64, "18446744073709551615", uint64(18446744073709551615), false},
		{reflect.Uint64, "-9223372036854775809", nil, true},
		{reflect.Uint64, "18446744073709551616", nil, true},
		{reflect.Uint64, "-21474839999996499999", nil, true},
		{reflect.Uint64, "214748364999999989999", nil, true},
		{reflect.Uint64, "", nil, true},
		{reflect.Uint64, "abcde", nil, true},

		{reflect.Uint, "0", uint(0), false},
		{reflect.Uint, "12345", uint(12345), false},
		{reflect.Uint, "-12345", nil, true},
		{reflect.Uint, "-9223372036854775808", nil, true},
		{reflect.Uint, "18446744073709551615", uint(18446744073709551615), false},
		{reflect.Uint, "-9223372036854775809", nil, true},
		{reflect.Uint, "18446744073709551616", nil, true},
		{reflect.Uint, "-21474839999996499999", nil, true},
		{reflect.Uint, "214748364999999989999", nil, true},
		{reflect.Uint, "", nil, true},
		{reflect.Uint, "abcde", nil, true},

		{reflect.Bool, "true", true, false},
		{reflect.Bool, "false", false, false},
		{reflect.Bool, "0", false, false},
		{reflect.Bool, "1", true, false},
		{reflect.Bool, "1333", nil, true},
		{reflect.Bool, "", nil, true},
		{reflect.Bool, "nottrue", nil, true},

		{reflect.String, "a string", "a string", false},
		{reflect.String, "", "", false},
	}

	shortArgs = []string{
		"-s", stringArg,
		"-i", strconv.FormatUint(uint64(uintArg), 10),
		"-j", strconv.FormatUint(uint64Arg, 10),
		"-k", strconv.FormatUint(uint64(uint32Arg), 10),
		"-l", strconv.FormatUint(uint64(uint16Arg), 10),
		"-m", strconv.FormatUint(uint64(uint8Arg), 10),
		"-n", strconv.FormatInt(int64(intArg), 10),
		"-o", strconv.FormatInt(int64Arg, 10),
		"-p", strconv.FormatInt(int64(int32Arg), 10),
		"-q", strconv.FormatInt(int64(int16Arg), 10),
		"-r", strconv.FormatInt(int64(int8Arg), 10),
		"-b", strconv.FormatBool(boolArg),
	}

	longArgs = []string{
		"--str", stringArg,
		"--uint", strconv.FormatUint(uint64(uintArg), 10),
		"--uint64", strconv.FormatUint(uint64Arg, 10),
		"--uint32", strconv.FormatUint(uint64(uint32Arg), 10),
		"--uint16", strconv.FormatUint(uint64(uint16Arg), 10),
		"--uint8", strconv.FormatUint(uint64(uint8Arg), 10),
		"--int", strconv.FormatInt(int64(intArg), 10),
		"--int64", strconv.FormatInt(int64Arg, 10),
		"--int32", strconv.FormatInt(int64(int32Arg), 10),
		"--int16", strconv.FormatInt(int64(int16Arg), 10),
		"--int8", strconv.FormatInt(int64(int8Arg), 10),
		"--bool", strconv.FormatBool(boolArg),
	}
)

const (
	stringArg        = "string argument"
	uintArg   uint   = 18446744073709551615
	uint64Arg uint64 = 18446744073709551614
	uint32Arg uint32 = 4294967295
	uint16Arg uint16 = 65535
	uint8Arg  uint8  = 127
	intArg    int    = -9223372036854775808
	int64Arg  int64  = -9223372036854775807
	int32Arg  int32  = -2147483648
	int16Arg  int16  = -32768
	int8Arg   int8   = -127
	boolArg          = true

	stringEnvVar = "ENV_STR"
	uintEnvVar   = "ENV_UINT"
	uint64EnvVar = "ENV_UINT64"
	uint32EnvVar = "ENV_UINT32"
	uint16EnvVar = "ENV_UINT16"
	uint8EnvVar  = "ENV_UINT8"
	intEnvVar    = "ENV_INT"
	int64EnvVar  = "ENV_INT64"
	int32EnvVar  = "ENV_INT32"
	int16EnvVar  = "ENV_INT16"
	int8EnvVar   = "ENV_INT8"
	boolEnvVar   = "ENV_BOOL"

	readEnvVar = "TEST_ARG"

	defaultStringArg string = "default str"
	defaultUintArg   uint   = 6744073709551614
	defaultUint64Arg uint64 = 6744073709551615
	defaultUint32Arg uint32 = 42949672
	defaultUint16Arg uint16 = 6553
	defaultUint8Arg  uint8  = 222
	defaultIntArg    int    = 6744073709551614
	defaultInt64Arg  int64  = 6744073709551615
	defaultInt32Arg  int32  = 42949672
	defaultInt16Arg  int16  = 6553
	defaultInt8Arg   int8   = 44
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
	arguments.SetArgs(shortArgs)

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uintArg, argValues.uintArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, uint8Arg, argValues.uInt8Arg)
	assert.Equal(t, intArg, argValues.intArg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
	assert.Equal(t, int8Arg, argValues.int8Arg)
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
	arguments.SetArgs(longArgs)

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uintArg, argValues.uintArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, uint8Arg, argValues.uInt8Arg)
	assert.Equal(t, intArg, argValues.intArg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
	assert.Equal(t, int8Arg, argValues.int8Arg)
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
	_ = os.Setenv(uintEnvVar, strconv.FormatUint(uint64(uintArg), 10))
	_ = os.Setenv(uint64EnvVar, strconv.FormatUint(uint64Arg, 10))
	_ = os.Setenv(uint32EnvVar, strconv.FormatUint(uint64(uint32Arg), 10))
	_ = os.Setenv(uint16EnvVar, strconv.FormatUint(uint64(uint16Arg), 10))
	_ = os.Setenv(uint8EnvVar, strconv.FormatUint(uint64(uint8Arg), 10))
	_ = os.Setenv(intEnvVar, strconv.FormatInt(int64(intArg), 10))
	_ = os.Setenv(int64EnvVar, strconv.FormatInt(int64Arg, 10))
	_ = os.Setenv(int32EnvVar, strconv.FormatInt(int64(int32Arg), 10))
	_ = os.Setenv(int16EnvVar, strconv.FormatInt(int64(int16Arg), 10))
	_ = os.Setenv(int8EnvVar, strconv.FormatInt(int64(int8Arg), 10))
	_ = os.Setenv(boolEnvVar, strconv.FormatBool(boolArg))

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs([]string{})

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uintArg, argValues.uintArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, uint8Arg, argValues.uInt8Arg)
	assert.Equal(t, intArg, argValues.intArg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
	assert.Equal(t, int8Arg, argValues.int8Arg)
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
	assert.Equal(t, defaultUintArg, argValues.uintArg)
	assert.Equal(t, defaultUint64Arg, argValues.uInt64Arg)
	assert.Equal(t, defaultUint32Arg, argValues.uInt32Arg)
	assert.Equal(t, defaultUint16Arg, argValues.uInt16Arg)
	assert.Equal(t, defaultUint8Arg, argValues.uInt8Arg)
	assert.Equal(t, defaultIntArg, argValues.intArg)
	assert.Equal(t, defaultInt64Arg, argValues.int64Arg)
	assert.Equal(t, defaultInt32Arg, argValues.int32Arg)
	assert.Equal(t, defaultInt16Arg, argValues.int16Arg)
	assert.Equal(t, defaultInt8Arg, argValues.int8Arg)
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
	_ = os.Setenv(uintEnvVar, strconv.FormatUint(uint64(uintArg-10), 10))
	_ = os.Setenv(uint64EnvVar, strconv.FormatUint(uint64Arg-10, 10))
	_ = os.Setenv(uint32EnvVar, strconv.FormatUint(uint64(uint32Arg-10), 10))
	_ = os.Setenv(uint16EnvVar, strconv.FormatUint(uint64(uint16Arg-10), 10))
	_ = os.Setenv(uint8EnvVar, strconv.FormatUint(uint64(uint8Arg-10), 10))
	_ = os.Setenv(intEnvVar, strconv.FormatInt(int64(intArg+10), 10))
	_ = os.Setenv(int64EnvVar, strconv.FormatInt(int64Arg+10, 10))
	_ = os.Setenv(int32EnvVar, strconv.FormatInt(int64(int32Arg+10), 10))
	_ = os.Setenv(int16EnvVar, strconv.FormatInt(int64(int16Arg+10), 10))
	_ = os.Setenv(int8EnvVar, strconv.FormatInt(int64(int8Arg+10), 10))
	_ = os.Setenv(boolEnvVar, strconv.FormatBool(!boolArg))

	arguments := NewArgs("use", "short", func(strings []string) error {
		functionExecuted = true
		return nil
	})
	setupArgs(arguments, argValues)
	arguments.SetArgs(longArgs)

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uintArg, argValues.uintArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, uint8Arg, argValues.uInt8Arg)
	assert.Equal(t, intArg, argValues.intArg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
	assert.Equal(t, int8Arg, argValues.int8Arg)
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
	arguments.SetArgs(shortArgs)

	err := arguments.Execute()

	assert.Equal(t, stringArg, argValues.stringArg)
	assert.Equal(t, uintArg, argValues.uintArg)
	assert.Equal(t, uint64Arg, argValues.uInt64Arg)
	assert.Equal(t, uint32Arg, argValues.uInt32Arg)
	assert.Equal(t, uint16Arg, argValues.uInt16Arg)
	assert.Equal(t, uint8Arg, argValues.uInt8Arg)
	assert.Equal(t, intArg, argValues.intArg)
	assert.Equal(t, int64Arg, argValues.int64Arg)
	assert.Equal(t, int32Arg, argValues.int32Arg)
	assert.Equal(t, int16Arg, argValues.int16Arg)
	assert.Equal(t, int8Arg, argValues.int8Arg)
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
	arguments.SetArgs(shortArgs)

	err := arguments.Help()

	assert.Nil(t, err)
}

func TestReadEnvVariable_AllTypes(t *testing.T) {
	for envTestDataRecordIdx, envTestDataRecord := range envTestDataRecords {
		log.Printf("Processing record %2d: %+v\n", envTestDataRecordIdx, envTestDataRecord)

		_ = os.Setenv(readEnvVar, envTestDataRecord.envValue)
		supportedType := supportedArgumentKinds[envTestDataRecord.envKind]
		value, err := readEnvVariable(readEnvVar, envTestDataRecord.envKind, supportedType)

		if envTestDataRecord.expectedValue != nil {
			assert.Equal(t, envTestDataRecord.expectedValue, value, "Idx=%d: %+v", envTestDataRecordIdx, envTestDataRecord)
		} else {
			assert.Nil(t, value, "Idx=%d: %+v", envTestDataRecordIdx, envTestDataRecord)
		}
		if envTestDataRecord.expectedError {
			assert.NotNil(t, err, "Idx=%d: %+v", envTestDataRecordIdx, envTestDataRecord)
		} else {
			assert.Nil(t, err, "Idx=%d: %+v", envTestDataRecordIdx, envTestDataRecord)
		}
	}
}

func setupArgs(arguments *Args, argValues *argumentValues) {
	_ = arguments.SetVarP(&argValues.stringArg, "str", "s", "ENV_STR", defaultStringArg, "Use str")
	_ = arguments.SetVarP(&argValues.uintArg, "uint", "i", uintEnvVar, defaultUintArg, "Use uint")
	_ = arguments.SetVarP(&argValues.uInt64Arg, "uint64", "j", uint64EnvVar, defaultUint64Arg, "Use uint64")
	_ = arguments.SetVarP(&argValues.uInt32Arg, "uint32", "k", uint32EnvVar, defaultUint32Arg, "Use uint32")
	_ = arguments.SetVarP(&argValues.uInt16Arg, "uint16", "l", uint16EnvVar, defaultUint16Arg, "Use uint16")
	_ = arguments.SetVarP(&argValues.uInt8Arg, "uint8", "m", uint8EnvVar, defaultUint8Arg, "Use uint8")
	_ = arguments.SetVarP(&argValues.intArg, "int", "n", intEnvVar, defaultIntArg, "Use int")
	_ = arguments.SetVarP(&argValues.int64Arg, "int64", "o", int64EnvVar, defaultInt64Arg, "Use int64")
	_ = arguments.SetVarP(&argValues.int32Arg, "int32", "p", int32EnvVar, defaultInt32Arg, "Use int32")
	_ = arguments.SetVarP(&argValues.int16Arg, "int16", "q", int16EnvVar, defaultInt16Arg, "Use int16")
	_ = arguments.SetVarP(&argValues.int8Arg, "int8", "r", int8EnvVar, defaultInt8Arg, "Use int8")
	_ = arguments.SetVarP(&argValues.booleanArg, "bool", "b", "ENV_BOOL", defaultBoolArg, "Use bool")
}

func ClearEnvironment() {
	_ = os.Unsetenv(stringEnvVar)
	_ = os.Unsetenv(uintEnvVar)
	_ = os.Unsetenv(uint64EnvVar)
	_ = os.Unsetenv(uint32EnvVar)
	_ = os.Unsetenv(uint16EnvVar)
	_ = os.Unsetenv(uint8EnvVar)
	_ = os.Unsetenv(intEnvVar)
	_ = os.Unsetenv(int64EnvVar)
	_ = os.Unsetenv(int32EnvVar)
	_ = os.Unsetenv(int16EnvVar)
	_ = os.Unsetenv(int8EnvVar)
	_ = os.Unsetenv(boolEnvVar)
}
