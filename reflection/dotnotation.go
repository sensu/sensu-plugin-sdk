package reflection

import (
	"fmt"
	"reflect"
	"strings"
)

type nameValuePair struct {
	key   string
	value string
}

// DotNotation serializes an interface into a dot-notation based key/value pairs. As an example
// {abc: 123, def: {ghi: 456, jkl:"789"}} will be converted into the following entries:
// abc = 123
// def.ghi = 456
// def.jkl = 789
func DotNotation(data interface{}) []nameValuePair {
	values := make([]nameValuePair, 0, 64)
	return dotNotationImpl(data, values, "")
}

func dotNotationImpl(data interface{}, values []nameValuePair, prefix string) []nameValuePair {
	value := reflect.Indirect(reflect.ValueOf(data))

	if !value.IsValid() {
		return append(values, nameValuePair{prefix, "nil"})
	}

	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			fieldName := value.Type().Field(i).Name
			fieldValue := field.Interface()
			values = dotNotationImpl(fieldValue, values, buildKey(prefix, fieldName))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			sliceItem := value.Index(i).Interface()
			values = dotNotationImpl(sliceItem, values, fmt.Sprintf("%s[%d]", prefix, i))
		}
	case reflect.Ptr:
		if value.IsNil() {
			values = append(values, nameValuePair{prefix, "nil"})
		} else {
			valueStr := fmt.Sprintf("%v", value.Interface())
			values = append(values, nameValuePair{prefix, valueStr})
		}
	default:
		valueStr := fmt.Sprintf("%v", value.Interface())
		values = append(values, nameValuePair{prefix, valueStr})
	}

	return values
}

func buildKey(prefix string, postfix string) string {
	if len(prefix) > 0 {
		return prefix + "." + postfix
	} else {
		return postfix
	}
}

func DotNotationToString(pairs []nameValuePair, separator string) string {
	var stringBuilder strings.Builder
	for i := 0; i < len(pairs); i++ {
		stringBuilder.WriteString(pairs[i].key)
		stringBuilder.WriteString(separator)
		stringBuilder.WriteString(pairs[i].value)
		stringBuilder.WriteString("\n")
	}
	return stringBuilder.String()
}
