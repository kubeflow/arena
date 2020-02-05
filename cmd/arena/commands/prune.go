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

	"time"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type PruneArgs struct {
	since time.Duration
}

func NewPruneCommand() *cobra.Command {
	var (
		pruneArgs PruneArgs
	)
	var command = &cobra.Command{
		Use:   "prune history job",
		Short: "prune history job",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)

			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
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
			for _, trainer := range trainers {
				trainingJobs, err := trainer.ListTrainingJobs(namespace)
				if err != nil {
					log.Errorf("Failed due to %v", err)
					os.Exit(1)
				}
				jobs = append(jobs, trainingJobs...)
			}

			deleted := false
			if pruneArgs.since == -1 {
				fmt.Println("Your need to specify the relative duration live time of the job that need to be cleaned by --since. Like --since 10h")
				return
			}
			for _, job := range jobs {
				if GetJobRealStatus(job) != "RUNNING" {
					if job.Age() > pruneArgs.since {
						deleted = true
						fmt.Printf("Delete %s %s with Age %s \n", job.Trainer(), job.Name(), util.ShortHumanDuration(job.Age()))
						err = deleteTrainingJob(job.Name(), "")
						if err != nil {
							fmt.Printf("Failed to delete %s %s, err: %++v", job.Trainer(), job.Name(), err)
						}
					}
				}
			}
			if !deleted {
				fmt.Println("No job need to be deleted")
			}
		},
	}

	command.Flags().DurationVarP(&pruneArgs.since, "since", "s", -1, "Clean job that live longer than relative duration like 5s, 2m, or 3h.")
	return command
}
