package sensu

import (
	"encoding/json"
	"fmt"
	"os"

	"log"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-licensing/api/licensing"
)

type GoHandler struct {
	basePlugin
	validationFunction func(event *types.Event) error
	executeFunction    func(event *types.Event) error
	enterprise         bool
}

func NewGoHandler(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *GoHandler {
	goHandler := &GoHandler{
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
			errorExitStatus:        1,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
	}

	goHandler.pluginWorkflowFunction = goHandler.goHandlerWorkflow
	if err := goHandler.initPlugin(); err != nil {
		log.Printf("failed to initialize handler plugin: %s", err)
	}

	return goHandler
}

func NewEnterpriseGoHandler(config *PluginConfig, options []*PluginConfigOption,
	validationFunction func(event *types.Event) error, executeFunction func(event *types.Event) error) *GoHandler {
	goHandler := &GoHandler{
		basePlugin: basePlugin{
			config:                 config,
			options:                options,
			sensuEvent:             nil,
			eventReader:            os.Stdin,
			readEvent:              true,
			eventMandatory:         true,
			configurationOverrides: true,
			errorExitStatus:        1,
		},
		validationFunction: validationFunction,
		executeFunction:    executeFunction,
		enterprise:         true,
	}

	goHandler.pluginWorkflowFunction = goHandler.goHandlerWorkflow
	if err := goHandler.initPlugin(); err != nil {
		log.Printf("failed to initialize handler plugin: %s", err)
	}

	return goHandler
}

// Executes the handler's workflow
func (goHandler *GoHandler) goHandlerWorkflow(_ []string) (int, error) {
	event := goHandler.sensuEvent
	if goHandler.enterprise {
		var licenseFile *licensing.LicenseFile
		license := os.Getenv("SENSU_LICENSE_FILE")
		if license == "" {
			return 1, fmt.Errorf("valid sensu license is required to execute")
		}
		err := json.Unmarshal([]byte(license), &licenseFile)
		if err != nil {
			return 1, fmt.Errorf("error reading license file: %s", err)
		}
		err = licenseFile.Validate()
		if err != nil {
			return 1, fmt.Errorf("error validating license file: %s", err)
		}
	}

	// Validate input using validateFunction
	err := goHandler.validationFunction(event)
	if err != nil {
		return 1, fmt.Errorf("error validating input: %s", err)
	}

	// Execute handler logic using executeFunction
	err = goHandler.executeFunction(event)
	if err != nil {
		return 1, fmt.Errorf("error executing handler: %s", err)
	}

	return 0, nil
}
