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
	"os"

	serveget "github.com/kubeflow/arena/pkg/printer/serving/get"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	//"io"
)

var (
	// get format from command option
	printFormat string
	// get serving type from command option
	stype string
)

// NewServingGetCommand starts the command
func NewServingGetCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "get ServingJobName",
		Short: "display details of a serving job",
		Run: func(cmd *cobra.Command, args []string) {
			// no serving name is an error
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			// set loglevel
			util.SetLogLevel(logLevel)
			// initate kubenetes client
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				log.Errorf(err.Error())
				os.Exit(1)
			}
			servingName := args[0]
			serveget.ServingGetExecute(client, servingName, namespace, stype, servingVersion, printFormat)
		},
	}
	//command.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "all namespace")
	command.Flags().StringVar(&servingVersion, "version", "", "assign the serving job version")
	command.Flags().StringVar(&printFormat, "format", "wide", `set the print format,format can be "yaml" or "json"`)
	command.Flags().StringVar(&stype, "type", "", `assign the serving job type,type can be "tf"("tensorflow"),"trt"("tensorrt"),"custom"`)

	return command

}
