// Copyright 2024 The Kubeflow Authors.
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
	"time"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

const (
	WaitTimeout = 5 * time.Minute
)

func getActionConfig(namespace string) (*action.Configuration, error) {
	envSettings := cli.New()
	envSettings.SetNamespace(namespace)
	actionConfig := &action.Configuration{}
	err := actionConfig.Init(envSettings.RESTClientGetter(), envSettings.Namespace(), "", log.Debugf)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %v", err)
	}
	return actionConfig, nil
}

func LoadChart(path string) (*chart.Chart, error) {
	return loader.Load(path)
}

// GetChartVersion returns the chart version.
func GetChartVersion(chartPath string) (version string, err error) {
	chart, err := LoadChart(chartPath)
	if err != nil || chart == nil {
		return "", err
	}

	return chart.Metadata.Version, nil
}

func Template(releaseName, releaseNamespace, chartPath string, values map[string]interface{}) (*release.Release, error) {
	actionConfig, err := getActionConfig(releaseNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %v", err)
	}

	installAction := action.NewInstall(actionConfig)
	installAction.ReleaseName = releaseName
	installAction.Namespace = releaseNamespace
	installAction.DryRun = true
	installAction.Wait = false
	installAction.Timeout = WaitTimeout

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %s: %v", chartPath, err)
	}

	release, err := installAction.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to install release %s: %v", releaseName, err)
	}

	return release, nil
}

func Get(releaseName, releaseNamespace string) (*release.Release, error) {
	actionConfig, err := getActionConfig(releaseNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %v", err)
	}

	getAction := action.NewGet(actionConfig)
	return getAction.Run(releaseName)
}

func List(releaseNamespace string) ([]*release.Release, error) {
	actionConfig, err := getActionConfig(releaseNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %v", err)
	}

	listAction := action.NewList(actionConfig)
	return listAction.Run()
}

func Install(releaseName, releaseNamespace, chartPath string, values map[string]interface{}) (*release.Release, error) {
	actionConfig, err := getActionConfig(releaseNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %v", err)
	}

	installAction := action.NewInstall(actionConfig)
	installAction.ReleaseName = releaseName
	installAction.Namespace = releaseNamespace
	installAction.Wait = false
	installAction.Timeout = WaitTimeout

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %s: %v", chartPath, err)
	}

	release, err := installAction.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to install release %s: %v", releaseName, err)
	}

	return release, nil
}

func Upgrade(releaseName, releaseNamespace, chartPath string, values map[string]interface{}) (*release.Release, error) {
	actionConfig, err := getActionConfig(releaseNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %v", err)
	}

	upgradeAction := action.NewUpgrade(actionConfig)
	upgradeAction.Wait = false
	upgradeAction.Timeout = WaitTimeout

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart %s: %v", chartPath, err)
	}

	release, err := upgradeAction.Run(releaseName, chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade release %s: %v", releaseName, err)
	}

	return release, nil
}

func Uninstall(releaseName, releaseNamespace string) error {
	actionConfig, err := getActionConfig(releaseNamespace)
	if err != nil {
		return fmt.Errorf("failed to init helm action config: %v", err)
	}

	uninstallAction := action.NewUninstall(actionConfig)
	uninstallAction.Wait = false
	uninstallAction.Timeout = WaitTimeout

	_, err = uninstallAction.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall release %s: %v", releaseName, err)
	}

	return nil
}

// GenerateValueFile save the Helm values to a temporary file.
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

// GetChartName returns the name of the chart.
func GetChartName(chart string) string {
	return filepath.Base(chart)
}

// toYaml writes the Helm values to the given file.
func toYaml(values interface{}, file *os.File) error {
	log.Debugf("values: %+v", values)
	data, err := yaml.Marshal(values)
	if err != nil {
		log.Errorf("Failed to marshal value %v due to %v", values, err)
		return err
	}

	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		log.Errorf("Failed to write %v to %s due to %v", data, file.Name(), err)
	}
	return err
}
