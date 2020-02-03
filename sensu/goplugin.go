package sensu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugins-go-library/args"
)

// GoPlugin defines the GoPlugin interface to be implemented by all types of plugins
type GoPlugin interface {
	Execute()
}

// PluginConfigOption defines an option to be read by the plugin on startup. An
// option can be passed using a command line argument, an environment variable or
// at for some plugin types using a configuration override from the Sensu event.
//
//
type PluginConfigOption struct {
	// Value is the value to read the configured flag or environment variable into.
	// Pass a pointer to any value in your plugin in order to fill it in with the
	// data from a flag or environment variable. The parsing will be done with
	// a function supplied by viper. (See ViperFunc)
	Value interface{}

	// Path is the path to the Sensu annotation to consult when parsing config.
	Path string

	// Env is the environment variable to consult when parsing config.
	Env string

	// Argument is the command line argument to consult when parsing config.
	Argument string

	// Shorthand is the shorthand command line argument to consult when parsing config.
	Shorthand string

	// Default is the default value of the config option.
	Default interface{}

	// Usage adds help context to the command-line flag.
	Usage string
}

// PluginConfig defines the base plugin configuration.
type PluginConfig struct {
	Name     string
	Short    string
	Timeout  uint64
	Keyspace string
}

// basePlugin defines the basic configuration to be used by all plugin types.
type basePlugin struct {
	config                 *PluginConfig
	options                []*PluginConfigOption
	sensuEvent             *types.Event
	eventReader            io.Reader
	pluginWorkflowFunction func([]string) (int, error)
	cmdArgs                *args.Args
	readEvent              bool
	eventMandatory         bool
	configurationOverrides bool
	exitStatus             int
	errorExitStatus        int
	exitFunction           func(int)
	errorLogFunction       func(format string, a ...interface{})
}

func (goPlugin *basePlugin) readSensuEvent() error {
	eventJSON, err := ioutil.ReadAll(goPlugin.eventReader)
	if err != nil {
		if goPlugin.eventMandatory {
			return fmt.Errorf("Failed to read STDIN: %s", err)
		} else {
			// if event is not mandatory return without going any further
			return nil
		}
	}

	sensuEvent := &types.Event{}
	err = json.Unmarshal(eventJSON, sensuEvent)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal STDIN data: %s", err)
	}

	if err = validateEvent(sensuEvent); err != nil {
		return err
	}

	goPlugin.sensuEvent = sensuEvent
	return nil
}

func (goPlugin *basePlugin) initPlugin() {
	goPlugin.cmdArgs = args.NewArgs(goPlugin.config.Name, goPlugin.config.Short, goPlugin.cobraExecuteFunction)
	goPlugin.exitFunction = os.Exit
	goPlugin.errorLogFunction = func(format string, a ...interface{}) {
		_, _ = fmt.Fprintf(os.Stderr, format, a)
	}
}

func (goPlugin *basePlugin) setupArguments() error {
	for _, option := range goPlugin.options {
		if option.Value == nil {
			return fmt.Errorf("Option value must not be nil for %s", option.Argument)
		}

		switch (option.Value).(type) {
		case *string:
			valuePtr, _ := option.Value.(*string)
			goPlugin.cmdArgs.StringVarP(valuePtr, option.Argument, option.Shorthand, option.Env,
				option.Default.(string), option.Usage)
		case *uint64:
			valuePtr, _ := option.Value.(*uint64)
			goPlugin.cmdArgs.Uint64VarP(valuePtr, option.Argument, option.Shorthand, option.Env,
				option.Default.(uint64), option.Usage)
		case *bool:
			valuePtr, _ := option.Value.(*bool)
			goPlugin.cmdArgs.BoolVarP(valuePtr, option.Argument, option.Shorthand, option.Env,
				option.Default.(bool), option.Usage)
		}
	}

	return nil
}

// cobraExecuteFunction is called by the argument's execute. The configuration overrides will be processed if necessary
// and the pluginWorkflowFunction function executed
func (goPlugin *basePlugin) cobraExecuteFunction(args []string) error {
	// Read the Sensu event if required
	if goPlugin.readEvent {
		err := goPlugin.readSensuEvent()
		if err != nil {
			goPlugin.exitStatus = goPlugin.errorExitStatus
			return err
		}
	}

	// If there is an event process configuration overrides if necessary
	if goPlugin.sensuEvent != nil && goPlugin.configurationOverrides {
		err := configurationOverrides(goPlugin.config, goPlugin.options, goPlugin.sensuEvent)
		if err != nil {
			goPlugin.exitStatus = goPlugin.errorExitStatus
			return err
		}
	}

	exitStatus, err := goPlugin.pluginWorkflowFunction(args)
	if err != nil {
		fmt.Printf("Error executing plugin: %s", err)
	}
	goPlugin.exitStatus = exitStatus

	return err
}

func (goPlugin *basePlugin) Execute() {
	// Validate the arguments are set
	if goPlugin.cmdArgs == nil {
		goPlugin.errorLogFunction("Error executing %s: Arguments must be initialized\n", goPlugin.config.Name)
		goPlugin.exitFunction(goPlugin.errorExitStatus)
	}

	err := goPlugin.setupArguments()
	if err != nil {
		goPlugin.errorLogFunction("Error executing %s: %s\n", goPlugin.config.Name, err)
		goPlugin.exitFunction(goPlugin.errorExitStatus)
	}

	// This will call the pluginWorkflowFunction function which implements the custom logic for that type of plugin.
	err = goPlugin.cmdArgs.Execute()
	if err != nil {
		goPlugin.errorLogFunction("Error executing %s: %v\n", goPlugin.config.Name, err)
	}

	goPlugin.exitFunction(goPlugin.exitStatus)
}

func validateEvent(event *types.Event) error {
	if event.Timestamp <= 0 {
		return errors.New("timestamp is missing or must be greater than zero")
	}

	return event.Validate()
}

func setOptionValue(opt *PluginConfigOption, valueStr string) error {
	optVal := reflect.Indirect(reflect.ValueOf(opt.Value))
	if optVal.Type().Kind() == reflect.String {
		optVal.Set(reflect.ValueOf(valueStr))
		return nil
	}
	return json.Unmarshal([]byte(valueStr), &opt.Value)
}

func configurationOverrides(config *PluginConfig, options []*PluginConfigOption, event *types.Event) error {
	if config.Keyspace == "" {
		return nil
	}
	for _, opt := range options {
		if len(opt.Path) > 0 {
			// compile the Annotation keyspace to look for configuration overrides
			key := path.Join(config.Keyspace, opt.Path)
			switch {
			case len(event.Check.Annotations[key]) > 0:
				err := setOptionValue(opt, event.Check.Annotations[key])
				if err != nil {
					return err
				}
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n",
					key, event.Check.Annotations[key])
			case len(event.Entity.Annotations[key]) > 0:
				err := setOptionValue(opt, event.Entity.Annotations[key])
				if err != nil {
					return err
				}
				log.Printf("Overriding default handler configuration with value of \"Entity.Annotations.%s\" (\"%s\")\n",
					key, event.Entity.Annotations[key])
			}
		}
	}
	return nil
}
