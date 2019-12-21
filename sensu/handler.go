package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"os"
)

type Handler struct {
	basePlugin
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
}

func InitHandler(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *Handler {
	handler := &Handler{
		basePlugin: basePlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              true,
			configurationOverrides: true,
			errorExitStatus:        1,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}

	handler.pluginExecuteFunction = handler.execute
	handler.initPlugin()

	return handler
}

// Executes the handler
func (handler *Handler) execute(_ []string) (int, error) {
	// Validate input using validateFunction
	err := handler.validationFunction(handler.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	err = handler.executeFunction(handler.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error executing handler: %s", err)
	}

	return 0, nil
}
