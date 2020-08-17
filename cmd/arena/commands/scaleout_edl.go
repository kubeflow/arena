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
	"github.com/kubeflow/arena/pkg/operators/edl-operator/api/v1alpha1"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	scaleout_edl_chart = util.GetChartsFolder() + "/scaleout"
	scaleoutEnvs       []string
	scaleoutDuration   time.Duration
)

const (
	scaleOutScript = "/usr/local/bin/scaler.sh --add"
)

func NewScaleOutEDLJobCommand() *cobra.Command {
	var (
		submitArgs ScaleOutEDLJobArgs
	)

	submitArgs.Mode = "scaleout"

	var command = &cobra.Command{
		Use:     "edljob",
		Short:   "scaleout a edljob",
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

			err = submitScaleOutEDLJob(args, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		},
	}

	command.Flags().StringVar(&submitArgs.Name, "name", "", "required, edl job name")
	command.MarkFlagRequired("name")
	command.Flags().DurationVarP(&scaleoutDuration, "timeout", "t", 60*time.Second, "timeout of callback scaler script.")
	command.Flags().IntVar(&submitArgs.Retry, "retry", 0, "retry times.")
	command.Flags().IntVar(&submitArgs.Count, "count", 1, "the nums of you want to add or delete worker.")
	command.Flags().StringVar(&submitArgs.Script, "script", scaleOutScript, "script of scaling.")
	command.Flags().StringArrayVarP(&scaleoutEnvs, "env", "e", []string{}, "the environment variables.")
	return command
}

type ScaleEDLJobArgs struct {
	Mode string `yaml:"mode"` // --mode
	//--name string     required, edl job name
	Name string `yaml:"edlName"`
	//--timeout int     timeout of callback scaler script.
	Timeout int `yaml:"timeout"`
	//--retry int       retry times.
	Retry int `yaml:"retry"`
	//--count int       the nums of you want to add or delete worker.
	Count int `yaml:"count"`
	//--script string        script of scaling.
	Script string `yaml:"script"`
	//-e, --env stringArray      the environment variables
	Envs map[string]string `yaml:"envs"`
}

func getEDLJobMaxReplicas(job TrainingJob) (maxReplicas int) {
	edlJob := job.GetTrainJob().(v1alpha1.TrainingJob)
	_, worker := parseAnnotations(edlJob)
	maxReplicas = MAXWORKERS
	if worker != nil {
		if _, ok := worker["maxReplicas"]; ok {
			maxReplicas = int(worker["maxReplicas"].(float64))
		}
	}
	return maxReplicas
}

func getEDLJobMinReplicas(job TrainingJob) (minReplicas int) {
	edlJob := job.GetTrainJob().(v1alpha1.TrainingJob)
	_, worker := parseAnnotations(edlJob)
	minReplicas = MINWORKERS
	if worker != nil {
		if _, ok := worker["minReplicas"]; ok {
			minReplicas = int(worker["minReplicas"].(float64))
		}
	}
	return minReplicas
}

func getEDLJobCurrentReplicas(job TrainingJob) (currentReplicas int) {
	return len(job.AllPods()) - 1
}

type ScaleOutEDLJobArgs struct {
	ScaleEDLJobArgs `yaml:",inline"`
}

func (submitArgs *ScaleOutEDLJobArgs) prepare() (err error) {
	log.Debugf("scaleoutEnvs: %v", scaleoutEnvs)
	if len(scaleoutEnvs) > 0 {
		submitArgs.Envs = transformSliceToMap(scaleoutEnvs, "=")
	}

	submitArgs.Timeout = int(scaleoutDuration.Seconds())

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
		maxWorkers := getEDLJobMaxReplicas(job)
		log.Debugf("currentWorkers: %v, maxWorkers: %v", currentWorkers, maxWorkers)
		if currentWorkers+submitArgs.Count > maxWorkers {
			return fmt.Errorf("The number of scaling out plus the number of current workers exceeds the max-workers. please try again later.")
		}
		return nil
	} else {
		return fmt.Errorf("the job: %s status: %s , is not RUNNING or SCALING, please try again later.", edljobName, job.GetStatus())
	}
}

func submitScaleOutEDLJob(args []string, submitArgs *ScaleOutEDLJobArgs) (err error) {
	err = submitArgs.prepare()
	if err != nil {
		return err
	}
	scaleName := fmt.Sprintf("%s-%d", submitArgs.Name, time.Now().Unix())
	log.Debugf("submitArgs: %v", submitArgs)
	err = workflow.SubmitJob(scaleName, submitArgs.Mode, namespace, submitArgs, scaleout_edl_chart)
	if err != nil {
		return err
	}

	log.Infof("The scaleout job %s has been submitted successfully", scaleName)
	return nil
}
