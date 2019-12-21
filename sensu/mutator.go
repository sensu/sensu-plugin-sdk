package sensu

import (
	"encoding/json"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"io"
	"os"
)

type Mutator struct {
	basePlugin
	out                io.Writer
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) (*types.Event, error)
}

func InitMutator(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error,
	executeFunction func(event *types.Event) (*types.Event, error)) *Mutator {
	mutator := &Mutator{
		basePlugin: basePlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              true,
			configurationOverrides: true,
			exitFunction:           os.Exit,
			errorExitStatus:        1,
		},
		out:                os.Stdout,
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}
	mutator.pluginExecuteFunction = mutator.execute
	mutator.initPlugin()
	return mutator
}

// Executes the mutator
func (mutator *Mutator) execute(_ []string) (int, error) {
	// Validate input using validateFunction
	err := mutator.validationFunction(mutator.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	event, err := mutator.executeFunction(mutator.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error executing mutator: %s", err)
	}

	if event != nil {
		eventBytes, err := json.Marshal(event)
		if err != nil {
			return 1, fmt.Errorf("error marshaling output event to json: %s", err)
		}

		_, _ = fmt.Fprintf(mutator.out, "%s", string(eventBytes))
	} else {
		_, _ = fmt.Fprint(mutator.out, "{}")
	}

	return 0, err
}
