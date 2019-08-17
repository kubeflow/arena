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

package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
)

type DiagnoseArgs struct {
	outputDir string
	jobType   string
}

type TrainingJobResources struct {
	statefulsetList *v1beta1.StatefulSetList
	jobList         *batchv1.JobList
	podList         *v1.PodList
	operatorPodList *v1.PodList
}

var (
	env = os.Environ()
)

const (
	DIR_PATH      = "/tmp/"
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

func NewDiagnoseCommand() *cobra.Command {
	diagnoseArgs := DiagnoseArgs{}

	var command = &cobra.Command{
		Use:   "diagnose job",
		Short: "diagnose relevant logs of a job",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			jobName := args[0]

			util.SetLogLevel(logLevel)
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = updateNamespace(cmd)
			if err != nil {
				log.Errorf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			job, err := searchTrainingJob(jobName, diagnoseArgs.jobType, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			path := DIR_PATH
			if len(diagnoseArgs.outputDir) != 0 {
				path = diagnoseArgs.outputDir
			}

			jobResources := job.GetTrainingJobResources(client, jobName)
			fmt.Printf("Collecting relevant logs of job %s in %v, please wait...\n", jobName, path)
			diagnoseJob(client, jobResources, diagnoseArgs.jobType, jobName, path)
			fmt.Println("Done.")
		},
	}
	command.Flags().StringVarP(&diagnoseArgs.outputDir, "outputDir", "o", "", "The output direction of the collected logs.")
	command.Flags().StringVarP(&diagnoseArgs.jobType, "type", "t", "", "The type of the selected job.")
	if err := command.MarkFlagRequired("type"); err != nil {
		fmt.Println(err)
		return nil
	}
	return command
}

func diagnoseJob(client *kubernetes.Clientset, jobResources TrainingJobResources, jobType, jobName, dirPath string) {
	timestamp, dirName := prepareFolder(jobType, jobName, dirPath)
	// 0. prepare environment
	kubePath, err := exec.LookPath(KUBECTL_BIN)
	if err != nil {
		log.Errorf("Failed to get the kubectl path due to %v", err)
		return
	}
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	// create description sub folder
	descDirName := filepath.Join(dirName, "description")
	_ = createDir(descDirName)

	// 1. describe job
	describeJob(kubePath, dirName, jobType, jobName)

	// 2. describe sts
	describeSts(jobResources, kubePath, descDirName)

	// 3. describe batch job
	describeBatchJob(jobResources, kubePath, descDirName)

	// 4. describe pod
	describePod(jobResources, kubePath, descDirName)

	// 5. describe the corresponding nodes
	describeNode(client, jobResources, kubePath, descDirName)

	// create logs sub folder
	logDirName := filepath.Join(dirName, "logs")
	_ = createDir(logDirName)

	// 6. pod logs
	logsPod(jobResources, kubePath, logDirName)

	// 7. job-operator logs
	logsOperatorPod(jobResources, kubePath, logDirName)

	// 8. get events
	getEvents(kubePath, dirName, timestamp)
}

// create the dest prepare folder
func prepareFolder(jobType, jobName, dirPath string) (string, string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	dirName := filepath.Join(dirPath, fmt.Sprintf("diagnose-%s-%s-%s", jobType, jobName, timestamp))
	fmt.Printf("The diagnose folder is %s.\n", dirName)

	if err := createDir(dirName); err != nil {
		log.Errorf("Failed to create the log folder due to %v", err)
		os.Exit(1)
	}
	return timestamp, dirName
}

func describeJob(kubePath, dirName, jobType, jobName string) {
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
	saveKubectlCmdResult(kubePath, describeArgs, dirName, jobName)
}

func describeSts(jobResources TrainingJobResources, kubePath, descDirName string) {
	statefulset := jobResources.statefulsetList
	if statefulset != nil {
		dir := filepath.Join(descDirName, "statefuleset")
		_ = createDir(dir)
		for _, s := range statefulset.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_STS, s.Name}
			saveKubectlCmdResult(kubePath, describeArgs, dir, s.Name)
		}
	}
}

func describeBatchJob(jobResources TrainingJobResources, kubePath, descDirName string) {
	batchJob := jobResources.jobList
	if batchJob != nil {
		dir := filepath.Join(descDirName, "batchjob")
		_ = createDir(dir)
		for _, j := range batchJob.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_JOB, j.Name}
			saveKubectlCmdResult(kubePath, describeArgs, dir, j.Name)
		}
	}
}

func describePod(jobResources TrainingJobResources, kubePath, descDirName string) {
	pod := jobResources.podList
	if pod != nil {
		dir := filepath.Join(descDirName, "pod")
		_ = createDir(dir)
		for _, p := range pod.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_POD, p.Name}
			saveKubectlCmdResult(kubePath, describeArgs, dir, p.Name)
		}
	}
}

func describeNode(client *kubernetes.Clientset, jobResources TrainingJobResources, kubePath, descDirName string) {
	nodesName := getNodesOfPod(client, jobResources.podList)
	if nodesName != nil {
		dir := filepath.Join(descDirName, "node")
		_ = createDir(dir)
		for _, nodeName := range nodesName {
			describeArgs := []string{KUBE_DESCRIBE, KUBE_NODE, nodeName}
			saveKubectlCmdResult(kubePath, describeArgs, dir, nodeName)
		}
	}
}

func logsPod(jobResources TrainingJobResources, kubePath, logDirName string) {
	pod := jobResources.podList
	if pod != nil {
		dir := filepath.Join(logDirName, "pod")
		_ = createDir(dir)
		for _, p := range pod.Items {
			logArgs := []string{KUBE_LOGS, "-n", namespace, p.Name}
			saveKubectlCmdResult(kubePath, logArgs, dir, p.Name)
		}
	}
}

func logsOperatorPod(jobResources TrainingJobResources, kubePath, logDirName string) {
	operatorPod := jobResources.operatorPodList
	if operatorPod != nil {
		dir := filepath.Join(logDirName, "operator-pod")
		_ = createDir(dir)
		for _, p := range operatorPod.Items {
			logArgs := []string{KUBE_LOGS, "-n", "arena-system", p.Name}
			saveKubectlCmdResult(kubePath, logArgs, dir, p.Name)
		}
	}
}

func getEvents(kubePath, dirName, timestamp string) {
	eventDirName := filepath.Join(dirName, "event")
	_ = createDir(eventDirName)
	eventArgs := []string{KUBE_GET, KUBE_EVENTS, "-o", "wide", "--all-namespaces"}
	saveKubectlCmdResult(kubePath, eventArgs, eventDirName, fmt.Sprintf("event-%s", timestamp))
}

// filter out the corresponding nodes
func getNodesOfPod(client *kubernetes.Clientset, podList *v1.PodList) []string {
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

func createDir(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, os.ModePerm)
		}
		return err
	}
	return nil
}

func saveKubectlCmdResult(kubePath string, args []string, dirName, fileName string) {
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
