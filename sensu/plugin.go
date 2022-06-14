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
	"reflect"
	"strings"
	"unicode"

	"github.com/google/go-cmp/cmp"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64 | ~float32 | ~float64 | ~bool | ~string
}

// SliceOptionValue is like OptionValue but applies to SlicePluginConfigOption.
type SliceOptionValue interface {
	~int | ~int32 | ~int64 | ~uint | ~float32 | ~float64 | ~bool | ~string
}

// MapOptionValue is like OptionValue but applies to MapPluginConfigOption.
type MapOptionValue interface {
	~int | ~int64 | ~string
}

// SlicePluginConfigOption is like PluginConfigOption but works with slices of T.
type SlicePluginConfigOption[T SliceOptionValue] struct {
	// Value is the value to read the configured flag or environment variable into.
	// It's expected that Value is non-nil.
	Value *[]T

	// Path is the path to the Sensu annotation to consult when parsing config.
	Path string

	// Env is the environment variable to consult when parsing config.
	Env string

	// Argument is the command line argument to consult when parsing config.
	Argument string

	// Shorthand is the shorthand command line argument to consult when parsing config.
	Shorthand string

	// Default is the default value of the config option.
	Default []T

	// Usage adds help context to the command-line flag.
	Usage string

	// If secret option do not copy Argument value into Default
	Secret bool

	// Restrict prevents the values listed here from being used. If Restrict
	// and Allow are both set, the value set of Restrict is ignored. If the
	// Default value is in the Restrict set, then setting the option value is
	// required.
	Restrict []T

	// Allow prevents using values other than the ones listed here. The Default
	// value is always implicitly part of the allowed set, regardless of what
	// the value of Allow is. If Allow is set, then the values listed in
	// Restrict are ignored. If Allow is not set, or is set to an empty map,
	// then Restrict is consulted.
	Allow []T

	// UseCobraStringArray instructs cobra to use the StringArrayVarP call
	// instead of the StringSliceVarP call. This enables slightly different
	// behaviour: space and comma separated values are considered single values,
	// not lists to be split by cobra.
	//
	// Cobra only supports this mode of operation for lists of strings. If
	// the SlicePluginConfigOption is instantiated with a non-string type, then
	// this option will have no effect.
	UseCobraStringArray bool
}

// MapPluginConfigOption is like PluginConfigOption, but permits using maps.
// The map keys are strings.
type MapPluginConfigOption[T MapOptionValue] struct {
	// Value is the value to read the configured flag or environment variable into.
	// It's expected that Value is non-nil
	Value *map[string]T

	// Path is the path to the Sensu annotation to consult when parsing config.
	Path string

	// Env is the environment variable to consult when parsing config.
	Env string

	// Argument is the command line argument to consult when parsing config.
	Argument string

	// Shorthand is the shorthand command line argument to consult when parsing config.
	Shorthand string

	// Default is the default value of the config option.
	Default map[string]T

	// Usage adds help context to the command-line flag.
	Usage string

	// If secret option do not copy Argument value into Default
	Secret bool

	// Restrict prevents the values listed here from being used. If Restrict
	// and Allow are both set, the value set of Restrict is ignored. If the
	// Default value is in the Restrict set, then setting the option value is
	// required.
	Restrict map[string]T

	// Allow prevents using values other than the ones listed here. The Default
	// value is always implicitly part of the allowed set, regardless of what
	// the value of Allow is. If Allow is set, then the values listed in
	// Restrict are ignored. If Allow is not set, or is set to an empty map,
	// then Restrict is consulted.
	Allow map[string]T
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

	// Restrict prevents the values listed here from being used. If Restrict
	// and Allow are both set, the value set of Restrict is ignored. If the
	// Default value is in the Restrict set, then setting the option value is
	// required.
	Restrict []T

	// Allow prevents using values other than the ones listed here. The Default
	// value is always implicitly part of the allowed set, regardless of what
	// the value of Allow is. If Allow is set, then the values listed in
	// Restrict are ignored. If Allow is not set, or is set to an empty slice,
	// then Restrict is consulted.
	Allow []T
}

// PluginConfig defines the base plugin configuration.
type PluginConfig struct {
	Name     string
	Short    string
	Timeout  uint64
	Keyspace string
}

// pluginFramework defines the basic configuration to be used by all plugin types.
type pluginFramework struct {
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

func (p *pluginFramework) SetWorkflow(f func([]string) (int, error)) {
	p.pluginWorkflowFunction = f
}

func (p *pluginFramework) readSensuEvent() error {
	eventJSON, err := ioutil.ReadAll(p.eventReader)
	if err != nil {
		if p.eventMandatory {
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
	if p.eventValidation {
		if err = validateEvent(sensuEvent); err != nil {
			return err
		}
	}

	p.sensuEvent = sensuEvent
	return nil
}

// Init sets up the framework's configuration parsing and execution environment.
func (p *pluginFramework) Init() error {
	if p.pluginWorkflowFunction == nil {
		return errors.New("workflow function is nil")
	}
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

func (p *pluginFramework) setupFlags(cmd *cobra.Command) error {
	for _, opt := range p.options {
		if err := opt.SetupFlag(cmd); err != nil {
			return err
		}
	}
	return nil
}

func kebabCase(snakeCase string) string {
	var result strings.Builder
	result.Grow(2 * len(snakeCase))
	for i, run3 := range snakeCase {
		if unicode.IsUpper(run3) && i > 0 {
			result.WriteRune('-')
		}
		if unicode.IsUpper(run3) {
			result.WriteRune(unicode.ToLower(run3))
		} else {
			result.WriteRune(run3)
		}
	}
	return result.String()
}

// SetupFlag sets up the option's command line flag, and also binds the
// associated environment variable, and default value.
func (p *PluginConfigOption[T]) SetupFlag(cmd *cobra.Command) error {
	if len(p.Argument) == 0 {
		return nil
	}
	if p.Value == nil {
		return fmt.Errorf("setup flag: %s: couldn't write into nil value", p.Argument)
	}
	err := viper.BindEnv(p.Argument, p.Env)
	if err != nil {
		return err
	}
	viper.SetDefault(p.Argument, p.Default)
	switch value := (interface{}(p.Value)).(type) {
	case *bool:
		cmd.Flags().BoolVarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetBool(p.Argument), p.Usage)
	case *int:
		cmd.Flags().IntVarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetInt(p.Argument), p.Usage)
	case *int32:
		cmd.Flags().Int32VarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetInt32(p.Argument), p.Usage)
	case *int64:
		cmd.Flags().Int64VarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetInt64(p.Argument), p.Usage)
	case *uint:
		cmd.Flags().UintVarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetUint(p.Argument), p.Usage)
	case *uint32:
		cmd.Flags().Uint32VarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetUint32(p.Argument), p.Usage)
	case *uint64:
		cmd.Flags().Uint64VarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetUint64(p.Argument), p.Usage)
	case *float32:
		cmd.Flags().Float32VarP(value, kebabCase(p.Argument), p.Shorthand, float32(viper.GetFloat64(p.Argument)), p.Usage)
	case *float64:
		cmd.Flags().Float64VarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetFloat64(p.Argument), p.Usage)
	case *map[string]string:
		cmd.Flags().StringToStringVarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetStringMapString(p.Argument), p.Usage)
	case *[]string:
		cmd.Flags().StringSliceVarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetStringSlice(p.Argument), p.Usage)
	case *string:
		cmd.Flags().StringVarP(value, kebabCase(p.Argument), p.Shorthand, viper.GetString(p.Argument), p.Usage)
	default:
		rvalue := reflect.Indirect(reflect.ValueOf(p.Value))
		ptr := rvalue.Addr().Interface()
		switch rvalue.Kind() {
		case reflect.Bool:
			cmd.Flags().BoolVarP(ptr.(*bool), kebabCase(p.Argument), p.Shorthand, viper.GetBool(p.Argument), p.Usage)
		case reflect.Int:
			cmd.Flags().IntVarP(ptr.(*int), kebabCase(p.Argument), p.Shorthand, viper.GetInt(p.Argument), p.Usage)
		case reflect.Int32:
			cmd.Flags().Int32VarP(ptr.(*int32), kebabCase(p.Argument), p.Shorthand, viper.GetInt32(p.Argument), p.Usage)
		case reflect.Int64:
			cmd.Flags().Int64VarP(ptr.(*int64), kebabCase(p.Argument), p.Shorthand, viper.GetInt64(p.Argument), p.Usage)
		case reflect.Uint:
			cmd.Flags().UintVarP(ptr.(*uint), kebabCase(p.Argument), p.Shorthand, viper.GetUint(p.Argument), p.Usage)
		case reflect.Uint32:
			cmd.Flags().Uint32VarP(ptr.(*uint32), kebabCase(p.Argument), p.Shorthand, viper.GetUint32(p.Argument), p.Usage)
		case reflect.Uint64:
			cmd.Flags().Uint64VarP(ptr.(*uint64), kebabCase(p.Argument), p.Shorthand, viper.GetUint64(p.Argument), p.Usage)
		case reflect.Float64:
			cmd.Flags().Float64VarP(ptr.(*float64), kebabCase(p.Argument), p.Shorthand, viper.GetFloat64(p.Argument), p.Usage)
		case reflect.String:
			ptr := reflect.ValueOf(p.Value).Convert(reflect.TypeOf(new(string))).Interface().(*string)
			cmd.Flags().StringVarP(ptr, kebabCase(p.Argument), p.Shorthand, viper.GetString(p.Argument), p.Usage)
		default:
			return fmt.Errorf("setup flag: %s: unknown value type", p.Argument)
		}
	}
	flag := cmd.Flags().Lookup(kebabCase(p.Argument))
	// Set empty DefValue string if option is a secret
	// DefValue is only used for pflag usage message construction
	if p.Secret {
		flag.DefValue = ""
	}
	return nil
}

// integer overflow isn't real, it can't hurt you
func castIntSlice[T int32 | int64 | uint32 | uint64 | uint](values []int) []T {
	result := make([]T, len(values))
	for i := range values {
		result[i] = T(values[i])
	}
	return result
}

// SetupFlag sets up the option's command line flag, and also binds the
// associated environment variable, and default value.
func (p *SlicePluginConfigOption[T]) SetupFlag(cmd *cobra.Command) error {
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
	viper.SetDefault(p.Argument, p.Default)
	switch value := (interface{}(p.Value)).(type) {
	case *[]bool:
		cmd.Flags().BoolSliceVarP(value, p.Argument, p.Shorthand, nil, p.Usage) // FIXME: viper lacks GetBoolSlice function
	case *[]int:
		cmd.Flags().IntSliceVarP(value, p.Argument, p.Shorthand, viper.GetIntSlice(p.Argument), p.Usage)
	case *[]int32:
		cmd.Flags().Int32SliceVarP(value, p.Argument, p.Shorthand, castIntSlice[int32](viper.GetIntSlice(p.Argument)), p.Usage)
	case *[]int64:
		cmd.Flags().Int64SliceVarP(value, p.Argument, p.Shorthand, castIntSlice[int64](viper.GetIntSlice(p.Argument)), p.Usage)
	case *[]uint:
		cmd.Flags().UintSliceVarP(value, p.Argument, p.Shorthand, castIntSlice[uint](viper.GetIntSlice(p.Argument)), p.Usage)
	case *[]float32:
		cmd.Flags().Float32SliceVarP(value, p.Argument, p.Shorthand, nil, p.Usage) // FIXME: viper lacks GetFloatSlice function
	case *[]float64:
		cmd.Flags().Float64SliceVarP(value, p.Argument, p.Shorthand, nil, p.Usage) // FIXME: viper lacks GetFloatSlice function
	case *[]string:
		if p.UseCobraStringArray {
			cmd.Flags().StringArrayVarP(value, p.Argument, p.Shorthand, viper.GetStringSlice(p.Argument), p.Usage)
		} else {
			cmd.Flags().StringSliceVarP(value, p.Argument, p.Shorthand, viper.GetStringSlice(p.Argument), p.Usage)
		}
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

func castMap[T int | int64](m map[string]interface{}) map[string]T {
	result := make(map[string]T, len(m))
	for k, v := range m {
		if value, ok := v.(T); ok {
			result[k] = value
		}
	}
	return result
}

// SetupFlag sets up the option's command line flag, and also binds the
// associated environment variable, and default value.
func (p *MapPluginConfigOption[T]) SetupFlag(cmd *cobra.Command) error {
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
	viper.SetDefault(p.Argument, p.Default)

	switch value := (interface{}(p.Value)).(type) {
	case *map[string]int:
		cmd.Flags().StringToIntVarP(value, p.Argument, p.Shorthand, castMap[int](viper.GetStringMap(p.Argument)), p.Usage)
	case *map[string]int64:
		cmd.Flags().StringToInt64VarP(value, p.Argument, p.Shorthand, castMap[int64](viper.GetStringMap(p.Argument)), p.Usage)
	case *map[string]string:
		cmd.Flags().StringToStringVarP(value, p.Argument, p.Shorthand, viper.GetStringMapString(p.Argument), p.Usage)
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

// GetStdinEvent gets the event that was received on stdin, if any. Can return
// nil values.
func (p *pluginFramework) GetStdinEvent() *corev2.Event {
	return p.sensuEvent
}

type allowRestrictValidator interface {
	validateAllowRestrict() error
}

// cobraExecuteFunction is called by the argument's execute. The configuration overrides will be processed if necessary
// and the pluginWorkflowFunction function executed
func (p *pluginFramework) cobraExecuteFunction(args []string) error {
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

	for _, option := range p.options {
		if v, ok := option.(allowRestrictValidator); ok {
			if err := v.validateAllowRestrict(); err != nil {
				p.exitStatus = p.errorExitStatus
				return err
			}
		}
	}

	exitStatus, err := p.pluginWorkflowFunction(args)
	p.exitStatus = exitStatus

	return err
}

// Execute executes the plugin. Check, Handler, and Mutator all call this in
// their own Execute functions.
func (p *pluginFramework) Execute() {
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

// SetValue sets the configuration value based on either a raw string or
// json-encoded string.
func (p *PluginConfigOption[T]) SetValue(valueStr string) (err error) {
	if p.Value == nil {
		return errors.New("PluginConfigOption.Value not set!")
	}
	defer func() {
		if err == nil {
			err = p.validateAllowRestrict()
		}
	}()
	if err := json.Unmarshal([]byte(valueStr), p.Value); err == nil {
		return nil
	}
	if value, ok := (interface{}(p.Value)).(*string); ok {
		*value = valueStr
		return nil
	}
	return fmt.Errorf("invalid value for %T: %v", *p.Value, valueStr)
}

// SetValue sets the configuration value based on either a raw string or
// json-encoded string.
func (p *SlicePluginConfigOption[T]) SetValue(valueStr string) (err error) {
	if p.Value == nil {
		return errors.New("PluginConfigOption.Value not set!")
	}
	defer func() {
		if err == nil {
			err = p.validateAllowRestrict()
		}
	}()
	if err := json.Unmarshal([]byte(valueStr), p.Value); err == nil {
		return nil
	}
	if slice, ok := ((interface{})(p.Value)).(*[]string); ok {
		*slice = []string{valueStr}
		return nil
	}
	var t T
	if err := json.Unmarshal([]byte(valueStr), &t); err == nil {
		*p.Value = []T{t}
		return nil
	}
	return fmt.Errorf("invalid value for %T: %v", *p.Value, valueStr)
}

// SetValue sets the configuration value based on a json-encoded string.
func (p *MapPluginConfigOption[T]) SetValue(valueStr string) (err error) {
	if p.Value == nil {
		return errors.New("PluginConfigOption.Value not set!")
	}
	defer func() {
		if err == nil {
			err = p.validateAllowRestrict()
		}
	}()
	return json.Unmarshal([]byte(valueStr), p.Value)
}

func (p *MapPluginConfigOption[T]) validateAllowRestrict() error {
	if len(p.Allow) > 0 {
		for k, v := range *p.Value {
			if p.Allow[k] != v && p.Default[k] != v {
				return fmt.Errorf("%s: key %v = value %v not one of %v", p.Argument, k, v, p.Allow)
			}
		}
		return nil
	}

	for k, v := range *p.Value {
		if p.Restrict[k] == v {
			return fmt.Errorf("%s: key %v = value %v not allowed", p.Argument, k, v)
		}
	}
	return nil
}

func (p *SlicePluginConfigOption[T]) validateAllowRestrict() error {
	if len(p.Allow) > 0 {
		allow := append(p.Allow, p.Default...)
		for _, pvalue := range *p.Value {
			var found bool
			for _, value := range allow {
				if value == pvalue {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%s: value %v not one of %v", p.Argument, pvalue, p.Allow)
			}
		}
		return nil
	}
	for _, pvalue := range *p.Value {
		for _, value := range p.Restrict {
			if value == pvalue {
				return fmt.Errorf("%s: value not allowed to be %v", p.Argument, value)
			}
		}
	}
	return nil
}

func (p *PluginConfigOption[T]) validateAllowRestrict() error {
	if len(p.Allow) > 0 {
		allow := append(p.Allow, p.Default)
		for _, value := range allow {
			if cmp.Equal(value, *p.Value) {
				return nil
			}
		}
		return fmt.Errorf("%s: value not one of %v", p.Argument, p.Allow)
	}
	for _, value := range p.Restrict {
		if cmp.Equal(value, *p.Value) {
			return fmt.Errorf("%s: value not allowed to be %v", p.Argument, value)
		}
	}

	return nil
}

// SetAnnotationValue sets the option value based on a prefix indicated by
// keyspace, and an event object. The check annotation will be resolved first,
// followed by the entity annotation.
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

// SetAnnotationValue sets the option value based on a prefix indicated by
// keyspace, and an event object. The check annotation will be resolved first,
// followed by the entity annotation.
func (p *SlicePluginConfigOption[T]) SetAnnotationValue(keySpace string, event *corev2.Event) (SetAnnotationResult, error) {
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

// SetAnnotationValue sets the option value based on a prefix indicated by
// keyspace, and an event object. The check annotation will be resolved first,
// followed by the entity annotation.
func (p *MapPluginConfigOption[T]) SetAnnotationValue(keySpace string, event *corev2.Event) (SetAnnotationResult, error) {
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
