package cli

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/log"
)

var (
	kubeconfig  string
	kubeContext string
	namespace   string
	debugMode   bool
	verbose     int32
)

var rootCmd = &cobra.Command{
	Use:   "arena",
	Short: "Arena v2 - AI workload CLI for Kubernetes",
	Long:  `Arena v2 is a lightweight CLI for submitting AI training jobs to Kubernetes.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if err := log.SetVerbosity(flag.CommandLine, verbose); err != nil {
			log.Warning("failed to set verbosity", "error", err.Error())
		}
	},
}

func Execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	rootCmd.SetContext(ctx)
	return rootCmd.Execute()
}

// ExecuteWithArgs executes the root command with the given arguments.
// It is primarily used for integration testing where the CLI needs to be
// invoked programmatically with specific flags and subcommands.
func ExecuteWithArgs(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	rootCmd.SetContext(ctx)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// DebugMode returns whether debug mode is enabled.
func DebugMode() bool {
	return debugMode
}

func init() {
	// Initialize klog with the standard flag set
	log.Init(flag.CommandLine)

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&kubeContext, "context", "", "kubeconfig context to use")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace (priority: flag > YAML > kubeconfig context > default)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode with detailed error output")
	rootCmd.PersistentFlags().Int32VarP(&verbose, "verbose", "v", 0, "verbosity level (higher = more detailed logs)")
}
