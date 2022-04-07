package sensu

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetOptionValue_String(t *testing.T) {
	finalValue := ""
	option := stringOpt
	option.Value = &finalValue
	err := option.SetValue("abc")
	assert.Nil(t, err)
	assert.Equal(t, "abc", finalValue)
}

func TestSetOptionValue_EmptyString(t *testing.T) {
	finalValue := ""
	option := stringOpt
	option.Value = &finalValue
	err := option.SetValue("")
	assert.Nil(t, err)
	assert.Equal(t, "", finalValue)
}

func TestSetOptionValue_Slice(t *testing.T) {
	finalValue := []string{"def"}
	option := stringSliceOpt
	option.Value = &finalValue
	err := option.SetValue("abc")
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc"}, finalValue)
}

func TestSetOptionValue_EmptySlice(t *testing.T) {
	finalValue := []string{"def"}
	option := stringSliceOpt
	option.Value = &finalValue
	err := option.SetValue("")
	assert.Nil(t, err)
	assert.Equal(t, []string{""}, finalValue)
}

func TestSetOptionValue_JSONArray(t *testing.T) {
	var finalValue []string
	option := stringSliceOpt
	option.Value = &finalValue
	err := option.SetValue(`["abc","def"]`)
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc", "def"}, finalValue)
}

func TestSetOptionValue_ValidUint64(t *testing.T) {
	var finalValue uint64
	option := uint64Opt
	option.Value = &finalValue
	err := option.SetValue("123")
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), finalValue)
}

func TestSetOptionValue_InvalidUint64(t *testing.T) {
	var finalValue uint64
	option := uint64Opt
	option.Value = &finalValue
	err := option.SetValue("abc")
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), finalValue)
}

func TestSetOptionValue_TrueBool(t *testing.T) {
	var finalValue bool
	option := boolOpt
	option.Value = &finalValue
	err := option.SetValue("true")
	assert.Nil(t, err)
	assert.Equal(t, true, finalValue)
}

func TestSetOptionValue_FalseBool(t *testing.T) {
	finalValue := true
	option := boolOpt
	option.Value = &finalValue
	err := option.SetValue("false")
	assert.Nil(t, err)
	assert.Equal(t, false, finalValue)
}

func TestSetOptionValue_Int64(t *testing.T) {
	var value int64
	option := int64Opt
	option.Value = &value
	assert.NoError(t, option.SetValue("42"))
	assert.Equal(t, value, int64(42))
}

func TestSetOptionValue_InvalidBool(t *testing.T) {
	var finalValue bool
	option := boolOpt
	option.Value = &finalValue
	err := option.SetValue("yes")
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
