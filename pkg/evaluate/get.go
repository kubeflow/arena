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

package evaluate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"gopkg.in/yaml.v2"
)

var getEvaluateJobTemplate = `
JobID:              %v
Name:               %v
Namespace:          %v
ModelName:          %v
ModelPath:          %v
ModelVersion:       %v
DatasetPath:        %v
Status:             %v
CreationTimestamp:  %v
`

func GetEvaluateJob(name, namespace string) (*types.EvaluateJobInfo, error) {
	job, err := k8saccesser.GetK8sResourceAccesser().GetJob(name, namespace)
	if err != nil {
		return nil, err
	}
	return buildEvaluateJob(job), nil
}

func DisplayEvaluateJob(job *types.EvaluateJobInfo, format types.FormatStyle) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(job, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case "yaml":
		data, _ := yaml.Marshal(job)
		fmt.Printf("%v", string(data))
		return
	case "", "wide":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		printLine(w, fmt.Sprintf(strings.Trim(getEvaluateJobTemplate, "\n"),
			job.JobID,
			job.Name,
			job.Namespace,
			job.ModelName,
			job.ModelPath,
			job.ModelVersion,
			job.DatasetPath,
			job.Status,
			job.CreationTimestamp,
		))

		_ = w.Flush()
		return
	}
}
