package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugins-go-library/args"
	"os"
)

type GoHandler struct {
	GoPlugin
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
}

func NewGoHandler(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *GoHandler {
	goHandler := &GoHandler{
		GoPlugin: GoPlugin{
			config:      config,
			options:     options,
			sensuEvent:  nil,
			eventReader: os.Stdin,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}
	cmdArgs := args.NewArgs(config.Name, config.Short, goHandler.cobraExecute)
	goHandler.cmdArgs = cmdArgs

	return goHandler
}

func (goHandler *GoHandler) Execute() error {
	// Setup arguments
	err := goHandler.setupArgs()
	if err != nil {
		return err
	}

	// This will call cobraExecute so put the rest of the logic in there
	err = goHandler.cmdArgs.Execute()
	if err != nil {
		return err
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
