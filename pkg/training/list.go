// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package training

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
)

func ListTrainingJobs(namespace string, allNamespaces bool, jobType types.TrainingJobType) ([]TrainingJob, error) {
	jobs := []TrainingJob{}
	trainers := GetAllTrainers()
	if jobType == types.UnknownTrainingJob {
		return nil, fmt.Errorf("Unsupport job type,arena only supports: [%v]", utils.GetSupportTrainingJobTypesInfo())
	}
	var wg sync.WaitGroup
	locker := new(sync.RWMutex)
	noPrivileges := false
	for tType, t := range trainers {
		wg.Add(1)
		trainer := t
		trainerType := tType
		go func() {
			defer wg.Done()
			if !trainer.IsEnabled() {
				return
			}
			if !isNeededTrainingType(trainerType, jobType) {
				return
			}
			trainingJobs, err := trainer.ListTrainingJobs(namespace, allNamespaces)
			if err != nil {
				if strings.Contains(err.Error(), "forbidden: User") {
					item := fmt.Sprintf("namespace %v", namespace)
					if allNamespaces {
						item = "all namespaces"
					}
					log.Debugf("the user has no privileges to list the %v in %v,reason: %v", trainerType, item, err)
					noPrivileges = true
					return
				}
				log.Debugf("trainer %v failed to list training jobs: %v", trainerType, err)
				return
			}
			locker.Lock()
			jobs = append(jobs, trainingJobs...)
			locker.Unlock()
		}()
	}
	wg.Wait()
	if noPrivileges {
		item := fmt.Sprintf("namespace %v", namespace)
		if allNamespaces {
			item = "all namespaces"
		}
		return nil, fmt.Errorf("the user has no privileges to list the training jobs in %v", item)
	}
	jobs = makeTrainingJobOrderdByAge(jobs)
	return jobs, nil
}

func DisplayTrainingJobList(jobInfoList []TrainingJob, format string, allNamespaces bool) {
	jobInfos := []*types.TrainingJobInfo{}
	services, nodes := PrepareServicesAndNodesForTensorboard(jobInfoList, allNamespaces)
	switch format {
	case "json":
		for _, jobInfo := range jobInfoList {
			jobInfos = append(jobInfos, BuildJobInfo(jobInfo, true, services, nodes))
		}
		data, _ := json.MarshalIndent(jobInfos, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case "yaml":
		for _, jobInfo := range jobInfoList {
			jobInfos = append(jobInfos, BuildJobInfo(jobInfo, true, services, nodes))
		}
		data, _ := yaml.Marshal(jobInfos)
		fmt.Printf("%v", string(data))
		return
	case "", "wide":
		for _, jobInfo := range jobInfoList {
			jobInfos = append(jobInfos, BuildJobInfo(jobInfo, false, services, nodes))
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		header := []string{}
		if allNamespaces {
			header = append(header, "NAMESPACE")
		}
		header = append(header, []string{"NAME", "STATUS", "TRAINER", "DURATION", "GPU(Requested)", "GPU(Allocated)", "NODE"}...)
		PrintLine(w, header...)
		for _, jobInfo := range jobInfos {
			hostIP := "N/A"
			for _, i := range jobInfo.Instances {
				if i.IsChief {
					hostIP = i.NodeIP
				}
			}
			items := []string{}
			if allNamespaces {
				items = append(items, jobInfo.Namespace)
			}
			jobInfo.Duration = strings.Replace(jobInfo.Duration, "s", "", -1)
			duration, err := strconv.ParseInt(jobInfo.Duration, 10, 64)
			if err != nil {
				log.Debugf("failed to parse duration: %v", err)

			}
			allocatedGPUs := "N/A"
			if jobInfo.Status == types.TrainingJobPending || jobInfo.Status == types.TrainingJobRunning {
				allocatedGPUs = fmt.Sprintf("%v", jobInfo.AllocatedGPU)
			}
			items = append(items, []string{
				jobInfo.Name,
				fmt.Sprintf("%v", jobInfo.Status),
				strings.ToUpper(string(jobInfo.Trainer)),
				util.ShortHumanDuration(time.Duration(duration) * time.Second),
				fmt.Sprintf("%v", jobInfo.RequestGPU),
				allocatedGPUs,
				hostIP,
			}...)
			PrintLine(w, items...)
		}
		_ = w.Flush()
		return

	}
}

func PrintLine(w io.Writer, fields ...string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

func CheckPrintFormat(format string) error {
	switch format {
	case "yaml", "json", "wide", "":
		return nil
	}
	return fmt.Errorf("Unknown format,only support: [yaml,json,wide]")
}

func isNeededTrainingType(jobType types.TrainingJobType, targetJobType types.TrainingJobType) bool {
	if targetJobType == types.AllTrainingJob {
		return true
	}
	return jobType == targetJobType
}
