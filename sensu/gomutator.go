package sensu

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sensu/sensu-go/types"
)

type GoMutator struct {
	basePlugin
	out                io.Writer
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) (*types.Event, error)
}

func NewGoMutator(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error,
	executeFunction func(event *types.Event) (*types.Event, error)) *GoMutator {
	goMutator := &GoMutator{
		basePlugin: basePlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              true,
			eventMandatory:         true,
			eventValidation:        true,
			configurationOverrides: true,
			exitFunction:           os.Exit,
			errorExitStatus:        1,
		},
		out:                os.Stdout,
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}
	goMutator.pluginWorkflowFunction = goMutator.goMutatorWorkflow
	if err := goMutator.initPlugin(); err != nil {
		log.Printf("failed to initialize mutator plugin: %s", err)
	}
	return goMutator
}

// Executes the handler's workflow
func (goMutator *GoMutator) goMutatorWorkflow(_ []string) (int, error) {
	// Validate input using validateFunction
	err := goMutator.validationFunction(goMutator.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	event, err := goMutator.executeFunction(goMutator.sensuEvent)
	if err != nil {
		return 1, fmt.Errorf("error executing mutator: %s", err)
	}

	if event != nil {
		eventBytes, err := json.Marshal(event)
		if err != nil {
			return 1, fmt.Errorf("error marshaling output event to json: %s", err)
		}

		_, _ = fmt.Fprintf(goMutator.out, "%s", string(eventBytes))
	} else {
		_, _ = fmt.Fprint(goMutator.out, "{}")
	}

	return 0, err
}
