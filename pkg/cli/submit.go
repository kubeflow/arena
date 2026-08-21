package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/task"
)

func newSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a training job",
		Long:  `Submit a training job using arena v1-compatible syntax`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			framework, v1TypeUnsupported := lookupFramework(args[0])
			if v1TypeUnsupported {
				return &CLIError{
					Message:     fmt.Sprintf("framework type %q is an arena v1 type not supported by arena-v2 yet", args[0]),
					ValidValues: v2FrameworkNames(),
				}
			}
			if framework == "" {
				return &CLIError{
					Message:     fmt.Sprintf("unsupported framework type: %q", args[0]),
					ValidValues: acceptedFrameworkTypes(),
				}
			}
			// The parent cannot mark --name/--image required at registration
			// (cobra would reject a bare `submit` before RunE prints help), so the
			// fallback path validates them here, mirroring the subcommands' error.
			if err := validateSubmitRequiredFlags(cmd); err != nil {
				return err
			}
			return runSubmit(cmd, framework, originalFramework(args[0]), args[1:])
		},
	}

	registerSubmitCommonFlags(cmd)
	registerPyTorchSubmitFlags(cmd)
	registerTFSubmitFlags(cmd)
	registerMPISubmitFlags(cmd)
	registerSubmitCompatFlags(cmd)

	cmd.ValidArgsFunction = completeFrameworkType

	// The parent keeps the full flag set — the fallback path and the test
	// suite parse it — but hides it from help: `submit -h` lists only the
	// framework subcommands, matching arena v1.
	cmd.Flags().VisitAll(func(f *pflag.Flag) { f.Hidden = true })

	for _, def := range frameworkRegistry {
		if def.canonical == "" {
			continue
		}
		cmd.AddCommand(newSubmitFrameworkSubcommand(def))
	}

	return cmd
}

// validateSubmitRequiredFlags reports unset --name/--image in the exact
// format cobra's ValidateRequiredFlags uses (including the missing-only list
// and lexical order), so fallback-path errors read identically to the
// per-framework subcommands'. Unlike MarkFlagRequired it mutates no shared
// command state, keeping a bare `submit` on the print-help path.
func validateSubmitRequiredFlags(cmd *cobra.Command) error {
	var missing []string
	for _, name := range []string{"image", "name"} {
		if f := cmd.Flags().Lookup(name); f != nil && !f.Changed {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	quoted := make([]string, 0, len(missing))
	for _, name := range missing {
		quoted = append(quoted, `"`+name+`"`)
	}
	return fmt.Errorf("required flag(s) %s not set", strings.Join(quoted, ", "))
}

// runSubmit is the shared submission core used by the parent submit command
// (the fallback path for anything that is not a framework subcommand) and by
// the per-framework subcommands. trailingArgs are the positional args that
// follow the framework type; they become the run command.
func runSubmit(cmd *cobra.Command, framework, originalFrameworkName string, trailingArgs []string) error {
	if err := validateOutputFormat(); err != nil {
		return err
	}
	if err := applyV1LogLevel(submitLogLevel); err != nil {
		return err
	}

	applyDryRunOutputDefault(cmd, submitDryRun)

	// Build the task with PyTorch N-1 conversion
	t := buildSubmitTask(framework, trailingArgs)

	// Apply overrides
	flags := buildSubmitFlags()

	if err := task.ApplyOverrides(t, flags); err != nil {
		return fmt.Errorf("failed to apply overrides: %w", err)
	}

	// Validate
	t.SetDefaults()
	if err := task.Validate(t); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	var (
		k8sClient *client.Client
		err       error
	)
	if !submitDryRun {
		k8sClient, err = client.NewClient(kubeconfig, kubeContext)
		if err != nil {
			return fmt.Errorf("failed to create K8s client: %w", err)
		}
	}
	return submitCRD(cmdContext(cmd), cmd.OutOrStdout(), k8sClient, t, originalFrameworkName, submitDryRun)
}

// buildSubmitTask constructs a Task from submit CLI flags with framework-specific conversions.
// PyTorch: --workers N means N total processes (1 master + N-1 workers), so worker.replicas = N-1.
// Other frameworks: --workers N means N workers, so worker.replicas = N.
func buildSubmitTask(framework string, trailingArgs []string) *task.Task {
	t := &task.Task{
		Name:      submitName,
		Image:     submitImage,
		Framework: task.Framework{Name: framework},
		Worker:    &task.Worker{Replicas: submitWorkers},
	}
	if t.Worker.Replicas < 1 {
		t.Worker.Replicas = 1
	}

	// v1 PyTorch compat: --workers N means N total (1 master + N-1 workers)
	// Conversion happens here in submit path (not in launch path)
	if framework == constants.FrameworkPyTorch {
		if submitWorkers <= 1 {
			// --workers=1 means master-only (no worker block)
			t.Worker = nil
			t.Master = &task.RoleConfig{}
		} else {
			t.Worker.Replicas = submitWorkers - 1
		}
	}

	// Set run from trailing args
	if len(trailingArgs) > 0 {
		t.Run = strings.Join(trailingArgs, " ")
	}

	return t
}

// applyV1LogLevel maps the deprecated v1 --loglevel flag onto v2's klog
// verbosity: debug->2, info->1, warn/error->0.
func applyV1LogLevel(level string) error {
	if level == "" {
		return nil
	}
	var verbosity int32
	switch strings.ToLower(level) {
	case "debug":
		verbosity = 2
	case "info":
		verbosity = 1
	case "warn", "error":
		verbosity = 0
	default:
		return fmt.Errorf("invalid --loglevel %q: must be debug, info, warn, or error", level)
	}
	if err := log.SetVerbosity(flag.CommandLine, verbosity); err != nil {
		return fmt.Errorf("failed to set log level: %w", err)
	}
	return nil
}
