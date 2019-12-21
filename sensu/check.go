package sensu

import (
	"fmt"
	"github.com/sensu/sensu-go/types"
	"os"
)

type Check struct {
	basePlugin
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
}

func InitCheck(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *Check {
	check := &Check{
		basePlugin: basePlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              false,
			configurationOverrides: true,
			errorExitStatus:        1,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}

	check.pluginExecuteFunction = check.execute
	check.initPlugin()

	return check
}

// Executes the check
func (check *Check) execute(_ []string) (int, error) {
	// Validate input using validateFunction
	err := check.validationFunction(check.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	err = check.executeFunction(check.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error executing handler: %s", err)
	}

	return 0, nil
}
