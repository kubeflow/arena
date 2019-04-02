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
	"os"

	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/spf13/cobra"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"fmt"
	"github.com/kubeflow/arena/pkg/util"
)

// NewDeleteCommand
func NewServingDeleteCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "delete a serving job",
		Short: "delete a serving job and its associated pods",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			setupKubeconfig()
			err := updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			for _, jobName := range args {
				deleteServingJob(jobName)
			}
		},
	}

	return command
}

func deleteServingJob(servingJob string) error {
	var servingTypes []string
	err := helm.DeleteRelease(servingJob)
	if err == nil {
		log.Infof("Delete the job %s successfully.", servingJob)
		return nil
	}

	log.Debugf("%s wasn't deleted by helm due to %v", servingJob, err)

	// 2. Handle serving jobs created by arena
	servingTypes = getServingTypes(servingJob, namespace)
	if len(servingTypes) == 0 {
		return fmt.Errorf("There is no serving job found with the name %s, please check it with `arena serve list | grep %s`",
			servingJob,
			servingJob)
	}

	err = workflow.DeleteJob(servingJob, namespace, servingTypes[0])
	if err != nil {
		return err
	}
	log.Infof("The Serving job %s has been deleted successfully", servingJob)
	return nil
}

func deleteServingJobByHelm(servingJob string) error {
	return helm.DeleteRelease(servingJob)
}
