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
	"time"
)

var (
	scaleout_chart = util.GetChartsFolder() + "/scaleout"
	scalein_chart  = util.GetChartsFolder() + "/scalein"
	scaleEnvs      []string
)

func NewScaleJobCommand() *cobra.Command {
	var (
		submitArgs submitScaleJobArgs
	)

	submitArgs.Mode = "scale"

	var command = &cobra.Command{
		Use:     "scale",
		Short:   "scale job",
		Aliases: []string{},
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
			if isScaleOut() {
				fmt.Println("scale out")
				submitArgs.Add = true
				err = submitScaleOutJob(args, &submitArgs)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else if isScaleIn() {
				fmt.Println("scale in")
				submitArgs.Delete = true
				err = submitScaleInJob(args, &submitArgs)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
		},
	}

	command.Flags().StringVar(&submitArgs.Name, "name", "", "required, edl job name")
	command.MarkFlagRequired("name")
	command.Flags().BoolVar(&submitArgs.Add, "add", false, "scale out")
	command.Flags().BoolVar(&submitArgs.Delete, "delete", false, "scale in")
	command.Flags().IntVar(&submitArgs.Timeout, "timeout", 60, "timeout of callback scaler script.")
	command.Flags().IntVar(&submitArgs.Retry, "retry", 0, "retry times.")
	command.Flags().IntVar(&submitArgs.Count, "count", 1, "the nums of you want to add or delete worker.")
	command.Flags().StringVar(&submitArgs.Script, "script", "", "script of scaling.")
	command.Flags().StringArrayVarP(&scaleEnvs, "env", "e", []string{}, "the environment variables.")
	return command
}

type submitScaleJobArgs struct {
	Mode string `yaml:"mode"` // --mode
	//--name string     required, edl job name
	Name string `yaml:"edlName"`
	//--add  bool       scale out
	Add bool `yaml:"add"`
	//--delete bool     scale in
	Delete bool `yaml:"Delete"`
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

func (submitArgs *submitScaleJobArgs) processScript() {
	if submitArgs.Script == "" {
		if submitArgs.Add {
			submitArgs.Script = "/usr/local/bin/scaler.sh --add"
		} else if submitArgs.Delete {
			submitArgs.Script = "/usr/local/bin/scaler.sh --delete"
		}
	}
}

func (submitArgs *submitScaleJobArgs) prepare() (err error) {
	submitArgs.processScript()
	log.Debugf("scaleEnvs: %v", scaleEnvs)
	if len(scaleEnvs) > 0 {
		submitArgs.Envs = transformSliceToMap(scaleEnvs, "=")
	}
	return nil
}

func submitScaleOutJob(args []string, submitArgs *submitScaleJobArgs) (err error) {
	err = submitArgs.prepare()
	if err != nil {
		return err
	}
	scaleName := fmt.Sprintf("%s-%d", submitArgs.Name, time.Now().Unix())
	log.Infof("submitArgs: %v", submitArgs)
	err = workflow.SubmitJob(scaleName, submitArgs.Mode, namespace, submitArgs, scaleout_chart)
	if err != nil {
		return err
	}

	log.Infof("The scaleout job %s has been submitted successfully", scaleName)
	return nil
}

func submitScaleInJob(args []string, submitArgs *submitScaleJobArgs) (err error) {
	err = submitArgs.prepare()
	if err != nil {
		return err
	}
	log.Infof("submitArgs: %v", submitArgs)
	scaleName := fmt.Sprintf("%s-%d", submitArgs.Name, time.Now().Unix())
	err = workflow.SubmitJob(scaleName, submitArgs.Mode, namespace, submitArgs, scalein_chart)
	if err != nil {
		return err
	}

	log.Infof("The scalein job %s has been submitted successfully", scaleName)
	return nil
}

func isScaleOut() bool {
	for _, arg := range os.Args {
		if arg == "--add" {
			return true
		}
	}
	return false
}

func isScaleIn() bool {
	for _, arg := range os.Args {
		if arg == "--delete" {
			return true
		}
	}
	return false
}
