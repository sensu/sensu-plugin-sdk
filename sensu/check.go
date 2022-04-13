package sensu

import (
	"fmt"
	"log"
	"os"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	CheckStateOK       = 0
	CheckStateWarning  = 1
	CheckStateCritical = 2
	CheckStateUnknown  = 3
)

// GoCheck is a framework for writing sensu checks.
// Deprecated: use Check
type GoCheck = Check

// Check is a framework for writing sensu checks.
type Check struct {
	framework          pluginFramework
	validationFunction func(event *corev2.Event) (int, error)
	executeFunction    func(event *corev2.Event) (int, error)
}

// NewCheck creates a new check.
func NewCheck(config *PluginConfig, options []ConfigOption,
	validationFunction func(*corev2.Event) (int, error),
	executeFunction func(*corev2.Event) (int, error), readEvent bool) *Check {
	check := &Check{
		framework: pluginFramework{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			eventValidation:        false,
			readEvent:              readEvent,
			configurationOverrides: true,
			verbose:                false,
			errorExitStatus:        1,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}

	check.framework.SetWorkflow(check.workflow)
	if err := check.framework.Init(); err != nil {
		log.Printf("failed to initialize check plugin: %s", err)
	}

	return check
}

// NewGoCheck creates a new check.
// Deprecated: use NewCheck
var NewGoCheck = NewCheck

// Executes the check
func (c *Check) workflow(_ []string) (int, error) {
	// Validate input using validateFunction
	status, err := c.validationFunction(c.framework.GetStdinEvent())
	if err != nil {
		return status, ErrValidationFailed(err.Error())
	}

	// Execute check logic using executeFunction
	status, err = c.executeFunction(c.framework.GetStdinEvent())
	if err != nil {
		return status, fmt.Errorf("error executing check: %s", err)
	}

	return status, nil
}

// Execute is the check's entry point.
func (c *Check) Execute() {
	c.framework.Execute()
}
