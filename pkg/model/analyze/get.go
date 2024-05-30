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

package analyze

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"gopkg.in/yaml.v2"
)

var getJobTemplate = `
Name:       %v
Namespace:  %v
Type:       %v
Status:     %v
Duration:   %v
Age:        %v
%v
`

func SearchModelJob(namespace, name string, modelJobType types.ModelJobType) (ModelJob, error) {
	if modelJobType == types.UnknownModelJob {
		return nil, fmt.Errorf("unknown model job type,arena only supports: [%s]", utils.GetSupportModelJobTypesInfo())
	}

	processor := NewModelProcessor(modelJobType)
	return processor.GetModelJob(namespace, name)
}

func PrintModelJob(job ModelJob, format types.FormatStyle) {
	switch format {
	case types.JsonFormat:
		data, _ := json.MarshalIndent(job.Convert2JobInfo(), "", "    ")
		fmt.Printf("%v", string(data))
		return
	case types.YamlFormat:
		data, _ := yaml.Marshal(job.Convert2JobInfo())
		fmt.Printf("%v", string(data))
		return
	}
	jobInfo := job.Convert2JobInfo()

	lines := []string{"Parameters:"}

	for k, v := range jobInfo.Params {
		lines = append(lines, fmt.Sprintf("\t%s\t%s", k, v))
	}

	totalGPUs := float64(0)
	for _, i := range jobInfo.Instances {
		totalGPUs += i.RequestGPUs
	}
	title := ""
	step := ""
	gpuLine := ""
	if totalGPUs != 0 {
		title = "\tGPU"
		step = "\t---"
		gpuLine = fmt.Sprintf("GPU:        %v", totalGPUs)
	}

	fragment := []string{gpuLine, "", "Instances:", fmt.Sprintf("  NAME\tSTATUS\tAGE\tREADY\tRESTARTS%v\tNODE", title)}
	lines = append(lines, fragment...)
	lines = append(lines, fmt.Sprintf("  ----\t------\t---\t-----\t--------%v\t----", step))
	for _, i := range jobInfo.Instances {
		value := fmt.Sprintf("%v", i.RequestGPUs)
		items := []string{
			fmt.Sprintf("  %v", i.Name),
			fmt.Sprintf("%v", i.Status),
			fmt.Sprintf("%v", i.Age),
			fmt.Sprintf("%v/%v", i.ReadyContainer, i.TotalContainer),
			fmt.Sprintf("%v", i.RestartCount),
		}
		if totalGPUs != 0 {
			items = append(items, value)
		}
		items = append(items, i.NodeName)
		lines = append(lines, strings.Join(items, "\t"))
	}
	lines = append(lines, "")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	output := fmt.Sprintf(strings.Trim(getJobTemplate, "\n"),
		jobInfo.Name,
		jobInfo.Namespace,
		jobInfo.Type,
		jobInfo.Status,
		jobInfo.Duration,
		jobInfo.Age,
		strings.Join(lines, "\n"),
	)
	fmt.Fprint(w, output)
	w.Flush()
}
