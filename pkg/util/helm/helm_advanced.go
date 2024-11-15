// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chartutil"
)

/*
* Generate value file
 */
func GenerateValueFile(values interface{}) (valueFileName string, err error) {
	// 1. generate the template file
	valueFile, err := os.CreateTemp(os.TempDir(), "values")
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", valueFile.Name(), err)
		return "", err
	}

	valueFileName = valueFile.Name()
	log.Debugf("Save the values file %s", valueFileName)

	// 2. dump the object into the template file
	err = toYaml(values, valueFile)
	return valueFileName, err
}

// GenerateHelmTemplate generates helm manifests with the given valuesFile.
func GenerateHelmTemplate(name string, namespace string, valuesFile string, chartPath string, options ...string) (templateFileName string, err error) {
	tempName := fmt.Sprintf("%s.yaml", name)
	templateFile, err := os.CreateTemp("", tempName)
	if err != nil {
		return templateFileName, err
	}
	defer templateFile.Close()
	templateFileName = templateFile.Name()

	values, err := chartutil.ReadValuesFile(valuesFile)
	if err != nil {
		return templateFileName, fmt.Errorf("failed to read values from file %s: %v", valuesFile, err)
	}

	release, err := Template(name, namespace, chartPath, values)
	if err != nil {
		return templateFileName, fmt.Errorf("failed to generate helm manifests %s: %v", name, err)
	}

	_, err = templateFile.WriteString(release.Manifest)
	if err != nil {
		return templateFileName, fmt.Errorf("failed to write helm manifests to file %s: %v", templateFileName, err)
	}

	return templateFileName, nil
}

func GetChartName(chart string) string {
	return filepath.Base(chart)
}
