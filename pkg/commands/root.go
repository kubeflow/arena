package commands

import (
	datacommand "github.com/kubeflow/arena/pkg/commands/data"
	"github.com/kubeflow/arena/pkg/commands/serving"
	topcommand "github.com/kubeflow/arena/pkg/commands/top"
	"github.com/kubeflow/arena/pkg/commands/training"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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
	command.PersistentFlags().StringP("namespace", "n", "default", "the namespace of the job")
	command.AddCommand(training.NewSubmitCommand())
	command.AddCommand(training.NewScaleOutCommand())
	command.AddCommand(training.NewScaleInCommand())
	command.AddCommand(serving.NewServeCommand())
	command.AddCommand(training.NewListCommand())
	command.AddCommand(training.NewPruneCommand())
	command.AddCommand(training.NewGetCommand())
	command.AddCommand(training.NewLogViewerCommand())
	command.AddCommand(training.NewLogsCommand())
	command.AddCommand(training.NewDeleteCommand())
	command.AddCommand(topcommand.NewTopCommand())
	command.AddCommand(NewVersionCmd(CLIName))
	command.AddCommand(datacommand.NewDataCommand())
	command.AddCommand(NewCompletionCommand())
	return command
}
