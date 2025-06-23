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
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func ListModelJobs(namespace string, allNamespaces bool, modelJobType types.ModelJobType) ([]ModelJob, error) {
	if modelJobType == types.UnknownModelJob {
		return nil, fmt.Errorf("unknown serving job type,arena only supports: [%s]", utils.GetSupportModelJobTypesInfo())
	}

	processor := NewModelProcessor(modelJobType)
	return processor.ListModelJobs(namespace, allNamespaces)
}

func PrintAllModelJobs(jobs []ModelJob, allNamespaces bool, format types.FormatStyle) {
	var jobInfos []types.ModelJobInfo
	for _, job := range jobs {
		jobInfos = append(jobInfos, job.Convert2JobInfo())
	}
	switch format {
	case types.JsonFormat:
		data, _ := json.MarshalIndent(jobInfos, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case types.YamlFormat:
		data, _ := yaml.Marshal(jobInfos)
		fmt.Printf("%v", string(data))
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var header []string
	if allNamespaces {
		header = append(header, "NAMESPACE")
	}
	header = append(header, []string{"NAME", "STATUS", "TYPE", "DURATION", "AGE", "GPU(Requested)"}...)
	printLine(w, header...)
	for _, jobInfo := range jobInfos {
		var items []string
		if allNamespaces {
			items = append(items, jobInfo.Namespace)
		}
		jobInfo.Duration = strings.ReplaceAll(jobInfo.Duration, "s", "")
		duration, err := strconv.ParseInt(jobInfo.Duration, 10, 64)
		if err != nil {
			log.Debugf("failed to parse duration: %v", err)
		}

		items = append(items, []string{
			jobInfo.Name,
			jobInfo.Status,
			jobInfo.Type,
			util.ShortHumanDuration(time.Duration(duration) * time.Second),
			jobInfo.Age,
			fmt.Sprintf("%v", jobInfo.RequestGPUs),
		}...)
		printLine(w, items...)
	}
	_ = w.Flush()
}

func printLine(w io.Writer, fields ...string) {
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}
