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
	"strings"
	"text/tabwriter"

	"io"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "list all the training jobs",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)

			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				log.Errorf("Failed due to %v", err)
				os.Exit(1)
			}

			err = updateNamespace(cmd)
			if err != nil {
				log.Errorf("Failed due to %v", err)
				os.Exit(1)
			}

			// determine use cache
			useCache = true
			allPods, err = acquireAllPods(client)
			if err != nil {
				log.Errorf("Failed due to %v", err)
				os.Exit(1)
			}

			allJobs, err = acquireAllJobs(client)
			if err != nil {
				log.Errorf("Failed due to %v", err)
				os.Exit(1)
			}
			jobs := []TrainingJob{}
			trainers := NewTrainers(client)
			for _, trainer := range trainers {
				trainingJobs, err := trainer.ListTrainingJobs()
				if err != nil {
					log.Errorf("Failed due to %v", err)
					os.Exit(1)
				}
				jobs = append(jobs, trainingJobs...)
			}

			jobs = makeTrainingJobOrderdByAge(jobs)

			displayTrainingJobList(jobs, false)
		},
	}

	command.Flags().BoolVar(&allNamespaces, "allNamespaces", false, "show all the namespaces")

	return command
}
func ListTrainingJobs(kubeconfig, logLevel, ns string) ([]TrainingJob, error) {
	if err := InitCommonConfig(kubeconfig, logLevel, ns); err != nil {
		return nil, err
	}
	useCache = false
	var err error
	allPods, err = acquireAllPods(clientset)
	if err != nil {
		return nil, err
	}
	allJobs, err = acquireAllJobs(clientset)
	if err != nil {
		return nil, err
	}
	jobs := []TrainingJob{}
	trainers := NewTrainers(clientset)
	for _, trainer := range trainers {
		trainingJobs, err := trainer.ListTrainingJobs()
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, trainingJobs...)
	}

	jobs = makeTrainingJobOrderdByAge(jobs)
	return jobs, nil
}

/**
* original job list, deprecated
 */
func trainingJobList(client *kubernetes.Clientset) ([]TrainingJob, error) {
	useHelm := true
	releaseMap, err := helm.ListReleaseMap()
	// log.Printf("releaseMap %v", releaseMap)
	if err != nil {
		log.Debugf("Failed to helm list due to %v", err)
		useHelm = false
	}

	trainers := NewTrainers(client)
	jobs := []TrainingJob{}

	// 1. search by using helm
	if useHelm {
		for name, ns := range releaseMap {
			supportedChart := false
			for _, trainer := range trainers {
				if trainer.IsSupported(name, ns) {
					job, err := trainer.GetTrainingJob(name, ns)
					if err != nil {
						log.Errorf("Failed due to %v", err)
						return jobs, err
					}
					jobs = append(jobs, job)
					supportedChart = true
					break
				}
			}

			if !supportedChart {
				log.Debugf("Unknown chart %s\n", name)
			}

		}
	}

	// 2. search by using configmap
	cms := []types.TrainingJobInfo{}
	if allNamespaces {
		cms, err = kubectl.ListAppConfigMaps(client, metav1.NamespaceAll, knownTrainingTypes)
	} else {
		cms, err = kubectl.ListAppConfigMaps(client, namespace, knownTrainingTypes)
	}

	if err != nil {
		log.Errorf("Failed due to %v", err)
		return jobs, err
	}

	log.Debugf("job config maps: %v", cms)

	for _, cm := range cms {
		job, err := searchTrainingJob(cm.Name, cm.Type, cm.Namespace)
		if err != nil {
			log.Errorf("Failed due to %v", err)
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func displayTrainingJobList(jobInfoList []TrainingJob, displayGPU bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	labelField := []string{"NAME", "STATUS", "TRAINER", "AGE", "NODE"}

	PrintLine(w, labelField...)

	for _, jobInfo := range jobInfoList {
		status := GetJobRealStatus(jobInfo)
		hostIP := jobInfo.HostIPOfChief()
		PrintLine(w, jobInfo.Name(),
			status,
			strings.ToUpper(jobInfo.Trainer()),
			util.ShortHumanDuration(jobInfo.Age()),
			hostIP)
	}
	_ = w.Flush()
}

func PrintLine(w io.Writer, fields ...string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}
