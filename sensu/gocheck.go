package sensu

import (
	"fmt"
	"os"

	"github.com/sensu/sensu-go/types"
)

type GoCheck struct {
	basePlugin
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
}

func NewGoCheck(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *GoCheck {
	check := &GoCheck{
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

	check.pluginWorkflowFunction = check.goCheckWorkflow
	check.initPlugin()

	return check
}

// Executes the check
func (goCheck *GoCheck) goCheckWorkflow(_ []string) (int, error) {
	// Validate input using validateFunction
	err := goCheck.validationFunction(goCheck.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	err = goCheck.executeFunction(goCheck.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error executing handler: %s", err)
	}

	return 0, nil
}
