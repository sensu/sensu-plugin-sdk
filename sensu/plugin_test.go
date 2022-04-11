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

func TestSetOptionValue_SliceJSON(t *testing.T) {
	finalValue := []string{"def"}
	option := stringSliceOpt
	option.Value = &finalValue
	err := option.SetValue(`["abc"]`)
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc"}, finalValue)
}

func TestSetOptionValue_IntSlice(t *testing.T) {
	finalValue := []int{42}
	option := intSliceOpt
	option.Value = &finalValue
	err := option.SetValue(`[42]`)
	assert.Nil(t, err)
	assert.Equal(t, []int{42}, finalValue)
}

func TestSetOptionValue_Map(t *testing.T) {
	finalValue := map[string]string{"abc": "def"}
	option := stringMapOpt
	option.Value = &finalValue
	err := option.SetValue(`{"abc":"def"}`)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"abc": "def"}, finalValue)

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

func TestSetOptionValueAllow(t *testing.T) {
	var value string
	option := PluginConfigOption[string]{
		Argument:  "foobar",
		Default:   "default",
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Allow:     []string{"allowed"},
		Value:     &value,
	}
	if err := option.SetValue("default"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue("allowed"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(""); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue("nein"); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetOptionvalueRestrict(t *testing.T) {
	var value string
	option := PluginConfigOption[string]{
		Argument:  "foobar",
		Default:   "default",
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Restrict:  []string{"restricted"},
		Value:     &value,
	}
	if err := option.SetValue("good"); err != nil {
		t.Errorf("expected nil error, got %s", err)
	}
	if err := option.SetValue("restricted"); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetOptionValueAllowAndRestrict(t *testing.T) {
	var value string
	option := PluginConfigOption[string]{
		Argument:  "foobar",
		Default:   "default",
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Allow:     []string{"allowed"},
		Restrict:  []string{"allowed"},
		Value:     &value,
	}
	if err := option.SetValue("default"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue("allowed"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(""); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue("nein"); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetSliceOptionValueAllow(t *testing.T) {
	var value []string
	option := SlicePluginConfigOption[string]{
		Argument:  "foobar",
		Default:   []string{"default"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Allow:     []string{"allowed"},
		Value:     &value,
	}
	if err := option.SetValue("default"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`["default"]`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue("allowed"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`["allowed"]`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`[""]`); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue(""); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue("nein"); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue(`["nein"]`); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetSliceOptionvalueRestrict(t *testing.T) {
	var value []string
	option := SlicePluginConfigOption[string]{
		Argument:  "foobar",
		Default:   []string{"default"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Restrict:  []string{"restricted"},
		Value:     &value,
	}
	if err := option.SetValue("good"); err != nil {
		t.Errorf("expected nil error, got %s", err)
	}
	if err := option.SetValue(`["good"]`); err != nil {
		t.Errorf("expected nil error, got %s", err)
	}
	if err := option.SetValue("restricted"); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue(`["restricted"]`); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetSliceOptionValueAllowAndRestrict(t *testing.T) {
	var value []string
	option := SlicePluginConfigOption[string]{
		Argument:  "foobar",
		Default:   []string{"default"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Allow:     []string{"allowed"},
		Restrict:  []string{"allowed"},
		Value:     &value,
	}
	if err := option.SetValue("default"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`["default"]`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue("allowed"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`["allowed"]`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(""); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue(`"[]"`); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue("nein"); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue(`["nein"]`); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetMapOptionValueAllow(t *testing.T) {
	var value map[string]string
	option := MapPluginConfigOption[string]{
		Argument:  "foobar",
		Default:   map[string]string{"key": "default"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Allow:     map[string]string{"key": "allowed"},
		Value:     &value,
	}
	if err := option.SetValue(`{"key":"default"}`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`{"key":"allowed"}`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`{"key":"notallowed"}`); err == nil {
		t.Error("expected non-nil error")
	}
	if err := option.SetValue(`{"quay":"allowed"}`); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetMapOptionvalueRestrict(t *testing.T) {
	var value map[string]string
	option := MapPluginConfigOption[string]{
		Argument:  "foobar",
		Default:   map[string]string{"key": "default"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Restrict:  map[string]string{"key": "restricted"},
		Value:     &value,
	}
	if err := option.SetValue(`{"key":"good"}`); err != nil {
		t.Errorf("expected nil error, got %s", err)
	}
	if err := option.SetValue(`{"key":"restricted"}`); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestSetMapOptionValueAllowAndRestrict(t *testing.T) {
	var value map[string]string
	option := MapPluginConfigOption[string]{
		Argument:  "foobar",
		Default:   map[string]string{"key": "default"},
		Env:       "ENV_1",
		Path:      "path1",
		Shorthand: "d",
		Usage:     "First argument",
		Secret:    true,
		Allow:     map[string]string{"key": "allowed"},
		Restrict:  map[string]string{"key": "allowed"},
		Value:     &value,
	}
	if err := option.SetValue(`{"key":"default"}`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`{"key":"allowed"}`); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := option.SetValue(`{"key":"nein"}`); err == nil {
		t.Error("expected non-nil error")
	}
}
