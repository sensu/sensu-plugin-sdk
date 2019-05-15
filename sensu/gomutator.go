package sensu

import (
	"encoding/json"
	"fmt"
	"github.com/sensu/sensu-go/types"
	"io"
	"os"
)

type GoMutator struct {
	GoPlugin
	out                io.Writer
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) (*types.Event, error)
}

func NewGoMutator(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error,
	executeFunction func(event *types.Event) (*types.Event, error)) (*GoMutator, error) {
	goMutator := &GoMutator{
		GoPlugin: GoPlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              true,
			eventMandatory:         true,
			configurationOverrides: true,
		},
		out:                os.Stdout,
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}
	goMutator.pluginWorkflowFunction = goMutator.goMutatorWorkflow

	goMutator.initPlugin()

	return goMutator, nil
}

// Executes the handler's workflow
func (goMutator *GoMutator) goMutatorWorkflow(_ []string) error {
	// Validate input using validateFunction
	err := goMutator.validationFunction(goMutator.sensuEvent)
	if err != nil {
		return fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	event, err := goMutator.executeFunction(goMutator.sensuEvent)
	if err != nil {
		return fmt.Errorf("error executing mutator: %s", err)
	}

	if event != nil {
		eventBytes, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("error marshaling output event to json: %s", err)
		}

		_, err = fmt.Fprintf(goMutator.out, "%s", string(eventBytes))
	} else {
		_, err = fmt.Fprint(goMutator.out, "{}")
	}

	return err
}
