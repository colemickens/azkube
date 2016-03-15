package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	myriadTemplate           *template.Template
	myriadParametersTemplate *template.Template
	utilTemplate             *template.Template
	masterScript             string
	nodeScript               string

	variableRegex  = regexp.MustCompile(`\[\[\[([a-zA-Z]+)\]\]\]`)
	parameterRegex = regexp.MustCompile(`\{\{\{([a-zA-Z]+)\}\}\}`)
)

func b64(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

func InterpolateArmPlaceholders(flavor, scriptName string) (string, error) {
	scriptPath := filepath.Join("templates", flavor, scriptName)
	scriptBytes, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return "", nil
	}

	script := string(scriptBytes)
	if strings.Contains(script, "'") {
		return "", fmt.Errorf("template: failed to populate due to presence of single quote: %q", scriptPath)
	}

	script = template.JSEscapeString(script)

	script = variableRegex.ReplaceAllString(script, `', variables('$1'), '`)
	script = parameterRegex.ReplaceAllString(script, `', parameters('$1'), '`)

	script = `[base64(concat('` + script + `'))]`

	return script, nil
}

func PopulateTemplate(flavor, templateFilename string, state interface{}) (outputString string, err error) {
	templatePath := filepath.Join("templates", flavor, templateFilename)

	templateBytes, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	t, err := template.
		New(templatePath).
		Funcs(template.FuncMap{"b64": b64}).
		Parse(string(templateBytes))

	var outputBuffer bytes.Buffer

	err = t.Execute(&outputBuffer, state)
	if err != nil {
		return "", fmt.Errorf("template: failed to execute template (%q): %q", templatePath, err)
	}

	return string(outputBuffer.Bytes()), nil
}

func PopulateTemplateMap(flavor, templateFilename string, state interface{}) (outputMap map[string]interface{}, err error) {
	outputString, err := PopulateTemplate(flavor, templateFilename, state)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(outputString), &outputMap)
	if err != nil {
		return nil, fmt.Errorf("template: failed to unmarshal into map: %q", err)
	}

	return outputMap, nil
}
