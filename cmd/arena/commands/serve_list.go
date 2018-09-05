// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/kubeflow/arena/util/helm"
)

var serving_charts = map[string]bool{
	"tensorflow-serving-0.2.0": true,
}

func NewServingListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "list all the serving jobs",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)

			setupKubeconfig()
			_, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			releaseMap, err := helm.ListAllReleasesWithDetail()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			fmt.Fprintf(w, "NAME\tVERSION\tSTATUS\n")

			for name, cols := range releaseMap {
				log.Debugf("name: %s, cols: %s", name, cols)
				namespace := cols[len(cols)-1]
				chart := cols[len(cols)-2]
				status := cols[len(cols)-3]
				log.Debugf("namespace: %s, chart: %s, status:%s", namespace, chart, status)
				if serving_charts[chart] {
					index := strings.Index(name, "-")
					//serviceName := name[0:index]
					serviceVersion := ""
					if index > -1 {
						serviceVersion = name[index+1:]
					}
					nameAndVersion := strings.Split(name, "-")
					log.Debugf("nameAndVersion: %s, len(nameAndVersion): %d", nameAndVersion, len(nameAndVersion))
					fmt.Fprintf(w, "%s\t%s\t%s\n", name,
						serviceVersion,
						status)

				}
			}

			_ = w.Flush()
		},
	}

	return command
}
