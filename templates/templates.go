package templates

import (
	"bytes"
	"fmt"
	"text/template"
)

func EvalTemplate(templName, templStr string, templSrc interface{}) (string, error) {
	if templSrc == nil {
		return "", fmt.Errorf("must pass in template source")
	}
	if len(templStr) == 0 {
		return "", fmt.Errorf("must pass in template")
	}

	templ, err := template.New(templName).Parse(templStr)
	if err != nil {
		return "", fmt.Errorf("Error building template: %s", err)
	}

	buf := new(bytes.Buffer)
	err = templ.Execute(buf, templSrc)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %s", err)
	}

	return buf.String(), nil
}
