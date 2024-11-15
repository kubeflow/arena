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
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var helmCmd = []string{"arena-helm"}

/**
* install the release with cmd: helm install -f values.yaml chart_name
 */
func InstallRelease(name string, namespace string, values interface{}, chartName string) error {
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return err
	}

	// 1. generate the template file
	valueFile, err := os.CreateTemp(os.TempDir(), "values")
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", valueFile.Name(), err)
		return err
	} else {
		log.Debugf("Save the values file %s", valueFile.Name())
	}
	// defer os.Remove(valueFile.Name())

	// 2. dump the object into the template file
	err = toYaml(values, valueFile)
	if err != nil {
		return err
	}

	// 3. check if the chart file exists, if it's unix path, then check if it's exist
	if strings.HasPrefix(chartName, "/") {
		if _, err = os.Stat(chartName); os.IsNotExist(err) {
			// TODO: the chart will be put inside the binary in future
			return err
		}
	}

	// 4. prepare the arguments
	args := []string{"install", "-f", valueFile.Name(), "--namespace", namespace, name, chartName}
	log.Debugf("Exec %s, %v", binary, args)

	env := os.Environ()

	// return syscall.Exec(cmd, args, env)
	// 5. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	fmt.Println("")
	fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Fatalf("Failed to execute %s, %v with %v", binary, args, err)
	}

	// 6. clean up the value file if needed
	if log.GetLevel() != log.DebugLevel {
		err = os.Remove(valueFile.Name())
		if err != nil {
			log.Warnf("Failed to delete %s due to %v", valueFile.Name(), err)
		}
	}

	return nil
}

/**
* check if the release exist
 */
func CheckRelease(name string) (exist bool, err error) {
	_, err = exec.LookPath(helmCmd[0])
	if err != nil {
		return exist, err
	}

	cmd := exec.Command(helmCmd[0], "get", name)
	// support multiple cluster management
	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v", err)
		return exist, err
	}

	err = cmd.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus := status.ExitStatus()
				log.Debugf("Exit Status: %d", exitStatus)
				if exitStatus == 1 {
					err = nil
				}
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
			return exist, err
		}
	} else {
		waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if waitStatus.ExitStatus() == 0 {
			exist = true
		} else {
			if waitStatus.ExitStatus() != -1 {
				return exist, fmt.Errorf("unexpected return code %d when exec helm get %s", waitStatus.ExitStatus(), name)
			}
		}
	}

	return exist, err
}

func DeleteRelease(name string) error {
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return err
	}

	args := []string{"del", "--purge", name}
	cmd := exec.Command(binary, args...)

	// return syscall.Exec(cmd, args, env)
	out, err := cmd.Output()
	log.Debugf("Delete release's result: %s\n", string(out))
	return err
}

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
