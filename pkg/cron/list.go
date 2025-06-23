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

package cron

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"gopkg.in/yaml.v2"
)

func ListCrons(namespace string, allNamespaces bool) ([]*types.CronInfo, error) {
	return GetCronHandler().ListCrons(namespace, allNamespaces)
}

func DisplayAllCrons(crons []*types.CronInfo, allNamespaces bool, format types.FormatStyle) {
	switch format {
	case "json":
		for _, cron := range crons {
			data, _ := json.MarshalIndent(cron, "", "    ")
			fmt.Printf("%v", string(data))
		}
		return
	case "yaml":
		for _, cron := range crons {
			data, _ := yaml.Marshal(cron)
			fmt.Printf("%v", string(data))
		}
		return
	case "", "wide":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		var header []string
		if allNamespaces {
			header = append(header, "NAMESPACE")
		}
		header = append(header, []string{"NAME", "TYPE", "SCHEDULE", "SUSPEND", "DEADLINE", "CONCURRENCYPOLICY"}...)
		printLine(w, header...)

		for _, cron := range crons {
			var items []string
			if allNamespaces {
				items = append(items, cron.Namespace)
			}

			items = append(items, []string{
				cron.Name,
				cron.Type,
				cron.Schedule,
				strconv.FormatBool(cron.Suspend),
				cron.Deadline,
				cron.ConcurrencyPolicy,
			}...)
			printLine(w, items...)
		}
		_ = w.Flush()
		return
	}
}
