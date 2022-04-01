package sensu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GoPlugin defines the GoPlugin interface to be implemented by all types of plugins
type GoPlugin interface {
	Execute()
}

// SetAnnotationResult is returned by SetAnnotation, and indicates what kind of
// setter action was taken, if any.
type SetAnnotationResult struct {
	AnnotationKey    string
	AnnotationValue  string
	CheckAnnotation  bool
	EntityAnnotation bool
}

// ConfigOption is an interface. It exists so that users can create slices of
// configuration options with different types. For instance,
// []ConfigOption{&PluginConfigOption[int]{}, &PluginConfigOption[string]{}}
type ConfigOption interface {
	SetupFlag(*cobra.Command) error
	SetValue(string) error
	SetAnnotationValue(keySpace string, e *corev2.Event) (SetAnnotationResult, error)
}

// OptionValue is a type constraint that creates a compile-time guard against
// creating a PluginConfigOption with an unsupported data type.
type OptionValue interface {
	~int |  ~int32 | ~int64 | ~uint | ~uint32 | ~uint64 | ~float32 | ~float64 | ~bool | ~string | ~map[string]string | ~[]string
}

// PluginConfigOption defines an option to be read by the plugin on startup. An
// option can be passed using a command line argument, an environment variable or
// at for some plugin types using a configuration override from the Sensu event.
type PluginConfigOption[T OptionValue] struct {
	// Value is the value to read the configured flag or environment variable into.
	// Pass a pointer to any value in your plugin in order to fill it in with the
	// data from a flag or environment variable. The parsing will be done with
	// a function supplied by viper. See the viper documentation for details on
	// how various data types are parsed.
	Value *T

	// Path is the path to the Sensu annotation to consult when parsing config.
	Path string

	// Env is the environment variable to consult when parsing config.
	Env string

	// Argument is the command line argument to consult when parsing config.
	Argument string

	// Shorthand is the shorthand command line argument to consult when parsing config.
	Shorthand string

	// Default is the default value of the config option.
	Default T

	// Usage adds help context to the command-line flag.
	Usage string

	// If secret option do not copy Argument value into Default
	Secret bool
}

// PluginConfig defines the base plugin configuration.
type PluginConfig struct {
	Name     string
	Short    string
	Timeout  uint64
	Keyspace string
}

// basePlugin defines the basic configuration to be used by all plugin corev2.
type basePlugin struct {
	config                 *PluginConfig
	options                []ConfigOption
	sensuEvent             *corev2.Event
	eventReader            io.Reader
	pluginWorkflowFunction func([]string) (int, error)
	cmd                    *cobra.Command
	readEvent              bool
	eventMandatory         bool
	eventValidation        bool
	configurationOverrides bool
	verbose                bool
	exitStatus             int
	errorExitStatus        int
	exitFunction           func(int)
	errorLogFunction       func(format string, a ...interface{})
}

func (goPlugin *basePlugin) readSensuEvent() error {
	eventJSON, err := ioutil.ReadAll(goPlugin.eventReader)
	if err != nil {
		if goPlugin.eventMandatory {
			return fmt.Errorf("failed to read stdin: %s", err)
		} else {
			// if event is not mandatory return without going any further
			return nil
		}
	}

	sensuEvent := &corev2.Event{}
	err = json.Unmarshal(eventJSON, sensuEvent)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin event: %s", err)
	}
	if goPlugin.eventValidation {
		if err = validateEvent(sensuEvent); err != nil {
			return err
		}
	}

	goPlugin.sensuEvent = sensuEvent
	return nil
}

func (p *basePlugin) initPlugin() error {
	p.cmd = &cobra.Command{
		Use:           p.config.Name,
		Short:         p.config.Short,
		SilenceErrors: true,
	}
	p.cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		err := p.cobraExecuteFunction(args)
		if _, ok := err.(ErrValidationFailed); !ok {
			p.cmd.SilenceUsage = true
		} else {
			err = fmt.Errorf("error validating input: %s", err)
		}
		return err
	}
	p.exitFunction = os.Exit
	p.errorLogFunction = func(format string, a ...interface{}) {
		_, _ = fmt.Fprintf(os.Stderr, format, a...)
	}

	p.cmd.AddCommand(&cobra.Command{
		Use:           "version",
		Short:         "Print the version number of this plugin",
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version())
		},
	})

	return p.setupFlags(p.cmd)
}

func (p *basePlugin) setupFlags(cmd *cobra.Command) error {
	for _, opt := range p.options {
		if err := opt.SetupFlag(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (p *PluginConfigOption[T]) SetupFlag(cmd *cobra.Command) error {
	if len(p.Argument) == 0 {
		return nil
	}
	if p.Value == nil {
		return errors.New("setup flag: couldn't write into nil value")
	}
	err := viper.BindEnv(p.Argument, p.Env)
	if err != nil {
		return err
	}
	switch value := (interface{}(p.Value)).(type) {
	case *bool:
		cmd.Flags().BoolVarP(value, p.Argument, p.Shorthand, viper.GetBool(p.Argument), p.Usage)
	case *int:
		cmd.Flags().IntVarP(value, p.Argument, p.Shorthand, viper.GetInt(p.Argument), p.Usage)
	case *int32:
		cmd.Flags().Int32VarP(value, p.Argument, p.Shorthand, viper.GetInt32(p.Argument), p.Usage)
	case *int64:
		cmd.Flags().Int64VarP(value, p.Argument, p.Shorthand, viper.GetInt64(p.Argument), p.Usage)
	case *uint:
		cmd.Flags().UintVarP(value, p.Argument, p.Shorthand, viper.GetUint(p.Argument), p.Usage)
	case *uint32:
		cmd.Flags().Uint32VarP(value, p.Argument, p.Shorthand, viper.GetUint32(p.Argument), p.Usage)
	case *uint64:
		cmd.Flags().Uint64VarP(value, p.Argument, p.Shorthand, viper.GetUint64(p.Argument), p.Usage)
	case *float32:
		cmd.Flags().Float32VarP(value, p.Argument, p.Shorthand, float32(viper.GetFloat64(p.Argument)), p.Usage)
	case *float64:
		cmd.Flags().Float64VarP(value, p.Argument, p.Shorthand, viper.GetFloat64(p.Argument), p.Usage)
	case *map[string]string:
		cmd.Flags().StringToStringVarP(value, p.Argument, p.Shorthand, viper.GetStringMapString(p.Argument), p.Usage)
	case *[]string:
		cmd.Flags().StringSliceVarP(value, p.Argument, p.Shorthand, viper.GetStringSlice(p.Argument), p.Usage)
	case *string:
		cmd.Flags().StringVarP(value, p.Argument, p.Shorthand, viper.GetString(p.Argument), p.Usage)
	default:
		return errors.New("setup flag: unknown value type")
	}
	flag := cmd.Flags().Lookup(p.Argument)
	// Set empty DefValue string if option is a secret
	// DefValue is only used for pflag usage message construction
	if p.Secret {
		flag.DefValue = ""
	}
	return nil
}

// cobraExecuteFunction is called by the argument's execute. The configuration overrides will be processed if necessary
// and the pluginWorkflowFunction function executed
func (p *basePlugin) cobraExecuteFunction(args []string) error {
	// Read the Sensu event if required
	if p.readEvent {
		err := p.readSensuEvent()
		if err != nil {
			p.exitStatus = p.errorExitStatus
			return err
		}
	}

	// If there is an event process configuration overrides if necessary
	if p.sensuEvent != nil && p.configurationOverrides {
		err := configurationOverrides(p.config, p.options, p.sensuEvent, p.verbose)
		if err != nil {
			p.exitStatus = p.errorExitStatus
			return err
		}
	}

	exitStatus, err := p.pluginWorkflowFunction(args)
	p.exitStatus = exitStatus

	return err
}

func (p *basePlugin) Execute() {
	// Validate the cmd is set
	if p.cmd == nil {
		p.errorLogFunction("Error executing %s: Arguments must be initialized\n", p.config.Name)
		p.exitFunction(p.errorExitStatus)
	}

	if err := p.cmd.Execute(); err != nil {
		p.errorLogFunction("Error executing %s: %v\n", p.config.Name, err)
	}

	p.exitFunction(p.exitStatus)
}

func validateEvent(event *corev2.Event) error {
	if event.Timestamp <= 0 {
		return errors.New("timestamp is missing or must be greater than zero")
	}

	return event.Validate()
}

func (p *PluginConfigOption[T]) SetValue(valueStr string) error {
	switch value := (interface{}(p.Value)).(type) {
	case *[]string:
		if err := json.Unmarshal([]byte(valueStr), value); err == nil {
			return nil
		}
		*value = []string{valueStr}
		return nil
	case *string:
		*value = valueStr
		return nil
	default:
		return json.Unmarshal([]byte(valueStr), value)
	}
}

func (p *PluginConfigOption[T]) SetAnnotationValue(keySpace string, event *corev2.Event) (SetAnnotationResult, error) {
	key := path.Join(keySpace, p.Path)
	downcase := strings.ToLower(key)
	keys := []string{downcase, key}
	var result SetAnnotationResult
	for _, key := range keys {
		var value string
		if event.Check != nil {
			value, _ = event.Check.Annotations[key]
			result.CheckAnnotation = len(value) > 0
		}
		if value == "" && event.Entity != nil {
			value, _ = event.Entity.Annotations[key]
			result.EntityAnnotation = len(value) > 0
		}
		if len(value) > 0 {
			result.AnnotationKey = key
			result.AnnotationValue = value
			return result, p.SetValue(value)
		}
	}
	return result, nil
}

func configurationOverrides(config *PluginConfig, options []ConfigOption, event *corev2.Event, verbose bool) error {
	if config.Keyspace == "" {
		return nil
	}
	for _, opt := range options {
		result, err := opt.SetAnnotationValue(config.Keyspace, event)
		if err != nil {
			return err
		}
		if verbose {
			var what string
			if result.CheckAnnotation {
				what = "check"
			} else if result.EntityAnnotation {
				what = "entity"
			} else {
				continue
			}
			msg := "overriding default plugin configuration with value of \"%s.annotations.%s\" (%q)"
			log.Printf(msg, what, result.AnnotationKey, result.AnnotationValue)
		}
	}
	return nil
}
