package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"os"
)

type GoHandler struct {
	GoPlugin
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
}

func NewGoHandler(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) (*GoHandler, error) {
	goHandler := &GoHandler{
		GoPlugin: GoPlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              true,
			eventMandatory:         true,
			configurationOverrides: true,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}
	goHandler.pluginWorkflowFunction = goHandler.goHandlerWorkflow

	goHandler.initPlugin()

	return goHandler, nil
}

// Executes the handler's workflow
func (goHandler *GoHandler) goHandlerWorkflow(_ []string) error {
	// Validate input using validateFunction
	err := goHandler.validationFunction(goHandler.sensuEvent)
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
