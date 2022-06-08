package sensu

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-licensing/api/licensing"
)

// Handler is a framework for writing Sensu handlers.
type Handler struct {
	framework          pluginFramework
	validationFunction func(event *corev2.Event) error
	executeFunction    func(event *corev2.Event) error
	enterprise         bool
}

// GoHandler is a framework for writing Sensu handlers.
// Deprecated: use Handler instead.
type GoHandler = Handler

// NewHandler creates a new handler.
func NewHandler(config *PluginConfig, options []ConfigOption,
	validationFunction func(event *corev2.Event) error, executeFunction func(event *corev2.Event) error) *Handler {
	handler := &Handler{
		framework: pluginFramework{
			config:                 config,
			options:                options,
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

	handler.framework.SetWorkflow(handler.workflow)
	if err := handler.framework.Init(); err != nil {
		log.Printf("failed to initialize handler plugin: %s", err)
	}

	return handler
}

// NewGoHandler creates a new handler.
// Deprecated: use NewHandler
var NewGoHandler = NewHandler

// NewEnterpriseHandler is like NewHandler, but requires a valid license.
func NewEnterpriseHandler(config *PluginConfig, options []ConfigOption,
	validationFunction func(event *corev2.Event) error, executeFunction func(event *corev2.Event) error) *Handler {
	handler := &Handler{
		framework: pluginFramework{
			config:                 config,
			options:                options,
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

	handler.framework.SetWorkflow(handler.workflow)
	if err := handler.framework.Init(); err != nil {
		log.Printf("failed to initialize handler plugin: %s", err)
	}

	return handler
}

// NewEnterpriseGoHandler is deprecated, use NewEnterpriseHandler
var NewEnterpriseGoHandler = NewEnterpriseHandler

// Executes the handler's workflow
func (h *Handler) workflow(_ []string) (int, error) {
	event := h.framework.GetStdinEvent()
	if h.enterprise {
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
	err := h.validationFunction(event)
	if err != nil {
		return 1, ErrValidationFailed(err.Error())
	}

	// Execute handler logic using executeFunction
	err = h.executeFunction(event)
	if err != nil {
		return 1, fmt.Errorf("error executing handler: %s", err)
	}

	return 0, nil
}

// Execute is the handler's entry point.
func (h *Handler) Execute() {
	h.framework.Execute()
}

// Disable Handler Event read
func (h *Handler) DisableReadEvent() {
	h.framework.SetEventRead(false)
}

// Disable Handler Event validation
func (h *Handler) DisableEventValidation() {
	h.framework.SetEventValidation(false)
}
