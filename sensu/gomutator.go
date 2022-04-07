package sensu

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type GoMutator struct {
	basePlugin
	out                io.Writer
	validationFunction func(event *corev2.Event) error
	executeFunction    func(event *corev2.Event) (*corev2.Event, error)
}

func NewGoMutator(config *PluginConfig, options []ConfigOption,
	validationFunction func(event *corev2.Event) error,
	executeFunction func(event *corev2.Event) (*corev2.Event, error)) *GoMutator {
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
			verbose:                true,
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
		return 1, ErrValidationFailed(err.Error())
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
