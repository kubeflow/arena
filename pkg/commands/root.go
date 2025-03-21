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

package commands

import (
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kubeflow/arena/pkg/commands/cron"
	"github.com/kubeflow/arena/pkg/commands/data"
	"github.com/kubeflow/arena/pkg/commands/evaluate"
	"github.com/kubeflow/arena/pkg/commands/model"
	"github.com/kubeflow/arena/pkg/commands/serving"
	"github.com/kubeflow/arena/pkg/commands/top"
	"github.com/kubeflow/arena/pkg/commands/training"
)

const (
	// CLIName is the name of the CLI
	CLIName = "arena"
)

// NewCommand returns a new instance of an Arena command
func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:           CLIName,
		Short:         "arena is the command line interface to Arena",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	// enable logging
	command.PersistentFlags().String("loglevel", "info", "Set the logging level. One of: debug|info|warn|error")
	command.PersistentFlags().Bool("pprof", false, "enable cpu profile")
	command.PersistentFlags().Bool("trace", false, "enable trace")
	command.PersistentFlags().String("arena-namespace", "arena-system", "The namespace of arena system service, like tf-operator")
	command.PersistentFlags().String("config", "", "Path to a kube config. Only required if out-of-cluster")
	command.PersistentFlags().StringP("namespace", "n", "", "the namespace of the job")
	command.PersistentFlags().Bool("helm-binary", false, "use helm binary to submit job")
	command.AddCommand(training.NewSubmitCommand())
	command.AddCommand(training.NewScaleOutCommand())
	command.AddCommand(training.NewScaleInCommand())
	command.AddCommand(serving.NewServeCommand())
	command.AddCommand(training.NewListCommand())
	command.AddCommand(training.NewPruneCommand())
	command.AddCommand(training.NewGetCommand())
	command.AddCommand(training.NewAttachCommand())
	command.AddCommand(training.NewLogViewerCommand())
	command.AddCommand(training.NewLogsCommand())
	command.AddCommand(training.NewDeleteCommand())
	command.AddCommand(top.NewTopCommand())
	command.AddCommand(NewVersionCmd(CLIName))
	command.AddCommand(data.NewDataCommand())
	command.AddCommand(cron.NewCronCommand())
	command.AddCommand(NewCompletionCommand())
	command.AddCommand(evaluate.NewEvaluateCommand())
	command.AddCommand(NewWhoamiCommand())
	command.AddCommand(model.NewModelCommand())
	return command
}
