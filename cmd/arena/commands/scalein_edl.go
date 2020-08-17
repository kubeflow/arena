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
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"time"
)

var (
	scalein_edl_chart = util.GetChartsFolder() + "/scalein"
	scaleinEnvs       []string
	scaleinDuration   time.Duration
)

const (
	scaleInScript = "/usr/local/bin/scaler.sh --delete"
)

func NewScaleInEDLJobCommand() *cobra.Command {
	var (
		submitArgs ScaleInEDLJobArgs
	)

	submitArgs.Mode = "scalein"

	var command = &cobra.Command{
		Use:     "edljob",
		Short:   "scalein a edljob",
		Aliases: []string{"edl"},
		Run: func(cmd *cobra.Command, args []string) {
			//fmt.Println("args:", args)
			//if len(args) == 0 {
			//	cmd.HelpFunc()(cmd, args)
			//	os.Exit(1)
			//}

			util.SetLogLevel(logLevel)
			setupKubeconfig()
			_, err := initKubeClient()
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

			err = submitScaleInEDLJob(args, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		},
	}

	command.Flags().StringVar(&submitArgs.Name, "name", "", "required, edl job name")
	command.MarkFlagRequired("name")
	command.Flags().DurationVarP(&scaleinDuration, "timeout", "t", 60*time.Second, "timeout of callback scaler script, like 5s, 2m, or 3h.")
	command.Flags().IntVar(&submitArgs.Retry, "retry", 0, "retry times.")
	command.Flags().IntVar(&submitArgs.Count, "count", 1, "the nums of you want to add or delete worker.")
	command.Flags().StringVar(&submitArgs.Script, "script", scaleInScript, "script of scaling.")
	command.Flags().StringArrayVarP(&scaleinEnvs, "env", "e", []string{}, "the environment variables.")
	return command
}

type ScaleInEDLJobArgs struct {
	ScaleEDLJobArgs `yaml:",inline"`
}

func ParseSinceSeconds(since string) (*int64, error) {
	invalidReturn := int64(0)
	parsedSince, err := strconv.ParseInt(since, 10, 64)
	if err != nil {
		return &invalidReturn, err
	}
	return &parsedSince, nil
}

func (submitArgs *ScaleInEDLJobArgs) prepare() (err error) {
	log.Debugf("scaleinEnvs: %v", scaleinEnvs)
	if len(scaleinEnvs) > 0 {
		submitArgs.Envs = transformSliceToMap(scaleinEnvs, "=")
	}
	submitArgs.Timeout = int(scaleinDuration.Seconds())

	edljobName := submitArgs.Name
	trainer := NewEDLJobTrainer(clientset)
	job, err := trainer.GetTrainingJob(edljobName, namespace)
	if err != nil {
		return fmt.Errorf("Check %s exist due to error %v", edljobName, err)
	}

	if job == nil {
		return fmt.Errorf("the job %s is not found, please check it firstly.", edljobName)
	}
	if "RUNNING" == job.GetStatus() || "SCALING" == job.GetStatus() {
		currentWorkers := getEDLJobCurrentReplicas(job)
		minWorkers := getEDLJobMinReplicas(job)
		log.Debugf("currentWorkers: %v, minWorkers: %v", currentWorkers, minWorkers)
		if currentWorkers-submitArgs.Count < minWorkers {
			return fmt.Errorf("the number of current workers minus the number of scaling in is less than the min-workers. please try again later.")
		}
		return nil
	} else {
		return fmt.Errorf("the job: %s status: %s , is not RUNNING or SCALING, please try again later.", edljobName, job.GetStatus())
	}
}

func submitScaleInEDLJob(args []string, submitArgs *ScaleInEDLJobArgs) (err error) {
	err = submitArgs.prepare()
	if err != nil {
		return err
	}
	scaleName := fmt.Sprintf("%s-%d", submitArgs.Name, time.Now().Unix())
	log.Debugf("submitArgs: %v", submitArgs)
	err = workflow.SubmitJob(scaleName, submitArgs.Mode, namespace, submitArgs, scalein_edl_chart)
	if err != nil {
		return err
	}

	log.Infof("The scalein job %s has been submitted successfully", scaleName)
	return nil
}
