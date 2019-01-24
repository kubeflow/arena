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

	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	trainingType string
)

// NewDeleteCommand
func NewDeleteCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "delete a training job",
		Short: "delete a training job and its associated pods",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			err := updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			setupKubeconfig()
			_, err = initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, jobName := range args {
				err = deleteTrainingJob(jobName)
				if err != nil {
					fmt.Printf("Failed to delete %s due to %v", jobName, err)
				}
			}
		},
	}
	command.Flags().StringVar(&trainingType, "--type", "", "The training type to delete, the possible option is tfjob, mpijob, horovodjob or standalonejob. (optional)")

	return command
}

func deleteTrainingJob(jobName string) error {
	// 1. Handle legacy training job
	err := helm.DeleteRelease(jobName)
	if err == nil {
		return nil
	}

	log.Debugf("it didn't deleted by helm due to %v", err)

	// 2. Handle training jobs created by arena
	var job TrainingJob

	if len(trainingType) > 0 {
		if isKnownTrainingType(trainingType) {
			job, err = getTrainingJobByType(clientset, name, namespace, trainingType)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%s is unknown training type, please choose a known type from %v",
				trainingType,
				knownTrainingTypes)
		}
	} else {
		jobs, err := getTrainingJobs(clientset, jobName, namespace)
		if err != nil {
			return err
		}

		if len(jobs) > 1 {
			return fmt.Errorf("There are more than 1 training job with the same name %s, please check it with `arena list | grep %s`",
				jobName,
				jobName)
		} else {
			job = jobs[0]
		}
	}

	err = workflow.DeleteJob(jobName, namespace, job.Trainer())
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been deleted successfully", jobName)
	// (TODO: cheyang)3. Handle training jobs created by others, to implement
	return nil
}

func deleteTrainingJobWithHelm(jobName string) error {
	return helm.DeleteRelease(jobName)
}

func isKnownTrainingType(trainingType string) bool {
	for _, knownType := range knownTrainingTypes {
		if trainingType == knownType {
			return true
		}
	}

	return false
}
