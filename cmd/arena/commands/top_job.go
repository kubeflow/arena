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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"strconv"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/util"
	"k8s.io/api/core/v1"
)

func NewTopJobCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "job",
		Short: "Display Resource (GPU) usage of jobs.",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			var (
				jobs []TrainingJob
			)

			useCache = true
			allPods, err = acquireAllPods(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			allJobs, err = acquireAllJobs(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			trainers := NewTrainers(client)
			for _, trainer := range trainers {
				trainingJobs, err := trainer.ListTrainingJobs()
				if err != nil {
					log.Errorf("Failed due to %v", err)
					os.Exit(1)
				}

				for _, job := range trainingJobs {
					if job.GetStatus() != string(v1.PodSucceeded) {
						jobs = append(jobs, job)
					}
				}
			}

			jobs = makeTrainingJobOrderdByGPUCount(jobs)
			// TODO(cheyang): Support different job describer, such as MPI job/tf job describer
			topTrainingJob(jobs)
		},
	}

	command.Flags().BoolVar(&allNamespaces, "allNamespaces", false, "show all the namespaces")

	return command
}

func topTrainingJob(jobInfoList []TrainingJob) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalAllocatedGPUs int64
		totalRequestedGPUs int64
	)

	labelField := []string{"NAME", "GPU(Requests)", "GPU(Allocated)", "STATUS", "TRAINER", "AGE", "NODE"}

	PrintLine(w, labelField...)

	for _, jobInfo := range jobInfoList {

		hostIP := jobInfo.HostIPOfChief()
		requestedGPU := jobInfo.RequestedGPU()
		allocatedGPU := jobInfo.AllocatedGPU()
		// status, hostIP := jobInfo.getStatus()
		totalAllocatedGPUs += allocatedGPU
		totalRequestedGPUs += requestedGPU
		PrintLine(w, jobInfo.Name(),
			strconv.FormatInt(requestedGPU, 10),
			strconv.FormatInt(allocatedGPU, 10),
			jobInfo.GetStatus(),
			jobInfo.Trainer(),
			util.ShortHumanDuration(jobInfo.Age()),
			hostIP,
		)
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Total Allocated GPUs of Training Job:\n")
	fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalAllocatedGPUs, 10))
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Total Requested GPUs of Training Job:\n")
	fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalRequestedGPUs, 10))

	_ = w.Flush()
}

func fromByteToMiB(value float64) float64 {
	return value / 1048576
}
