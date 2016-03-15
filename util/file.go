package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

func SaveDeploymentFile(directory, filename, contents string, filemode os.FileMode) error {
	return ioutil.WriteFile(
		path.Join(directory, filename),
		[]byte(contents),
		filemode)
}

func SaveDeploymentMap(directory, filename string, mapcontents map[string]interface{}, filemode os.FileMode) error {
	contents, err := json.MarshalIndent(mapcontents, "", "  ")
	if err != nil {
		return err
	}

	return SaveDeploymentFile(directory, filename, string(contents), filemode)
}
