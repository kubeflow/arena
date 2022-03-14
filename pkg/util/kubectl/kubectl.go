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

package kubectl

import (
	"context"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"io/ioutil"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/exec"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var kubectlCmd = []string{"arena-kubectl"}
var kubeClient *kubernetes.Clientset

/**
* dry-run creating kubernetes App Info for delete in future
* Exec /usr/local/bin/kubectl, [create --dry-run -f /tmp/values313606961 --namespace default]
**/

func SaveAppInfo(fileName, namespace string) (configFileName string, err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return "", err
	}

	args := []string{"create", "--dry-run=client", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)
	output := string(out)
	result := []string{}

	// fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Errorf("The output is %s\n", output)
		return "", err
	}

	// 1. generate the config file
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", configFile.Name(), err)
		return "", err
	}

	configFileName = configFile.Name()
	log.Debugf("Save the config file %s", configFileName)

	// 2. save app types to config file
	lines := strings.Split(output, "\n")
	log.Debugf("dry run result: %v", lines)

	for _, line := range lines {
		line := strings.TrimSpace(line)
		cols := strings.Fields(line)
		log.Debugf("cols: %s, %d", cols, len(cols))
		if len(cols) == 0 {
			continue
		}
		result = append(result, cols[0])
	}

	data := []byte(strings.Join(result, "\n"))
	defer configFile.Close()
	_, err = configFile.Write(data)
	if err != nil {
		log.Errorf("Failed to write %v to %s due to %v", data, configFileName, err)
		return configFileName, err
	}

	return configFileName, nil
}

/**
* Delete kubernetes config to uninstall app
* Exec /usr/local/bin/kubectl, [delete -f /tmp/values313606961 --namespace default]
**/
func UninstallApps(fileName, namespace string) (err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"delete", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)

	fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
* Delete kubernetes config to uninstall app
* Exec /usr/local/bin/kubectl, [delete -f /tmp/values313606961 --namespace default]
**/
func UninstallAppsWithAppInfoFile(appInfoFile, namespace string) error {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(appInfoFile)
	if err != nil {
		return err
	}
	resources := strings.Split(string(data), "\n")
	// Error from server (NotFound): tfjobs.kubeflow.org "tf-standalone-test-3" not found

	var wg sync.WaitGroup
	locker := new(sync.RWMutex)
	errs := []string{}
	for _, r := range resources {
		wg.Add(1)
		resource := r
		go func() {
			defer wg.Done()
			args := []string{binary, "delete", resource, "--namespace", namespace}
			log.Debugf("Exec bash -c %v", args)
			cmd := exec.Command("bash", "-c", strings.Join(args, " "))
			out, err := cmd.CombinedOutput()
			if err != nil {
				if !strings.Contains(string(out), "Error from server (NotFound): ") {
					locker.Lock()
					errs = append(errs, string(out))
					locker.Unlock()
				}
				return
			}
			log.Debugf("%v", out)
		}()
	}
	wg.Wait()
	if len(errs) != 0 {
		log.Debugf("Failed to uninstall app with app file,reason: %v", errs)
		return fmt.Errorf("%v", strings.Join(errs, "\n"))
	}
	return nil
}

/**
* Apply kubernetes config to install app
* Exec /usr/local/bin/kubectl, [apply -f /tmp/values313606961 --namespace default]
**/
func InstallApps(fileName, namespace string) (output string, err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return output, err
	}

	args := []string{"apply", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)

	log.Debugf("%s", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return string(out), err
}

/**
* This name should be <job-type>-<job-name>
* create configMap by using name, namespace and configFile
**/
func CreateAppConfigmap(name, trainingType, namespace, configFileName, appInfoFileName, chartName, chartVersion string) (err error) {
	if _, err = os.Stat(configFileName); os.IsNotExist(err) {
		return err
	}

	if _, err = os.Stat(appInfoFileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"create", "configmap", fmt.Sprintf("%s-%s", name, trainingType),
		"--namespace", namespace,
		fmt.Sprintf("--from-file=%s=%s", "values", configFileName),
		fmt.Sprintf("--from-file=%s=%s", "app", appInfoFileName),
		fmt.Sprintf("--from-literal=%s=%s", chartName, chartVersion)}
	// "--overrides='{\"metadata\":{\"label\":\"createdBy\": \"arena\"}}'"}
	out, err := kubectl(args)

	fmt.Printf("%s", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

func LabelAppConfigmap(name, trainingType, namespace, label string) (err error) {
	args := []string{"label", "configmap", fmt.Sprintf("%s-%s", name, trainingType),
		"--namespace", namespace,
		label}
	// "--overrides='{\"metadata\":{\"label\":\"createdBy\": \"arena\"}}'"}
	out, err := kubectl(args)

	fmt.Printf("%s", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
*
* delete configMap by using name, namespace
**/
func DeleteAppConfigMap(name, namespace string) (err error) {
	args := []string{"delete", "configmap", name, "--namespace", namespace}
	out, err := kubectl(args)

	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Debugf("%s", string(out))
	} else {
		fmt.Printf("%s", string(out))
	}

	return err
}

/**
*
* get configMap by using name, namespace
**/
func CheckAppConfigMap(name, namespace string) (found bool) {
	args := []string{"get", "configmap", name, "--namespace", namespace}
	out, err := kubectl(args)

	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(`Error from server (NotFound): configmaps "%v" not found`, name)) {
			found = false
		}
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Debugf("%s", string(out))
	} else {
		log.Debugf("%s", string(out))
		found = true
	}

	return found
}

/**
*
* save the key of configMap into a file
**/
func SaveAppConfigMapToFile(name, key, namespace string) (fileName string, err error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return "", err
	}

	file, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", file.Name(), err)
		return fileName, err
	}
	fileName = file.Name()

	args := []string{binary, "get", "configmap", name,
		"--namespace", namespace,
		fmt.Sprintf("-o=jsonpath='{.data.%s}'", key),
		">", fileName}
	log.Debugf("Exec bash -c %s", strings.Join(args, " "))

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	out, err := cmd.Output()
	fmt.Printf("%s", string(out))

	if err != nil {
		return fileName, fmt.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}
	return fileName, err
}

func kubectl(args []string) ([]byte, error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return nil, err
	}

	// 1. prepare the arguments
	// args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	log.Debugf("Exec %s, %v", binary, args)

	env := os.Environ()

	// return syscall.Exec(cmd, args, env)
	// 2. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	return cmd.CombinedOutput()
}

func GetCrdNames() ([]string, error) {
	crdNames := []string{}
	args := []string{"get", "crds"}
	out, err := kubectl(args)
	if err != nil {
		if strings.Contains(err.Error(), "No resources found.") {
			return crdNames, nil
		}
		return nil, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		item := strings.Trim(line, " ")
		if item == "" {
			continue
		}
		if strings.Contains(item, "No resources found.") {
			continue
		}
		if strings.Contains(item, "CREATED AT") {
			continue
		}
		crdNames = append(crdNames, strings.Trim(strings.Split(item, " ")[0], " "))
	}
	return crdNames, nil
}

func GetDeployment(name, namespace string) (*v1.Deployment, error) {
	arenaConfiger := config.GetArenaConfiger()
	client := arenaConfiger.GetClientSet()

	return client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func UpdateDeployment(deploy *v1.Deployment) error {
	arenaConfiger := config.GetArenaConfiger()
	client := arenaConfiger.GetClientSet()

	_, err := client.AppsV1().Deployments(deploy.Namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
	return err
}
