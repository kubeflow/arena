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
	"time"

	"github.com/kubeflow/arena/pkg/util"
	"k8s.io/api/core/v1"
)

func NewTopJobCommand() *cobra.Command {
	var printNotStop bool
	var instanceName string
	var command = &cobra.Command{
		Use:   "job",
		Short: "Display Resource (GPU) usage of jobs.",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
			setupKubeconfig()
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
				topJobName    string
				trainingType  string
				job           TrainingJob
				jobs          []TrainingJob
				showGpuMetric bool
			)

			if len(args) > 0 {
				// enable GPU Metrics
				showGpuMetric = true
				topJobName = args[0]
				job, err = searchTrainingJob(topJobName, "", namespace)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				trainingType = job.Trainer()
			}

		printStart:
			if showGpuMetric {
				jobs = []TrainingJob{}
				job, err = searchTrainingJob(topJobName, trainingType, namespace)
				if err != nil {
					fmt.Printf("Failed to show GPU metrics due to %v", err)
				}
				jobs = append(jobs, job)
			} else {
				if printNotStop {
					fmt.Println("You must specify the job name when using `-r` flag")
					os.Exit(1)
				}

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
					jobs = append(jobs, trainingJobs...)
				}
			}

			jobs = makeTrainingJobOrderdByGPUCount(jobs)
			// TODO(cheyang): Support different job describer, such as MPI job/tf job describer
			topTrainingJob(jobs, showGpuMetric, instanceName, printNotStop)
			if printNotStop {
				goto printStart
			}
		},
	}

	command.Flags().BoolVarP(&printNotStop, "refresh", "r", false, "Display continuously")
	command.Flags().StringVarP(&instanceName, "instance", "i", "", "Display instance top info")
	command.Flags().BoolVar(&allNamespaces, "allNamespaces", false, "show all the namespaces")

	return command
}

func topTrainingJob(jobInfoList []TrainingJob, showSpecificJobMetric bool, instanceName string, notStop bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalAllocatedGPUs int64
		totalRequestedGPUs int64
	)
	if showSpecificJobMetric {
		if !GpuMonitoringInstalled(clientset) {
			fmt.Fprint(w, fmt.Sprintf("You haven't deployed prometheus and gpu metrics exporter for your k8s cluster, please deploy prometheus and gpu metric exporter following %s", PROMETHEUS_INSTALL_DOC_URL))
			_ = w.Flush()
			return
		}
	}
	labelField := []string{"NAME", "GPU(Requests)", "GPU(Allocated)", "STATUS", "TRAINER", "AGE", "NODE"}
	if showSpecificJobMetric {
		labelField = []string{"INSTANCE NAME", "GPU(Device Index)", "GPU(Duty Cycle)", "GPU(Memory MiB)", "STATUS", "NODE"}
	}
	PrintLine(w, labelField...)

	if showSpecificJobMetric {
		for _, jobInfo := range jobInfoList {
			pods := jobInfo.AllPods()
			gpuMetric, err := GetJobGpuMetric(clientset, jobInfo)
			if err != nil {
				log.Debug("Failed Query job %s GPU metric, err: %++v", jobInfo.Name(), err)
				continue
			}
			for _, pod := range pods {
				if instanceName != "" && instanceName != pod.Name {
					continue
				}
				hostIP := ""
				status := string(pod.Status.Phase)
				if pod.Status.Phase == v1.PodRunning {
					hostIP = pod.Status.HostIP
				}
				if podMetric, ok := gpuMetric[pod.Name]; !ok || len(podMetric) == 0 {
					PrintLine(w,
						pod.Name,
						"N/A",
						"N/A",
						"N/A",
						status,
						hostIP,
					)
				} else {
					index := 0
					keys := SortMapKeys(podMetric)
					for _, gid := range keys {
						podName := pod.Name
						if index != 0 {
							podName = ""
						}
						guid := gid
						gpuMetric := podMetric[gid]
						PrintLine(w,
							podName,
							guid,
							fmt.Sprintf("%.0f%%", gpuMetric.GpuDutyCycle),
							fmt.Sprintf("%.1fMiB / %.1fMiB ", fromByteToMiB(gpuMetric.GpuMemoryUsed), fromByteToMiB(gpuMetric.GpuMemoryTotal)),
							status,
							hostIP,
						)
						index++
					}
				}
			}
		}
	} else {
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
	}

	if !showSpecificJobMetric {
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Total Allocated GPUs of Training Job:\n")
		fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalAllocatedGPUs, 10))
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Total Requested GPUs of Training Job:\n")
		fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalRequestedGPUs, 10))
	}

	if notStop {
		fmt.Print("\033[2J")
		fmt.Print("\033[0;0H")
		_ = w.Flush()
		time.Sleep(1 * time.Second)
		return
	}
	_ = w.Flush()
}

func fromByteToMiB(value float64) float64 {
	return value / 1048576
}
