package sensu

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sensu/sensu-enterprise-go-plugin/args"
	"github.com/sensu/sensu-go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

type HandlerConfigOption struct {
	Value     interface{}
	Path      string
	Env       string
	Argument  string      // command line argument
	Shorthand string      // short command line argument
	Default   interface{} // default value
	Usage     string
}

type HandlerConfig struct {
	Name     string
	Short    string
	Timeout  uint64
	Keyspace string
}

type GoHandler struct {
	config             *HandlerConfig
	options            []*HandlerConfigOption
	sensuEvent         *types.Event
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
	eventReader        io.Reader
	cmdArgs            *args.Args
}

func NewGoHandler(config *HandlerConfig, options []*HandlerConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *GoHandler {
	goHandler := &GoHandler{
		config:             config,
		options:            options,
		sensuEvent:         nil,
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
		eventReader:        os.Stdin,
	}
	cmdArgs := args.NewArgs(config.Name, config.Short, goHandler.cobraExecute)
	goHandler.cmdArgs = cmdArgs

	return goHandler
}

func (goHandler *GoHandler) Execute() error {
	// Setup arguments
	for _, option := range goHandler.options {
		if option.Value == nil {
			return fmt.Errorf("Option value must not be nil for option %s", option.Argument)
		}

		switch (option.Value).(type) {
		case *string:
			valuePtr, _ := option.Value.(*string)
			goHandler.cmdArgs.StringVarP(valuePtr, option.Argument, option.Shorthand, option.Env,
				option.Default.(string), option.Usage)
		case *uint64:
			valuePtr, _ := option.Value.(*uint64)
			goHandler.cmdArgs.Uint64VarP(valuePtr, option.Argument, option.Shorthand, option.Env,
				option.Default.(uint64), option.Usage)
		case *bool:
			valuePtr, _ := option.Value.(*bool)
			goHandler.cmdArgs.BoolVarP(valuePtr, option.Argument, option.Shorthand, option.Env,
				option.Default.(bool), option.Usage)
		}
	}

	// This will call cobraExecute so put the rest of the logic in there
	err := goHandler.cmdArgs.Execute()
	if err != nil {
		return err
	}

	return nil
}

func (goHandler *GoHandler) readSensuEvent() error {
	eventJSON, err := ioutil.ReadAll(goHandler.eventReader)
	if err != nil {
		return fmt.Errorf("Failed to read STDIN: %s", err)
	}

	sensuEvent := &types.Event{}
	err = json.Unmarshal(eventJSON, sensuEvent)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal STDIN data: %s", err)
	}

	if err = validateEvent(sensuEvent); err != nil {
		return err
	}

	goHandler.sensuEvent = sensuEvent
	return nil
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

func configurationOverrides(config *HandlerConfig, options []*HandlerConfigOption, event *types.Event) error {
	if config.Keyspace == "" {
		return nil
	}
	for _, opt := range options {
		if len(opt.Path) > 0 {
			// compile the Annotation keyspace to look for configuration overrides
			k := path.Join(config.Keyspace, opt.Path)
			switch {
			case len(event.Check.Annotations[k]) > 0:
				err := setOptionValue(opt, event.Check.Annotations[k])
				if err != nil {
					return err
				}
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n", k, event.Check.Annotations[k])
			case len(event.Entity.Annotations[k]) > 0:
				err := setOptionValue(opt, event.Entity.Annotations[k])
				if err != nil {
					return err
				}
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n", k, event.Entity.Annotations[k])
			}
		}
	}
	return nil
}

func setOptionValue(option *HandlerConfigOption, valueStr string) error {
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

// Intentionally does nothing since we're only using cobra to read the command line arguments
func (goHandler *GoHandler) cobraExecute(_ []string) error {
	// Read Sensu event
	err := goHandler.readSensuEvent()
	if err != nil {
		return err
	}

	// Override the configuration with the event information
	err = configurationOverrides(goHandler.config, goHandler.options, goHandler.sensuEvent)
	if err != nil {
		return err
	}

	// Validate input using validateFunction
	err = goHandler.validationFunction(goHandler.sensuEvent)
	if err != nil {
		return fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	err = goHandler.executeFunction(goHandler.sensuEvent)
	if err != nil {
		return fmt.Errorf("error executing handler: %s", err)
	}

	return nil
}
