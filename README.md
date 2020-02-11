# Sensu Go Plugin Library

[![GoDoc](https://godoc.org/github.com/sensu-community/sensu-plugin-sdk?status.svg)](https://godoc.org/github.com/sensu-community/sensu-plugin-sdk)
[![Build Status](https://img.shields.io/travis/sensu-community/sensu-plugin-sdk.svg)](https://travis-ci.org/sensu-community/sensu-plugin-sdk)

This project is a framework for building Sensu Go plugins. Plugins can be Checks, Handlers, or Mutators.
With this library the user only needs to define the plugin arguments, an input validation function and an execution function.

## Plugin Configuration

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
    Keyspace: "sensu.io/plugins/my-sensu-go-plugin/config",
  },
}
```

## Plugin Configuration Options

Configuration options are read from the following sources using the following precedence order. Each item takes precedence over the item below it:

* Sensu event check annotation
* Sensu event entity annotation
* Command line argument in short or long form
* Environment variable
* Default value

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

### Annotations configuration options override

Configuration options can be overridden using the Sensu event check or entity annotations.

For example, if we have a plugin using the **keyspace** `sensu.io/plugins/my-sensu-go-plugin/config` and a configuration option using the **path** `node-name`, the following annotation could be configured in an agent configuration file to override whatever value is configuration via the plugin's flags or environment variables:

```yaml
# /etc/sensu/agent.yml example
annotations:
  sensu.io/plugins/my-sensu-go-plugin/config/node-name: webserver01.example.com
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
  goHandler := sensu.NewGoHandler(&config.HandlerConfig, options, validateInput, executeHandler)
  err := goHandler.Execute()
}

```
