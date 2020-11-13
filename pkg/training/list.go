package training

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	yaml "gopkg.in/yaml.v2"

	"github.com/kubeflow/arena/pkg/util"
)

type SimpleJobInfo struct {
	Name    string `json:"name" yaml:"name"`
	Status  string `json:"status" yaml:"status"`
	Trainer string `json:"trainer" yaml:"trainer"`
	Age     string `json:"age" yaml:"age"`
	Node    string `json:"node" yaml:"node"`
}

func ListTrainingJobs(namespace string, allNamespaces bool) ([]TrainingJob, error) {
	jobs := []TrainingJob{}
	trainers := NewSupportedTrainers()
	for _, trainer := range trainers {
		trainingJobs, err := trainer.ListTrainingJobs(namespace, allNamespaces)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, trainingJobs...)
	}
	jobs = makeTrainingJobOrderdByAge(jobs)
	return jobs, nil
}

func DisplayTrainingJobList(jobInfoList []TrainingJob, format string) {
	jobSimpleInfos := []SimpleJobInfo{}
	for _, jobInfo := range jobInfoList {
		jobSimpleInfos = append(jobSimpleInfos, SimpleJobInfo{
			Name:    jobInfo.Name(),
			Status:  GetJobRealStatus(jobInfo),
			Trainer: strings.ToUpper(jobInfo.Trainer()),
			Age:     util.ShortHumanDuration(jobInfo.Age()),
			Node:    jobInfo.HostIPOfChief(),
		})
	}
	switch format {
	case "json":
		data, _ := json.MarshalIndent(jobSimpleInfos, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case "yaml":
		data, _ := yaml.Marshal(jobSimpleInfos)
		fmt.Printf("%v", string(data))
		return
	case "", "wide":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		labelField := []string{"NAME", "STATUS", "TRAINER", "AGE", "NODE"}
		PrintLine(w, labelField...)
		for _, jobSimpleInfo := range jobSimpleInfos {
			PrintLine(w,
				jobSimpleInfo.Name,
				jobSimpleInfo.Status,
				jobSimpleInfo.Trainer,
				jobSimpleInfo.Age,
				jobSimpleInfo.Node,
			)
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
