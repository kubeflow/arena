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
	"path"
	"time"

	podlogs "github.com/kubeflow/arena/pkg/podlogs"
	tlogs "github.com/kubeflow/arena/pkg/printer/base/logs"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

func NewLogsCommand() *cobra.Command {
	var outerArgs = &podlogs.OuterRequestArgs{}
	var command = &cobra.Command{
		Use:   "logs training job",
		Short: "print the logs for a task of the training job",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name = args[0]
			conf, err := clientConfig.ClientConfig()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			outerArgs.KubeClient = kubernetes.NewForConfigOrDie(conf)
			err = updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			// podName, err := getPodNameFromJob(printer.kubeClient, namespace, name)
			job, err := searchTrainingJob(name, "", namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			outerArgs.Namespace = namespace
			outerArgs.RetryCount = 5
			outerArgs.RetryTimeout = time.Millisecond
			names := []string{}
			for _, pod := range job.AllPods() {
				names = append(names, path.Base(pod.ObjectMeta.SelfLink))
			}
			chiefPod := job.ChiefPod()
			if len(names) > 1 && outerArgs.PodName == "" {
				names = []string{path.Base(chiefPod.ObjectMeta.SelfLink)}
			}
			logPrinter, err := tlogs.NewPodLogPrinter(names, outerArgs)
			if err != nil {
				log.Errorf(err.Error())
				os.Exit(1)
			}
			code, err := logPrinter.Print()
			if err != nil {
				log.Errorf("%s, %s", err.Error(), "please use \"runai get\" to get more information.")
				os.Exit(1)
			} else if code != 0 {
				os.Exit(code)
			}
		},
	}

	command.Flags().BoolVarP(&outerArgs.Follow, "follow", "f", false, "Specify if the logs should be streamed.")
	command.Flags().DurationVar(&outerArgs.SinceSeconds, "since", 0, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVar(&outerArgs.SinceTime, "since-time", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().IntVarP(&outerArgs.Tail, "tail", "t", -1, "Lines of recent log file to display. Defaults to -1 , showing all log lines.")
	command.Flags().BoolVar(&outerArgs.Timestamps, "timestamps", false, "Include timestamps on each line in the log output")

	// command.Flags().StringVar(&printer.pod, "instance", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVarP(&outerArgs.PodName, "pod", "p", "", "Specify the pod to get log")
	return command
}
