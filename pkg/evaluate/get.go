package evaluate

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
	"text/tabwriter"
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
