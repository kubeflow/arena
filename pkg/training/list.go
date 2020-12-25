package training

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
	for trainerType, trainer := range trainers {
		if !trainer.IsEnabled() {
			continue
		}
		if !isNeededTrainingType(trainerType, jobType) {
			continue
		}
		trainingJobs, err := trainer.ListTrainingJobs(namespace, allNamespaces)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, trainingJobs...)
	}
	jobs = makeTrainingJobOrderdByAge(jobs)
	return jobs, nil
}

func DisplayTrainingJobList(jobInfoList []TrainingJob, format string, allNamespaces bool) {
	jobInfos := []*types.TrainingJobInfo{}
	switch format {
	case "json":
		for _, jobInfo := range jobInfoList {
			jobInfos = append(jobInfos, BuildJobInfo(jobInfo, true))
		}
		data, _ := json.MarshalIndent(jobInfos, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case "yaml":
		for _, jobInfo := range jobInfoList {
			jobInfos = append(jobInfos, BuildJobInfo(jobInfo, true))
		}
		data, _ := yaml.Marshal(jobInfos)
		fmt.Printf("%v", string(data))
		return
	case "", "wide":
		for _, jobInfo := range jobInfoList {
			jobInfos = append(jobInfos, BuildJobInfo(jobInfo, false))
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
	return fmt.Errorf("Unknown format,only suppot: [yaml,json,wide]")
}

func isNeededTrainingType(jobType types.TrainingJobType, targetJobType types.TrainingJobType) bool {
	if targetJobType == types.AllTrainingJob {
		return true
	}
	return jobType == targetJobType
}
