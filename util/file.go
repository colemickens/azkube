package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/viper"
)

func SaveDeploymentFile(filename string, contents string, filemode os.FileMode) error {
	return ioutil.WriteFile(
		path.Join(
			viper.GetString(deployArgNames.OutputDirectory),
			filename),
		[]byte(contents),
		filemode)
}

func SaveDeploymentMap(filename string, mapcontents map[string]interface{}, filemode os.FileMode) error {
	contents, err := json.MarshalIndent(mapcontents, "", "  ")
	if err != nil {
		return err
	}

	return SaveDeploymentFile(filename, string(contents), filemode)
}
