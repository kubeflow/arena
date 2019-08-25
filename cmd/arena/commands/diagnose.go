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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/kubectl"
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

const DIR_PATH = "/tmp/"

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
	command.Flags().StringVarP(&diagnoseArgs.outputDir, "output-dir", "o", "", "The output direction of the collected logs.")
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
	kubePath := kubectl.PrepareEnv()

	// create description sub folder
	descDirName := filepath.Join(dirName, "description")
	_ = kubectl.CreateDir(descDirName)

	// 1. describe job
	kubectl.DescribeJob(kubePath, dirName, namespace, jobType, jobName)

	// 2. describe sts
	kubectl.DescribeSts(jobResources.statefulsetList, kubePath, namespace, descDirName)

	// 3. describe batch job
	kubectl.DescribeBatchJob(jobResources.jobList, kubePath, namespace, descDirName)

	// 4. describe pod
	kubectl.DescribePod(jobResources.podList, kubePath, namespace, descDirName)

	// 5. describe the corresponding nodes
	kubectl.DescribeNode(client, jobResources.podList, kubePath, namespace, descDirName)

	// create logs sub folder
	logDirName := filepath.Join(dirName, "logs")
	_ = kubectl.CreateDir(logDirName)

	// 6. pod logs
	kubectl.LogsPod(jobResources.podList, kubePath, namespace, logDirName)

	// 7. job-operator logs
	kubectl.LogsOperatorPod(jobResources.operatorPodList, kubePath, namespace, logDirName)

	// 8. get events
	kubectl.GetEvents(kubePath, dirName, timestamp)
}

// create the dest prepare folder
func prepareFolder(jobType, jobName, dirPath string) (string, string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	dirName := filepath.Join(dirPath, fmt.Sprintf("diagnose-%s-%s-%s", jobType, jobName, timestamp))
	fmt.Printf("The diagnose folder is %s.\n", dirName)

	if err := kubectl.CreateDir(dirName); err != nil {
		log.Errorf("Failed to create the log folder due to %v", err)
		os.Exit(1)
	}
	return timestamp, dirName
}
