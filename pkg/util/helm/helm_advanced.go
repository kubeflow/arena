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

	log "github.com/sirupsen/logrus"
)

/**
* generate helm template without tiller: helm template -f values.yaml chart_name
* Exec /usr/local/bin/helm, [template -f /tmp/values313606961 --namespace default --name hj /charts/tf-horovod]
* returns generated template file: templateFileName
 */
func GenerateHelmTemplateLegacy(name string, namespace string, valueFileName string, chartName string, options ...string) (templateFileName string, err error) {
	tempName := fmt.Sprintf("%s.yaml", name)
	templateFile, err := os.CreateTemp("", tempName)
	if err != nil {
		return templateFileName, err
	}
	templateFileName = templateFile.Name()

	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return templateFileName, err
	}

	// 3. check if the chart file exists
	// if strings.HasPrefix(chartName, "/") {
	if _, err = os.Stat(chartName); err != nil {
		return templateFileName, err
	}
	// }

	// 4. prepare the arguments
	args := []string{binary, "template", "-f", valueFileName,
		"--namespace", namespace,
		name,
	}
	if len(options) != 0 {
		args = append(args, options...)
	}

	args = append(args, []string{chartName, ">", templateFileName}...)

	log.Debugf("Exec bash -c %v", args)

	// return syscall.Exec(cmd, args, env)
	// 5. execute the command
	log.Debugf("Generating template  %v", args)
	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	// cmd.Env = env
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s", string(out))
	if err != nil {
		return templateFileName, fmt.Errorf("failed to execute %s, %v with %v", binary, args, err)
	}

	// // 6. clean up the value file if needed
	// if log.GetLevel() != log.DebugLevel {
	// 	err = os.Remove(valueFileName)
	// 	if err != nil {
	// 		log.Warnf("Failed to delete %s due to %v", valueFileName, err)
	// 	}
	// }

	return templateFileName, nil
}
