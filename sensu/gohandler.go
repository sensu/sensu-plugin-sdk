package sensu

import (
	"fmt"
	"os"

	"github.com/sensu/sensu-go/types"
	"log"
)

type GoHandler struct {
	basePlugin
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
}

func NewGoHandler(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *GoHandler {
	goHandler := &GoHandler{
		basePlugin: basePlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              true,
			eventMandatory:         true,
			configurationOverrides: true,
			errorExitStatus:        1,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}

	goHandler.pluginWorkflowFunction = goHandler.goHandlerWorkflow
	if err := goHandler.initPlugin(); err != nil {
		log.Fatalf("failed to initialize handler plugin: %s", err)
	}

	return goHandler
}

// Executes the handler's workflow
func (goHandler *GoHandler) goHandlerWorkflow(_ []string) (int, error) {
	// Validate input using validateFunction
	err := goHandler.validationFunction(goHandler.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	err = goHandler.executeFunction(goHandler.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error executing handler: %s", err)
	}

	return 0, nil
}
