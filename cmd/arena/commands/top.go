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
	"github.com/spf13/cobra"
	// podv1 "k8s.io/api/core/v1"
)

var (
	topLong = `Display Resource (GPU) usage.

Available Commands:
  node        Display Resource (GPU) usage of nodes
  job         Display Resource (GPU) usage of pods
    `
)

func NewTopCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "top",
		Short: "Display Resource (GPU) usage.",
		Long:  topLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	// create subcommands
	command.AddCommand(NewTopNodeCommand())
	command.AddCommand(NewTopJobCommand())

	return command
}
