package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
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

func init() {
	mustRead := func(f string) string {
		bytes, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err)
		}
		contents := string(bytes)
		return contents
	}

	masterScript = mustRead("templates/coreos/master-cloudconfig.in.yml")
	nodeScript = mustRead("templates/coreos/node-cloudconfig.in.yml")

	var err error

	myriadTemplate, err =
		template.New("myriadTemplate").Parse(mustRead("templates/coreos/azdeploy.in.json"))
	if err != nil {
		panic(err)
	}

	myriadParametersTemplate, err =
		template.New("myriadParameters").
			Funcs(template.FuncMap{"b64": b64}).
			Parse(mustRead("templates/coreos/parameters.in.json"))
	if err != nil {
		panic(err)
	}

	utilTemplate, err =
		template.New("util").
			Parse(mustRead("templates/coreos/util.in.sh"))
	if err != nil {
		panic(err)
	}
}

func ProduceUtilScript(flavorArgs FlavorArguments) (utilScript string, err error) {
	log.Info("template: populating util template")
	buf := bytes.Buffer{}

	err = utilTemplate.Execute(&buf, flavorArgs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ProduceTemplateAndParameters(flavorArgs FlavorArguments) (template, parameters map[string]interface{}, err error) {
	log.Info("template: preparing master script")
	masterScript, _ := prepareScript(masterScript)

	log.Info("template: preparing node script")
	nodeScript, _ := prepareScript(nodeScript)

	log.Info("template: populating myriad parameters")
	myriadParameters, err := populateTemplate(
		myriadParametersTemplate,
		flavorArgs)
	if err != nil {
		return nil, nil, err
	}

	log.Info("template: populating myriad template")
	myriadTemplate, err := populateTemplate(
		myriadTemplate,
		struct{ MasterScript, NodeScript string }{masterScript, nodeScript})
	if err != nil {
		return nil, nil, err
	}

	return myriadTemplate, myriadParameters, nil
}

func prepareScript(script string) (string, error) {
	if strings.Contains(script, "'") {
		panic("NO SINGLE QUOTES") // TODO(colemick): nicer...
	}

	script = template.JSEscapeString(script)

	script = variableRegex.ReplaceAllString(script, `', variables('$1'), '`)
	script = parameterRegex.ReplaceAllString(script, `', parameters('$1'), '`)

	script = `[base64(concat('` + script + `'))]`

	return script, nil
}

func populateTemplate(t *template.Template, state interface{}) (template map[string]interface{}, err error) {
	var myriadBuf bytes.Buffer
	var myriadMap map[string]interface{}

	err = t.Execute(&myriadBuf, state)
	if err != nil {
		return nil, fmt.Errorf("template: failed to execute template: %q", err)
	}

	err = json.Unmarshal(myriadBuf.Bytes(), &myriadMap)
	if err != nil {
		return nil, fmt.Errorf("template: failed to unmarshal into map: %q", err)
	}

	return myriadMap, nil
}
