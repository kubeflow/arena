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

	"github.com/kubeflow/arena/util"
	"github.com/kubeflow/arena/util/helm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
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
				fmt.Println(err)
				os.Exit(1)
			}
			releaseMap, err := helm.ListReleaseMap()
			// log.Printf("releaseMap %v", releaseMap)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

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
			jobs := []TrainingJob{}
			// for

			for name, ns := range releaseMap {
				supportedChart := false
				for _, trainer := range trainers {
					if trainer.IsSupported(name, ns) {
						job, err := trainer.GetTrainingJob(name, ns)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
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

			jobs = makeTrainingJobOrderdByAge(jobs)

			displayTrainingJobList(jobs, false)
		},
	}

	return command
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
			jobInfo.Age(),
			hostIP)
	}
	_ = w.Flush()
}

func PrintLine(w io.Writer, fields ...string)  {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}
