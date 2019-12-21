# Sensu Go Plugin Library

[![Build Status](https://img.shields.io/travis/sensu/sensu-plugins-go-library.svg)](https://travis-ci.org/sensu/sensu-plugins-go-library)

This project implements a library to help users build their own Sensu Go plugins.
With this library the user only needs to define the plugin arguments, an input validation function and an execution function.

## Plugin configuration

The plugin configuration contains the plugin information.

To define the plugin configuration use the following information.

```Go
// Define the plugin configuration
type Config struct {
  sensu.HandlerConfig
  // other configuration
}

// Create the plugin configuration
var config = Config{
  HandlerConfig: sensu.HandlerConfig{
    Name:     "sensu-go-plugin",
    Short:    "Performs my incredible logic",
    Timeout:  10,
    Keyspace: "sensu.io/plugins/mysensuplugin/config",
  },
}
```

## Plugin Options

These are the supported option types, in order or priority.
* Sensu Event Check configuration override
* Sensu Event Entity configuration override
* Command line argument in short or long form
* Environment variable

```Go
var (
  argumentValue string

  options = []*sensu.HandlerConfigOption{
    {
      Path:      "override-path",
      Env:       "COMMAND_LINE_ENVIRONMENT",
      Argument:  "command-line-argument",
      Shorthand: "c",
      Default:   "Default Value",
      Usage:     "The usage message printed for this option",
      Value:     &argumentValue,
    },
  }
)
```

## Input Validation Function

The validation function is used to validate the Sensu event and plugin input.
It must return an `error` if there is a problem with the options. If the input
is correct `nil` can be returned and the plugin execution will continue.

To define the validation function use the following signature.

```Go
func validateInput(_ *types.Event) error {
  // Validate the input here
  return nil
}
```

## Execution function

The execution function executes the plugin's logic. If there is an error while processing the plugin logic the execution function should return an `error`. If
the logic is executed successfully `nil` should be returned.

To define the execution function use the following signature.

```Go
func executeHandler(event *types.Event) error {
  // Handler logic
  return nil
}
```

## Putting Everything Together

Create a main function that creates the handler with the previously defined configuration,
options, validation function and execution function.

```Go
func main() {
  handler := sensu.InitHandler(&config.HandlerConfig, options, validateInput, executeHandler)
  err := handler.Execute()
}

```
