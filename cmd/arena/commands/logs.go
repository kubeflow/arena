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
	"io"
	"os"
	"strconv"
	"time"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type logEntry struct {
	displayName string
	pod         string
	time        time.Time
	line        string
}

func NewLogsCommand() *cobra.Command {
	var (
		printer   logPrinter
		since     string
		sinceTime string
		tail      int64
	)
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
			setupKubeconfig()
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

			printer.kubeClient = kubernetes.NewForConfigOrDie(conf)
			if tail > 0 {
				printer.tail = &tail
			}
			if sinceTime != "" {
				parsedTime, err := time.Parse(time.RFC3339, sinceTime)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				meta1Time := metav1.NewTime(parsedTime)
				printer.sinceTime = &meta1Time
			} else if since != "" {
				parsedSince, err := strconv.ParseInt(since, 10, 64)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				printer.sinceSeconds = &parsedSince
			}
			err = updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			// podName, err := getPodNameFromJob(printer.kubeClient, namespace, name)
			job, err := searchTrainingJob(name, trainingType, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			var pod v1.Pod

			if printer.pod != "" {
				for _, p := range job.AllPods() {
					if p.Name == printer.pod {
						pod = p
						break
					}
				}
				if pod.Name == "" {
					fmt.Printf("The instance %s is not belong to the job %s, so no log can be found. Please run 'arena get %s' to check status.", printer.pod, name, name)
					os.Exit(1)
				}
			} else {
				pod = job.ChiefPod()
			}

			if pod.Name == "" {
				fmt.Printf("The chief instance is not found, it should be deleted. Please run 'arena get %s' to check the INSTANCE column, and run 'arena logs %s -i INSTANCE_NAME' to check log.\n",
					name,
					name)
				fmt.Println("\nTo avoid that the instances are deleted automatically, you can set cleanTaskPolicy as None when submiting jobs. It means that all the instances are kept after training, and will hold the resources unless you can clean it up manully.")
				os.Exit(1)
			}

			err = printer.PrintPodLogs(pod.Name, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	command.Flags().StringVar(&trainingType, "type", "", "The training type to show logging, the possible option is tfjob, mpijob, horovodjob or standalonejob. (optional)")

	command.Flags().BoolVarP(&printer.follow, "follow", "f", false, "Specify if the logs should be streamed.")
	command.Flags().StringVar(&since, "since", "", "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVar(&sinceTime, "since-time", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().Int64Var(&tail, "tail", -1, "Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines otherwise 10, if a selector is provided.")
	command.Flags().BoolVar(&printer.timestamps, "timestamps", false, "Include timestamps on each line in the log output")

	// command.Flags().StringVar(&printer.pod, "instance", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVarP(&printer.pod, "instance", "i", "", "Specify the task instance to get log")
	return command
}

type logPrinter struct {
	pod          string
	follow       bool
	sinceSeconds *int64
	sinceTime    *metav1.Time
	tail         *int64
	timestamps   bool
	kubeClient   kubernetes.Interface
}

// PrintPodLogs prints logs for a single pod
func (p *logPrinter) PrintPodLogs(podName, namespace string) error {

	var logs []logEntry

	err := p.getPodLogs("", podName, namespace, p.follow, p.tail, p.sinceSeconds, p.sinceTime, func(entry logEntry) {
		logs = append(logs, entry)
	})
	if err != nil {
		return err
	}
	for _, entry := range logs {
		p.printLogEntry(entry)
	}
	return nil
}

func (p *logPrinter) getPodLogs(displayName string, podName string, podNamespace string, follow bool, tail *int64, sinceSeconds *int64, sinceTime *metav1.Time, callback func(entry logEntry)) error {
	err := p.ensureContainerStarted(podName, podNamespace, 5, time.Millisecond)
	if err != nil {
		return err
	}
	readCloser, err := p.kubeClient.CoreV1().Pods(podNamespace).GetLogs(podName, &v1.PodLogOptions{
		// Container:    p.container,
		Follow:       follow,
		Timestamps:   true,
		SinceSeconds: sinceSeconds,
		SinceTime:    sinceTime,
		TailLines:    tail,
	}).Stream()

	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, readCloser)

	defer readCloser.Close()
	return err
}

func (p *logPrinter) printLogEntry(entry logEntry) {
	line := entry.line
	if p.timestamps {
		line = entry.time.Format(time.RFC3339) + "	" + line
	}
	fmt.Println(line)
}

func (p *logPrinter) ensureContainerStarted(podName string, podNamespace string, retryCnt int, retryTimeout time.Duration) error {
	for retryCnt > 0 {
		pod, err := p.kubeClient.CoreV1().Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if len(pod.Status.ContainerStatuses) == 0 {
			time.Sleep(retryTimeout)
			retryCnt--
			continue
		}

		var containerStatus *v1.ContainerStatus = &pod.Status.ContainerStatuses[0]
		// for _, status := range pod.Status.ContainerStatuses {
		// 	if status.Name == container {
		// 		containerStatus = &status
		// 		break
		// 	}
		// }
		if containerStatus == nil || containerStatus.State.Waiting != nil {
			time.Sleep(retryTimeout)
			retryCnt--
		} else {
			return nil
		}
	}
	return fmt.Errorf("pod '%s' has not been started. Please run 'arena get %s' to check status.", podName, name)
}
