package args

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"reflect"
	"strconv"
)

// ExecutorFunction is a type that defines a parseFunction to be executed after
// parsing the arguments.
type ExecutorFunction func([]string) error

// Args is a wrapper on top of cobra to read program arguments. In addition to
// reading command line arguments it reads the arguments from the programs
// environment, the command line having priority. A default value is used if
// the environment variable and the command line argument are not present.
type Args struct {
	cmd  *cobra.Command
	runE ExecutorFunction
}

type supportedType struct {
	parseFunction interface{}
	functionName  string
	args          []reflect.Value
}

var (
	supportedArgumentKinds = map[reflect.Kind]*supportedType{
		reflect.Uint64: {
			parseFunction: strconv.ParseUint,
			functionName:  "strconv.ParseUint",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(64),
			},
		},
		reflect.Uint32: {
			parseFunction: strconv.ParseUint,
			functionName:  "strconv.ParseUint",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(32),
			},
		},
		reflect.Uint16: {
			parseFunction: strconv.ParseUint,
			functionName:  "strconv.ParseUint",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(16),
			},
		},
		reflect.Uint8: {
			parseFunction: strconv.ParseUint,
			functionName:  "strconv.ParseUint",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(8),
			},
		},
		reflect.Int64: {
			parseFunction: strconv.ParseInt,
			functionName:  "strconv.ParseInt",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(64),
			},
		},
		reflect.Int32: {
			parseFunction: strconv.ParseInt,
			functionName:  "strconv.ParseInt",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(32),
			},
		},
		reflect.Int16: {
			parseFunction: strconv.ParseInt,
			functionName:  "strconv.ParseInt",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(16),
			},
		},
		reflect.Int8: {
			parseFunction: strconv.ParseInt,
			functionName:  "strconv.ParseInt",
			args: []reflect.Value{
				reflect.ValueOf(10),
				reflect.ValueOf(8),
			},
		},
	}
)

// NewArgs creates an Args object based on the cobra library
func NewArgs(use string, short string, runE ExecutorFunction) *Args {
	args := &Args{
		cmd: &cobra.Command{
			Use:   use,
			Short: short,
		},
		runE: runE,
	}
	args.cmd.RunE = args.cobraRunE

	return args
}

// cobraRunE is the parseFunction to execute by cobra when done with parsing the
// arguments. It simply passes control over to the the Args.runE parseFunction.
func (args *Args) cobraRunE(cmd *cobra.Command, arguments []string) error {
	return args.runE(arguments)
}

// Execute uses the args and run through the command tree finding appropriate
// matches for commands and then corresponding flags.
func (args *Args) Execute() error {
	return args.cmd.Execute()
}

// Help prints out the help for the command.
func (args *Args) Help() error {
	return args.cmd.Help()
}

// StringVarP reads a string argument from the command line arguments or the
// program's environment. defaultValue is used if none is present.
func (args *Args) StringVarP(p *string, name, shorthand string, envKey string, defaultValue string, usage string) {
	envValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	}

	args.cmd.Flags().StringVarP(p, name, shorthand, envValue, usage)
}

// Uint64VarP reads a uint64 argument from the command line arguments or the
// program's environment. defaultValue is used if none is present or an invalid
// value is present in the environment.
func (args *Args) Uint64VarP(p *uint64, name, shorthand string, envKey string, defaultValue uint64, usage string) {
	var envValue uint64
	envStrValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	} else {
		parsedValue, err := strconv.ParseUint(envStrValue, 10, 64)
		if err == nil {
			envValue = parsedValue
		} else {
			envValue = defaultValue
		}
	}
	args.cmd.Flags().Uint64VarP(p, name, shorthand, envValue, usage)
}

// Uint32VarP reads a uint32 argument from the command line arguments or the
// program's environment. defaultValue is used if none is present or an invalid
// value is present in the environment.
func (args *Args) Uint32VarP(p *uint32, name, shorthand string, envKey string, defaultValue uint32, usage string) {
	var envValue uint32
	envStrValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	} else {
		parsedValue, err := strconv.ParseUint(envStrValue, 10, 32)
		if err == nil {
			envValue = uint32(parsedValue)
		} else {
			envValue = defaultValue
		}
	}
	args.cmd.Flags().Uint32VarP(p, name, shorthand, envValue, usage)
}

// Uint16VarP reads a uint16 argument from the command line arguments or the
// program's environment. defaultValue is used if none is present or an invalid
// value is present in the environment.
func (args *Args) Uint16VarP(p *uint16, name, shorthand string, envKey string, defaultValue uint16, usage string) {
	var envValue uint16
	envStrValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	} else {
		parsedValue, err := strconv.ParseUint(envStrValue, 10, 16)
		if err == nil {
			envValue = uint16(parsedValue)
		} else {
			envValue = defaultValue
		}
	}
	args.cmd.Flags().Uint16VarP(p, name, shorthand, envValue, usage)
}

// Int64VarP reads a int64 argument from the command line arguments or the
// program's environment. defaultValue is used if none is present or an invalid
// value is present in the environment.
func (args *Args) Int64VarP(p *int64, name, shorthand string, envKey string, defaultValue int64, usage string) {
	var envValue int64
	envStrValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	} else {
		parsedValue, err := strconv.ParseInt(envStrValue, 10, 64)
		if err == nil {
			envValue = parsedValue
		} else {
			envValue = defaultValue
		}
	}
	args.cmd.Flags().Int64VarP(p, name, shorthand, envValue, usage)
}

// Int32VarP reads a int32 argument from the command line arguments or the
// program's environment. defaultValue is used if none is present or an invalid
// value is present in the environment.
func (args *Args) Int32VarP(p *int32, name, shorthand string, envKey string, defaultValue int32, usage string) {
	var envValue int32
	envStrValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	} else {
		parsedValue, err := strconv.ParseInt(envStrValue, 10, 32)
		if err == nil {
			envValue = int32(parsedValue)
		} else {
			envValue = defaultValue
		}
	}
	args.cmd.Flags().Int32VarP(p, name, shorthand, envValue, usage)
}

// Int16VarP reads a int16 argument from the command line arguments or the
// program's environment. defaultValue is used if none is present or an invalid
// value is present in the environment.
func (args *Args) Int16VarP(p *int16, name, shorthand string, envKey string, defaultValue int16, usage string) {
	var envValue int16
	envStrValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	} else {
		parsedValue, err := strconv.ParseInt(envStrValue, 10, 16)
		if err == nil {
			envValue = int16(parsedValue)
		} else {
			envValue = defaultValue
		}
	}
	args.cmd.Flags().Int16VarP(p, name, shorthand, envValue, usage)
}

// BoolVarP reads a uint64 argument from the command line arguments or the
// program's environment. defaultValue is used if none is present or an invalid
// value is present in the environment.
func (args *Args) BoolVarP(p *bool, name, shorthand string, envKey string, defaultValue bool, usage string) {
	var envValue bool
	envStrValue, ok := os.LookupEnv(envKey)
	if !ok {
		envValue = defaultValue
	} else {
		parsedValue, err := strconv.ParseBool(envStrValue)
		if err == nil {
			envValue = parsedValue
		} else {
			envValue = defaultValue
		}
	}
	args.cmd.Flags().BoolVarP(p, name, shorthand, envValue, usage)
}

func (args *Args) SetVarP(destValue interface{}, name, shorthand, envKey string, defaultValue interface{}, usage string) error {

	if destValue == nil {
		return fmt.Errorf("destValue must not be nil")
	}

	interfaceType := reflect.TypeOf(destValue)
	interfaceKind := interfaceType.Kind()
	if interfaceKind == reflect.Ptr {
		element := interfaceType.Elem()
		elementKind := element.Kind()
		conversionFunction := supportedArgumentKinds[elementKind]

		log.Printf("Type: %v", conversionFunction)

		if conversionFunction != nil {

			arguments := append([]reflect.Value{reflect.ValueOf(strValue)}, conversionFunction.args...)
			conversionFunction := reflect.ValueOf(conversionFunction.parseFunction)
			returnValues := conversionFunction.Call(arguments)

			valueInterface := returnValues[0].Interface()
			errorInterface := returnValues[1].Interface()

			if errorInterface != nil {
				log.Printf("there is an error: %s", errorInterface.(error))
			} else {
				log.Printf("Returned value: %d", valueInterface)
			}
		} else {
			return fmt.Errorf("destValue type not supported: %s", interfaceType)
		}
	} else {
		return fmt.Errorf("destValue must be a pointer")
	}

	return nil
}

func readEnvVariable(envKey string, kind reflect.Kind, supportedType *supportedType) (reflect.Value, error) {
	envValue, found := os.LookupEnv(envKey)
}

func (args *Args) SetArgs(newArgs []string) {
	args.cmd.SetArgs(newArgs)
}
