package sensu

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Mutator is the framework for writing sensu mutators.
type Mutator struct {
	framework          pluginFramework
	out                io.Writer
	validationFunction func(event *corev2.Event) error
	executeFunction    func(event *corev2.Event) (*corev2.Event, error)
}

// GoMutator is deprecated, use Mutator
type GoMutator = Mutator

// NewMutator creates a new mutator.
func NewMutator(config *PluginConfig, options []ConfigOption,
	validationFunction func(event *corev2.Event) error,
	executeFunction func(event *corev2.Event) (*corev2.Event, error)) *Mutator {
	mutator := &Mutator{
		framework: pluginFramework{
			config:                 config,
			options:                options,
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
	mutator.framework.SetWorkflow(mutator.workflow)
	if err := mutator.framework.Init(); err != nil {
		log.Printf("failed to initialize mutator plugin: %s", err)
	}
	return mutator
}

// NewGoMutator is deprecated, use NewMutator
var NewGoMutator = NewMutator

// Executes the handler's workflow
func (m *Mutator) workflow(_ []string) (int, error) {
	// Validate input using validateFunction
	err := m.validationFunction(m.framework.GetStdinEvent())
	if err != nil {
		return 1, ErrValidationFailed(err.Error())
	}

	// Execute handler logic using executeFunction
	event, err := m.executeFunction(m.framework.GetStdinEvent())
	if err != nil {
		return 1, fmt.Errorf("error executing mutator: %s", err)
	}

	if event != nil {
		eventBytes, err := json.Marshal(event)
		if err != nil {
			return 1, fmt.Errorf("error marshaling output event to json: %s", err)
		}

		_, _ = fmt.Fprintf(m.out, "%s", string(eventBytes))
	} else {
		_, _ = fmt.Fprint(m.out, "{}")
	}

	return 0, err
}

// Execute is the mutator's entry point.
func (m *Mutator) Execute() {
	m.framework.Execute()
}
