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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
)

const (
	KUBECTL_BIN   = "kubectl"
	KUBE_DESCRIBE = "describe"
	KUBE_GET      = "get"
	KUBE_LOGS     = "logs"
	KUBE_EVENTS   = "events"
	KUBE_POD      = "pod"
	KUBE_STS      = "statefulsets"
	KUBE_JOB      = "jobs"
	KUBE_NODE     = "nodes"
)

var (
	kubectlCmd = []string{"kubectl"}
	env        = os.Environ()
)

/**
* dry-run creating kubernetes App Info for delete in future
* Exec /usr/local/bin/kubectl, [create --dry-run -f /tmp/values313606961 --namespace default]
**/

func SaveAppInfo(fileName, namespace string) (configFileName string, err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return "", err
	}

	args := []string{"create", "--dry-run", "--namespace", namespace, "-f", fileName}
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
func UninstallAppsWithAppInfoFile(appInfoFile, namespace string) (output string, err error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return "", err
	}

	if _, err = os.Stat(appInfoFile); err != nil {
		return "", err
	}

	args := []string{"cat", appInfoFile, "|", "xargs",
		binary, "delete", "--namespace", namespace}

	log.Debugf("Exec bash -c %v", args)

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}
	out, err := cmd.Output()
	log.Debugf("%s", string(out))

	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "bash -c", args, err)
	}

	return string(out), err
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
	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}
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
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	// return syscall.Exec(cmd, args, env)
	// 2. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	return cmd.CombinedOutput()
}

func DescribeJob(kubePath, dirName, namespace, jobType, jobName string) {
	var job string
	switch jobType {
	case "tfjob":
		job = "tfjobs.kubeflow.org"
	case "mpijob":
		job = "mpijobs.kubeflow.org"
	case "sparkjob":
		job = "sparkapplications.sparkoperator.k8s.io"
	case "volcanojob":
		job = "jobs.batch.volcano.sh"
	default:
		log.Fatalf("Unsupported job type: %s", jobType)
	}
	describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, job, jobName}
	SaveKubectlCmdResult(kubePath, describeArgs, dirName, jobName)
}

func DescribeSts(statefulsetList *v1beta1.StatefulSetList, kubePath, namespace, descDirName string) {
	if statefulsetList != nil {
		dir := filepath.Join(descDirName, "statefuleset")
		_ = CreateDir(dir)
		for _, s := range statefulsetList.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_STS, s.Name}
			SaveKubectlCmdResult(kubePath, describeArgs, dir, s.Name)
		}
	}
}

func DescribeBatchJob(jobList *batchv1.JobList, kubePath, namespace, descDirName string) {
	if jobList != nil {
		dir := filepath.Join(descDirName, "batchjob")
		_ = CreateDir(dir)
		for _, j := range jobList.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_JOB, j.Name}
			SaveKubectlCmdResult(kubePath, describeArgs, dir, j.Name)
		}
	}
}

func DescribePod(podList *v1.PodList, kubePath, namespace, descDirName string) {
	if podList != nil {
		dir := filepath.Join(descDirName, "pod")
		_ = CreateDir(dir)
		for _, p := range podList.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_POD, p.Name}
			SaveKubectlCmdResult(kubePath, describeArgs, dir, p.Name)
		}
	}
}

func DescribeNode(client *kubernetes.Clientset, podList *v1.PodList, kubePath, namespace, descDirName string) {
	nodesName := GetNodesOfPod(client, podList, namespace)
	if nodesName != nil {
		dir := filepath.Join(descDirName, "node")
		_ = CreateDir(dir)
		for _, nodeName := range nodesName {
			describeArgs := []string{KUBE_DESCRIBE, KUBE_NODE, nodeName}
			SaveKubectlCmdResult(kubePath, describeArgs, dir, nodeName)
		}
	}
}

func LogsPod(podList *v1.PodList, kubePath, namespace, logDirName string) {
	if podList != nil {
		dir := filepath.Join(logDirName, "pod")
		_ = CreateDir(dir)
		for _, p := range podList.Items {
			logArgs := []string{KUBE_LOGS, "-n", namespace, p.Name}
			SaveKubectlCmdResult(kubePath, logArgs, dir, p.Name)
		}
	}
}

func LogsOperatorPod(operatorPodList *v1.PodList, kubePath, namespace, logDirName string) {
	if operatorPodList != nil {
		dir := filepath.Join(logDirName, "operator-pod")
		_ = CreateDir(dir)
		for _, p := range operatorPodList.Items {
			logArgs := []string{KUBE_LOGS, "-n", "arena-system", p.Name}
			SaveKubectlCmdResult(kubePath, logArgs, dir, p.Name)
		}
	}
}

func GetEvents(kubePath, dirName, timestamp string) {
	eventDirName := filepath.Join(dirName, "event")
	_ = CreateDir(eventDirName)
	eventArgs := []string{KUBE_GET, KUBE_EVENTS, "-o", "wide", "--all-namespaces"}
	SaveKubectlCmdResult(kubePath, eventArgs, eventDirName, fmt.Sprintf("event-%s", timestamp))
}

// filter out the corresponding nodes
func GetNodesOfPod(client *kubernetes.Clientset, podList *v1.PodList, namespace string) []string {
	nodes := []string{}
	for _, pod := range podList.Items {
		p, err := client.CoreV1().Pods(namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Failed to get pod info of the cluster due to %v", err)
		} else {
			nodeName := p.Spec.NodeName
			if !util.StringInSlice(nodeName, nodes) {
				nodes = append(nodes, p.Spec.NodeName)
			}
		}
	}
	return nodes
}

func CreateDir(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, os.ModePerm)
		}
		return err
	}
	return nil
}

func SaveKubectlCmdResult(kubePath string, args []string, dirName, fileName string) {
	cmd := exec.Command(kubePath, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to collect info due to %v", err)
	} else {
		f, err := os.OpenFile(filepath.Join(dirName, fileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Failed to open file due to %v", err)
		}
		defer f.Close()

		_, _ = f.WriteString(fmt.Sprintf("$ %s %s\n", kubePath, strings.Join(args, " ")))
		_, err = f.Write(out)
		if err != nil {
			log.Errorf("Failed to write file due to %v", err)
		}
	}
}

// prepare environment
func PrepareEnv() string {
	kubePath, err := exec.LookPath(KUBECTL_BIN)
	if err != nil {
		log.Errorf("Failed to get the kubectl path due to %v", err)
		return ""
	}
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}
	return kubePath
}
