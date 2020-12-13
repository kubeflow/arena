package training

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	yaml "gopkg.in/yaml.v2"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
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
	for _, jobInfo := range jobInfoList {
		jobInfos = append(jobInfos, BuildJobInfo(jobInfo))
	}
	switch format {
	case "json":
		data, _ := json.MarshalIndent(jobInfos, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case "yaml":
		data, _ := yaml.Marshal(jobInfos)
		fmt.Printf("%v", string(data))
		return
	case "", "wide":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		header := []string{}
		if allNamespaces {
			header = append(header, "NAMESPACE")
		}
		header = append(header, []string{"NAME", "STATUS", "TRAINER", "AGE", "GPU(Requested)", "GPU(Allocated)", "NODE"}...)
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
			items = append(items, []string{
				jobInfo.Name,
				fmt.Sprintf("%v", jobInfo.Status),
				strings.ToUpper(string(jobInfo.Trainer)),
				fmt.Sprintf("%v", jobInfo.Duration),
				fmt.Sprintf("%v", jobInfo.RequestGPU),
				fmt.Sprintf("%v", jobInfo.AllocatedGPU),
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
