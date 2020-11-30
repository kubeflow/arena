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
	"github.com/kubeflow/arena/pkg/util"
)

type SimpleJobInfo struct {
	Name      string `json:"name" yaml:"name"`
	Status    string `json:"status" yaml:"status"`
	Trainer   string `json:"trainer" yaml:"trainer"`
	Age       string `json:"age" yaml:"age"`
	Node      string `json:"node" yaml:"node"`
	Namespace string `json:"namespace" yaml:"namespace"`
}

func ListTrainingJobs(namespace string, allNamespaces bool) ([]TrainingJob, error) {
	jobs := []TrainingJob{}
	trainers := GetAllTrainers()
	for _, trainer := range trainers {
		if !trainer.IsEnabled() {
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
	jobSimpleInfos := []SimpleJobInfo{}
	jobInfos := []*types.TrainingJobInfo{}
	for _, jobInfo := range jobInfoList {
		jobSimpleInfos = append(jobSimpleInfos, SimpleJobInfo{
			Name:      jobInfo.Name(),
			Status:    GetJobRealStatus(jobInfo),
			Trainer:   strings.ToUpper(string(jobInfo.Trainer())),
			Age:       util.ShortHumanDuration(jobInfo.Age()),
			Node:      jobInfo.HostIPOfChief(),
			Namespace: jobInfo.Namespace(),
		})
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
		labelField := []string{"NAME", "STATUS", "TRAINER", "AGE", "NODE"}
		if allNamespaces {
			labelField = append(labelField, "NAMESPACE")
		}
		PrintLine(w, labelField...)
		for _, jobSimpleInfo := range jobSimpleInfos {
			items := []string{
				jobSimpleInfo.Name,
				jobSimpleInfo.Status,
				jobSimpleInfo.Trainer,
				jobSimpleInfo.Age,
				jobSimpleInfo.Node,
			}
			if allNamespaces {
				items = append(items, jobSimpleInfo.Namespace)
			}
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
