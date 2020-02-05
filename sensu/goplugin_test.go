package sensu

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetOptionValue_String(t *testing.T) {
	finalValue := ""
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.Nil(t, err)
	assert.Equal(t, "abc", finalValue)
}

func TestSetOptionValue_EmptyString(t *testing.T) {
	finalValue := ""
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "")
	assert.Nil(t, err)
	assert.Equal(t, "", finalValue)
}

func TestSetOptionValue_ValidUint64(t *testing.T) {
	var finalValue uint64
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "123")
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), finalValue)
}

func TestSetOptionValue_InvalidUint64(t *testing.T) {
	var finalValue uint64
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), finalValue)
}

func TestSetOptionValue_TrueBool(t *testing.T) {
	var finalValue bool
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "true")
	assert.Nil(t, err)
	assert.Equal(t, true, finalValue)
}

func TestSetOptionValue_FalseBool(t *testing.T) {
	finalValue := true
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "false")
	assert.Nil(t, err)
	assert.Equal(t, false, finalValue)
}

func TestSetOptionValue_Int64(t *testing.T) {
	var value int64
	option := defaultOption1
	option.Value = &value
	assert.NoError(t, setOptionValue(&option, "42"))
	assert.Equal(t, value, int64(42))
}

func TestSetOptionValue_InvalidBool(t *testing.T) {
	var finalValue bool
	option := defaultOption1
	option.Value = &finalValue
	err := setOptionValue(&option, "yes")
	assert.NotNil(t, err)
	assert.Equal(t, false, finalValue)
}

func getFileReader(file string) io.Reader {
	reader, _ := os.Open(file)
	return reader
}

func clearEnvironment() {
	_ = os.Unsetenv("ENV_1")
	_ = os.Unsetenv("ENV_2")
	_ = os.Unsetenv("ENV_3")
}

func Test_setupFlag(t *testing.T) {
	var foo string

	tests := []struct {
		name           string
		option         *PluginConfigOption
		wantExecuteErr bool
		wantSetupErr   bool
	}{
		{
			name: "Missing required flag should return an error",
			option: &PluginConfigOption{
				Argument: "foo",
				Env:      "FOO",
				Value:    &foo,
				Required: true,
			},
			wantExecuteErr: true,
		},
		{
			name: "Missing optional flag should not return an error",
			option: &PluginConfigOption{
				Argument: "foo",
				Env:      "FOO",
				Value:    &foo,
				Required: false,
			},
			wantExecuteErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			if err := setupFlag(cmd, tt.option); (err != nil) != tt.wantSetupErr {
				t.Fatalf("setupFlag() error = %#v, wantErr %v", err, tt.wantSetupErr)
			}

			cmd.SetOutput(ioutil.Discard)
			if err := cmd.Execute(); (err != nil) != tt.wantExecuteErr {
				t.Fatalf("cmd.Execute() error = %#v, wantErr %v", err, tt.wantExecuteErr)
			}
		})
	}
}
