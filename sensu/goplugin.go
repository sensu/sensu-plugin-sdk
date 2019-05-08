package sensu

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugins-go-library/args"
	"io"
	"io/ioutil"
	"log"
	"path"
	"strconv"
)

type PluginConfigOption struct {
	Value     interface{}
	Path      string
	Env       string
	Argument  string      // command line argument
	Shorthand string      // short command line argument
	Default   interface{} // default value
	Usage     string
}

type PluginConfig struct {
	Name     string
	Short    string
	Timeout  uint64
	Keyspace string
}

type GoPlugin struct {
	config                 *PluginConfig
	options                []*PluginConfigOption
	sensuEvent             *types.Event
	eventReader            io.Reader
	pluginWorkflowFunction func(args []string) error
	cmdArgs                *args.Args
	readEvent              bool
	eventMandatory         bool
	configurationOverrides bool
}

func (goPlugin *GoPlugin) readSensuEvent() error {
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

func (goPlugin *GoPlugin) initPlugin() {
	goPlugin.cmdArgs = args.NewArgs(goPlugin.config.Name, goPlugin.config.Short, goPlugin.cobraExecuteFunction)
}

func (goPlugin *GoPlugin) setupArguments() error {
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
func (goPlugin *GoPlugin) cobraExecuteFunction(args []string) error {
	// Read the Sensu event if required
	if goPlugin.readEvent {
		err := goPlugin.readSensuEvent()
		if err != nil {
			return err
		}
	}

	// If there is an event process configuration overrides if necessary
	if goPlugin.sensuEvent != nil && goPlugin.configurationOverrides {
		err := configurationOverrides(goPlugin.config, goPlugin.options, goPlugin.sensuEvent)
		if err != nil {
			return err
		}
	}

	return goPlugin.pluginWorkflowFunction(args)
}

func (goPlugin *GoPlugin) Execute() error {
	// Validate the arguments are set
	if goPlugin.cmdArgs == nil {
		return fmt.Errorf("Arguments must be initialized")
	}

	err := goPlugin.setupArguments()
	if err != nil {
		return err
	}

	// This will call the pluginWorkflowFunction function which implements the custom logic for that type of plugin.
	return goPlugin.cmdArgs.Execute()
}

func validateEvent(event *types.Event) error {
	if event.Timestamp <= 0 {
		return errors.New("timestamp is missing or must be greater than zero")
	}

	if event.Entity == nil {
		return errors.New("entity is missing from event")
	}

	if !event.HasCheck() {
		return errors.New("check is missing from event")
	}

	if err := event.Entity.Validate(); err != nil {
		return err
	}

	if err := event.Check.Validate(); err != nil {
		return err
	}

	return nil
}

func setOptionValue(option *PluginConfigOption, valueStr string) error {
	switch option.Value.(type) {
	case *string:
		strOptionValue, ok := option.Value.(*string)
		if ok {
			*strOptionValue = valueStr
		}
	case *uint64:
		uint64OptionPtrValue, ok := option.Value.(*uint64)
		if ok {
			parsedValue, err := strconv.ParseUint(valueStr, 10, 64)
			if err != nil {
				return fmt.Errorf("Error parsing %s into a uint64 for option %s", valueStr, option.Argument)
			}
			*uint64OptionPtrValue = parsedValue
		}
	case *bool:
		boolOptionPtrValue, ok := option.Value.(*bool)
		if ok {
			parsedValue, err := strconv.ParseBool(valueStr)
			if err != nil {
				return fmt.Errorf("Error parsing %s into a bool for option %s", valueStr, option.Argument)
			}
			*boolOptionPtrValue = parsedValue
		}
	}
	return nil
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
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n",
					key, event.Entity.Annotations[key])
			}
		}
	}
	return nil
}
