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

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"

	"k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DiagnoseArgs struct {
	outputDir string
}

var (
	env = os.Environ()
)

const(
	DIR_PATH = "/tmp/"
	KUBECTL_BIN = "kubectl"
	KUBE_DESCRIBE = "describe"
	KUBE_GET = "get"
	KUBE_LOGS = "logs"
	KUBE_EVENTS = "events"
	KUBE_POD = "pod"
	KUBE_STS = "statefulsets"
	KUBE_JOB = "jobs"
	KUBE_NODE = "nodes"
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

			path := DIR_PATH
			if len(diagnoseArgs.outputDir) != 0 {
				path = diagnoseArgs.outputDir
			}
			fmt.Printf("Collecting relevant logs of job %s in %v, please wait...\n", jobName, path)
			diagnoseJob(client, jobName, path)
			fmt.Println("Done.")
		},
	}

	command.Flags().StringVarP(&diagnoseArgs.outputDir, "outputDir", "o", "", "The output direction of the collected logs.")
	return command
}

func diagnoseJob(client *kubernetes.Clientset, jobName, dirPath string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	dirName := filepath.Join(dirPath, fmt.Sprintf("diagnose-%s-%s", jobName, timestamp))
	fmt.Printf("The diagnose folder is %s.\n", dirName)
	// 0. create the dest folder
	if err := createDir(dirName); err != nil {
		log.Errorf("Failed to create the log folder due to %v", err)
		return
	}

	// 0. prepare environment
	kubePath, err := exec.LookPath(KUBECTL_BIN)
	if err != nil {
		log.Errorf("Failed to get the kubectl path due to %v", err)
		return
	}
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	descDirName := filepath.Join(dirName, "description")
	_ = createDir(descDirName)
	// 1. describe job
	// TODO: should judge the type of the dest job, only supprots mpi jobs now
	trainingType := knownTrainingTypes[1]
	describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, trainingType, jobName}
	saveKubectlCmdResult(kubePath, describeArgs, dirName, trainingType)

	// 2. describe sts
	statefulset, err := getStatefulSetsOfJob(client, jobName)
	if err != nil {
		log.Errorf("Failed to get statefulset of the job due to %v", err)
	} else if statefulset != nil {
		dir := filepath.Join(descDirName, "statefuleset")
		_ = createDir(dir)
		for _, s := range statefulset.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_STS, s.Name}
			saveKubectlCmdResult(kubePath, describeArgs, dir, s.Name)
		}
	}

	// 3. describe batch job
	batchJob, err := getBatchJobsOfJob(client, jobName)
	if err != nil {
		log.Errorf("Failed to get batchjob of the job due to %v", err)
	} else if batchJob != nil {
		dir := filepath.Join(descDirName, "batchjob")
		_ = createDir(dir)
		for _, j := range batchJob.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_JOB, j.Name}
			saveKubectlCmdResult(kubePath, describeArgs, dir, j.Name)
		}
	}

	// 4. describe pod
	pod, err := getPodsOfJob(client, jobName)
	if err != nil {
		log.Errorf("Failed to get pod of the job due to %v", err)
	} else if pod != nil {
		dir := filepath.Join(descDirName, "pod")
		_ = createDir(dir)
		for _, p := range pod.Items {
			describeArgs := []string{KUBE_DESCRIBE, "-n", namespace, KUBE_POD, p.Name}
			saveKubectlCmdResult(kubePath, describeArgs, dir, p.Name)
		}
	}

	// 5. describe the corresponding nodes
	nodesName := getNodesOfPod(client, pod)
	if nodesName != nil {
		dir := filepath.Join(descDirName, "node")
		_ = createDir(dir)
		for _, nodeName := range nodesName {
			describeArgs := []string{KUBE_DESCRIBE, KUBE_NODE, nodeName}
			saveKubectlCmdResult(kubePath, describeArgs, dir, nodeName)
		}
	}

	logDirName := filepath.Join(dirName, "logs")
	_ = createDir(logDirName)
	// 6. pod logs
	if pod != nil {
		dir := filepath.Join(logDirName, "pod")
		_ = createDir(dir)
		for _, p := range pod.Items {
			logArgs := []string{KUBE_LOGS, "-n", namespace, p.Name}
			saveKubectlCmdResult(kubePath, logArgs, dir, p.Name)
		}
	}

	// 7. job-operator logs
	operatorPod, err := getOperatorPodOfJob(client, jobName)
	if err != nil {
		log.Errorf("Failed to get operator pod of the job due to %v", err)
	} else if operatorPod != nil {
		dir := filepath.Join(logDirName, "operator-pod")
		_ = createDir(dir)
		for _, p := range operatorPod.Items {
			logArgs := []string{KUBE_LOGS, "-n", "arena-system", p.Name}
			saveKubectlCmdResult(kubePath, logArgs, dir, p.Name)
		}
	}

	// 8. get events
	eventDirName := filepath.Join(dirName, "event")
	_ = createDir(eventDirName)
	eventArgs := []string{KUBE_GET, KUBE_EVENTS, "-o", "wide", "--all-namespaces"}
	saveKubectlCmdResult(kubePath, eventArgs, eventDirName, fmt.Sprintf("event-%s", timestamp))
}

// filter out the dest batch jobs by the "release" label
func getPodsOfJob(client *kubernetes.Clientset, jobName string) (*v1.PodList, error) {
	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", jobName),
	})
	if err != nil {
		return nil, err
	}
	return pods, nil
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
			if !stringInSlice(nodeName, nodes) {
				nodes = append(nodes, p.Spec.NodeName)
			}
		}
	}
	return nodes
}

// FIXME: only supports looking for mpi job's stateful pod
func getStatefulSetsOfJob(client *kubernetes.Clientset, jobName string) (*v1beta1.StatefulSetList, error) {
	sts, err := client.AppsV1beta1().StatefulSets(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("mpi_job_name=%s", jobName),
	})
	if err != nil {
		return nil, err
	}
	return sts, nil
}

// FIXME: only supports looking for mpi job's batch job
func getBatchJobsOfJob(client *kubernetes.Clientset, jobName string) (*batchv1.JobList, error) {
	jobs, err := client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("mpi_job_name=%s", jobName),
	})
	if err != nil {
		log.Errorf("Failed to get batch jobs of the job due to %v", err)
		return nil, err
	}
	return jobs, nil
}

// FIXME: only supports looking for mpi job's operator pod
func getOperatorPodOfJob(client *kubernetes.Clientset, jobName string) (*v1.PodList, error) {
	pods, err := client.CoreV1().Pods("arena-system").List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: "app=mpi-operator",
	})
	if err != nil {
		return nil, err
	}
	return pods, nil
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

func stringInSlice(x string, list []string) bool {
	for _, y := range list {
		if y == x {
			return true
		}
	}
	return false
}
