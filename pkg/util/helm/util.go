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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Process --set-file options to read file contents into the values map
	if len(options) > 0 {
		if err := processSetFileOptions(values, options); err != nil {
			return templateFileName, fmt.Errorf("failed to process set-file options: %w", err)
		}
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

// processSetFileOptions parses --set-file options from the given slice,
// reads the referenced file contents from disk, and merges them into the
// supplied values map. Only options whose token begins with "--set-file"
// are processed; any other entries (e.g. --set, --namespace) are ignored
// so callers may pass a mixed slice of Helm options safely.
//
// Each --set-file entry must be in one of the following forms:
//
//	--set-file key=path
//	--set-file=key=path
//
// where "key" is a dotted path into the values map (e.g.
// "configFiles.<hash>.config-N.content") and "path" points to a file on
// disk whose contents will be injected as a string value.
//
// If an intermediate key in the dotted path already exists in the values
// map but is not itself a map[string]interface{}, processSetFileOptions
// returns an error to avoid silently dropping user-supplied values.
// Malformed options and unreadable files also produce errors.
func processSetFileOptions(values map[string]interface{}, options []string) error {
	for _, opt := range options {
		if !strings.HasPrefix(opt, "--set-file") {
			continue
		}

		var assignment string
		switch {
		case strings.HasPrefix(opt, "--set-file="):
			assignment = strings.TrimPrefix(opt, "--set-file=")
		case opt == "--set-file":
			return fmt.Errorf("malformed --set-file option %q: expected key=path", opt)
		default:
			// "--set-file key=path" form: strip the flag token and any
			// whitespace separating it from the key=value assignment.
			assignment = strings.TrimSpace(strings.TrimPrefix(opt, "--set-file"))
		}

		idx := strings.Index(assignment, "=")
		if idx <= 0 || idx == len(assignment)-1 {
			return fmt.Errorf("malformed --set-file option %q: expected key=path", opt)
		}
		key := strings.TrimSpace(assignment[:idx])
		filePath := strings.TrimSpace(assignment[idx+1:])
		if key == "" || filePath == "" {
			return fmt.Errorf("malformed --set-file option %q: expected key=path", opt)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read --set-file %q: %w", filePath, err)
		}

		if err := setNestedValue(values, key, string(content)); err != nil {
			return fmt.Errorf("failed to set --set-file value for %q: %w", key, err)
		}
	}
	return nil
}

// setNestedValue sets a value in a map at a given dotted key path (e.g. "a.b.c").
// If an intermediate key already exists in the values map but is not a
// map[string]interface{}, an error is returned to prevent silent data loss.
func setNestedValue(values map[string]interface{}, key string, value interface{}) error {
	parts := strings.Split(key, ".")
	current := values

	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("empty key segment in %q", key)
		}

		isLast := i == len(parts)-1
		if isLast {
			current[part] = value
			return nil
		}

		next, ok := current[part]
		if !ok {
			child := map[string]interface{}{}
			current[part] = child
			current = child
			continue
		}

		child, ok := next.(map[string]interface{})
		if !ok {
			return fmt.Errorf("cannot set %q: intermediate key %q is not a map (got %T): %w", key, part, next, errIntermediateNotMap)
		}
		current = child
	}
	return nil
}

var errIntermediateNotMap = errors.New("intermediate value is not a map")
