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

	arenaVersion "github.com/kubeflow/arena/pkg/version"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	var (
		short bool
	)
	versionCmd := cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Print version information"),
		Run: func(cmd *cobra.Command, args []string) {
			version := arenaVersion.GetVersion()
			fmt.Printf("Version: %s\n", version)
			if short {
				return
			}
			fmt.Printf("BuildDate: %s\n", version.BuildDate)
			fmt.Printf("GitCommit: %s\n", version.GitCommit)
			if version.GitTag != "" {
				fmt.Printf("GitTag: %s\n", version.GitTag)
			}
			fmt.Printf("GoVersion: %s\n", version.GoVersion)
			fmt.Printf("Compiler: %s\n", version.Compiler)
			fmt.Printf("Platform: %s\n", version.Platform)
		},
	}
	versionCmd.Flags().BoolVar(&short, "short", false, "print just the version number")
	return &versionCmd
}
