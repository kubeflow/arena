package cli

import (
	"bytes"
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

// NewRootCommand builds a fresh arena-v2 command tree. Every Execute call and
// every test constructs its own tree: pflag keeps per-flag state (Changed
// bits, parsed values) on the command objects and never resets it between
// parses, so sharing one tree across invocations leaks state.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "arena-v2",
		Short:         "Arena v2 - AI workload CLI for Kubernetes",
		Long:          `Arena v2 is a lightweight CLI for submitting AI training jobs to Kubernetes.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return log.SetVerbosity(flag.CommandLine, verbose)
		},
	}

	root.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	root.PersistentFlags().StringVar(&kubeContext, "context", "", "kubeconfig context to use")
	root.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace (priority: flag > YAML > kubeconfig context > default)")
	root.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode with detailed error output")
	root.PersistentFlags().Int32VarP(&verbose, "verbose", "v", 0, "verbosity level (higher = more detailed logs)")
	_ = root.RegisterFlagCompletionFunc("kubeconfig", completeFile)

	root.AddCommand(newJobCmd(), newSubmitCmd(), newTopCmd(), newCheckCmd(), newVersionCmd(), newCompletionCmd())

	return root
}

func Execute() error {
	return execute(NewRootCommand(), nil)
}

// ExecuteWithArgs executes a fresh root command with the given arguments.
// It is primarily used for integration testing where the CLI needs to be
// invoked programmatically with specific flags and subcommands.
func ExecuteWithArgs(args []string) error {
	return execute(NewRootCommand(), args)
}

// ExecuteWithArgsOutput is ExecuteWithArgs with stdout captured: everything
// the command writes via OutOrStdout (help, CRD dry-run output, tables) is
// returned alongside the error.
func ExecuteWithArgsOutput(args []string) (string, error) {
	root := NewRootCommand()
	var buf bytes.Buffer
	root.SetOut(&buf)
	err := execute(root, args)
	return buf.String(), err
}

func execute(root *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	root.SetContext(ctx)
	if args != nil {
		root.SetArgs(args)
	}
	return root.Execute()
}

// DebugMode returns whether debug mode is enabled.
func DebugMode() bool {
	return debugMode
}

func init() {
	// Initialize klog with the standard flag set
	log.Init(flag.CommandLine)
}
