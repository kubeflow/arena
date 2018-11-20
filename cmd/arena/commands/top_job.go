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

	"github.com/kubeflow/arena/util/helm"
	"strconv"
	"text/tabwriter"
	"k8s.io/api/core/v1"
)

func NewTopJobCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "job",
		Short: "Display Resource (GPU) usage of jobs.",
		Run: func(cmd *cobra.Command, args []string) {
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

			// allPods, err := acquireAllActivePods(client)
			// if err != nil {
			// 	fmt.Println(err)
			// 	os.Exit(1)
			// }

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
			var topPodName string
			if len(args) > 0 {
				topPodName = args[0]
			}

			for name, ns := range releaseMap {
				if topPodName != "" {
					if name != topPodName {
						continue
					}
				}
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
					log.Debugf("Unkown chart %s\n", name)
				}

			}

			jobs = makeTrainingJobOrderdByGPUCount(jobs)
			// TODO(cheyang): Support different job describer, such as MPI job/tf job describer
			showGpuMetric := topPodName != ""
			topTrainingJob(jobs, showGpuMetric)

		},
	}

	// command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}



func topTrainingJob(jobInfoList []TrainingJob, showSpecific bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalAllocatedGPUs int64
		totalRequestedGPUs int64
	)
	showSpecific = showSpecific && GpuMonitoringInstalled(clientset)
	labelField := []string{"NAME", "STATUS", "TRAINER", "AGE", "NODE", "GPU(Requests)", "GPU(Allocated)"}
	if showSpecific{
		labelField = []string{"INSTANCE NAME", "STATUS", "NODE", "GPU(Device Index)", "GPU(Duty Cycle)", "GPU(Memory MiB)"}
	}

	PrintLine(w, labelField...)

	if showSpecific {
		for _, jobInfo := range jobInfoList {
			pods := jobInfo.AllPods()
			gpuMetric, err := GetJobGpuMetric(clientset, jobInfo)
			if err != nil {
				log.Errorf("Failed Query job %s GPU metric, err: %++v", err)
				continue
			}
			for _, pod := range pods {
				hostIP := ""
				status := string(pod.Status.Phase)
				if pod.Status.Phase == v1.PodRunning {
					hostIP = pod.Status.HostIP
				}
				if podMetric, ok := gpuMetric[pod.Name]; !ok || len(podMetric) == 0 {
					PrintLine(w,
						pod.Name,
						status,
						hostIP,
						"N/A",
						"N/A",
						"N/A",
					)
				}else {
					index := 0
					for guid, gpuMetric := range podMetric {
						podName := pod.Name
						if index != 0 {
							podName = ""
							hostIP = ""
							status = ""
						}
						PrintLine(w,
							podName,
							status,
							hostIP,
							guid,
							fmt.Sprintf("%.0f%%", gpuMetric.GpuDutyCycle),
							fmt.Sprintf("%.0fMiB / %.0fMiB ", fromByteToMiB(gpuMetric.GpuMemoryUsed) ,  fromByteToMiB(gpuMetric.GpuMemoryTotal) ),
						)
						index ++
					}
				}
			}
		}
	}else {
		for _, jobInfo := range jobInfoList {

			hostIP := jobInfo.HostIPOfChief()
			requestedGPU := jobInfo.RequestedGPU()
			allocatedGPU := jobInfo.AllocatedGPU()
			// status, hostIP := jobInfo.getStatus()
			totalAllocatedGPUs += allocatedGPU
			totalRequestedGPUs += requestedGPU
			PrintLine(w, jobInfo.Name(),
				jobInfo.GetStatus(),
				jobInfo.Trainer(),
				jobInfo.Age(),
				hostIP,
				strconv.FormatInt(requestedGPU, 10),
				strconv.FormatInt(allocatedGPU, 10),
			)
		}
	}

	if !showSpecific {
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Total Allocated GPUs of Training Job:\n")
		fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalAllocatedGPUs, 10))
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Total Requested GPUs of Training Job:\n")
		fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalRequestedGPUs, 10))
	}

	_ = w.Flush()
}

func fromByteToMiB(value float64) float64 {
	return value / 1048576
}
