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
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListEvaluateJobs(namespace string, allNamespaces bool) ([]*types.EvaluateJobInfo, error) {
	if allNamespaces {
		namespace = metav1.NamespaceAll
	}

	selector := fmt.Sprintf("app=%v", types.EvaluateJob)

	jobs, err := k8saccesser.GetK8sResourceAccesser().ListJobs(namespace, selector, "", nil)
	if err != nil {
		return nil, err
	}

	var evaluateJobs []*types.EvaluateJobInfo
	for _, job := range jobs {
		evaluateJob := buildEvaluateJob(job)
		evaluateJobs = append(evaluateJobs, evaluateJob)
	}
	return evaluateJobs, nil
}

func DisplayAllEvaluateJobs(jobs []*types.EvaluateJobInfo, allNamespaces bool, format types.FormatStyle) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(jobs, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case "yaml":
		data, _ := yaml.Marshal(jobs)
		fmt.Printf("%v", string(data))
		return
	case "", "wide":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		var header []string
		if allNamespaces {
			header = append(header, "NAMESPACE")
		}
		header = append(header, []string{"NAME", "MODEL_NAME", "MODEL_VERSION", "STATUS", "CREATE_TIME"}...)
		printLine(w, header...)

		for _, job := range jobs {
			var items []string
			if allNamespaces {
				items = append(items, job.Namespace)
			}

			items = append(items, []string{
				job.Name,
				job.ModelName,
				job.ModelVersion,
				job.Status,
				job.CreationTimestamp,
			}...)
			printLine(w, items...)
		}
		_ = w.Flush()
		return
	}
}
