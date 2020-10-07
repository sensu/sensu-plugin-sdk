# Sensu Go Plugin Library

[![GoDoc](https://godoc.org/github.com/sensu-community/sensu-plugin-sdk?status.svg)](https://godoc.org/github.com/sensu-community/sensu-plugin-sdk)
![Go Test](https://github.com/sensu-community/sensu-plugin-sdk/workflows/Go%20Test/badge.svg)

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

### Annotations Configuration Options Override

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

## Enterprise plugins

An enterprise plugin requires a valid Sensu license to run. Initialize enterprise handlers with
`NewEnterpriseGoHandler`. If the license file passed from the handler's environment variables is
invalid, it should return an error without executing.

```Go
func main() {
  goHandler := sensu.NewEnterpriseGoHandler(&config.HandlerConfig, options, validateInput, executeHandler)
  err := goHandler.Execute()
}

```

Sensu Go >= 5.21 will add the `SENSU_LICENSE_FILE` environment variable to the handler execution.
To run the plugin independently of Sensu (ex. test/dev), you must set the env var:

```
SENSU_LICENSE_FILE=$(sensuctl license info --format json)
```

## Templates

The templates package provides a wrapper to the [`text/template`][1] package
allowing for the use of templates to expand event attributes.  An example of
this would be using the following as part of a handler:

```
--summary-template "{{.Entity.Name}}/{{.Check.Name}}"
```

Which, if given an event with an entity name of webserver01 and a check name of
check-nginx would yield `webserver01/check-nginx`.

### UnixTime template function

A Sensu Go event contains multiple timestamps (e.g. .Check.Issued,
.Check.Executed, .Check.LastOk) that are presented in UNIX timestamp format.  A
function named UnixTime is provided to print these values in a customizable
human readable format as part of a template.  To customize the output format of
the timestamp, use the same format as specified by Golang's [Time.Format][2].
Additional examples can be found [here][3].

**Note:** the predefined format constants are **not** available.

The example below demonstrates its use:

```
[...]
Service: {{.Entity.Name}}/{{.Check.Name}}
Executed: {{(UnixTime .Check.Executed).Format "2 Jan 2006 15:04:05"}}
Last OK: {{(UnixTime .Check.LastOK).Format "2 Jan 2006 15:04:05"}}
[...]
```

[1]: https://golang.org/pkg/text/template/
[2]: https://golang.org/pkg/time/#Time.Format
[3]: https://yourbasic.org/golang/format-parse-string-time-date-example/
