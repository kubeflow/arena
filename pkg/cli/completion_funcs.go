package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
)

// completionTimeout bounds how long a dynamic completion function may block.
// klog writes to stderr, so debug logs won't interfere with cobra's stdout protocol.
const completionTimeout = 3 * time.Second

// completeFrameworkType provides shell completion for the framework type argument
// of `arena job submit <type>`. Only completes the first positional arg; subsequent
// args (the run command) get no completions.
func completeFrameworkType(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{
		"pytorch\tPyTorch",
		"pytorchjob\tPyTorch",
		"tensorflow\tTensorFlow",
		"tfjob\tTensorFlow",
		"tf\tTensorFlow",
		"mpi\tMPI",
		"mpijob\tMPI",
		"horovod\tHorovod",
		"deepspeed\tDeepSpeed",
		"ray\tRay",
	}, cobra.ShellCompDirectiveNoFileComp
}

// completeJobName provides dynamic shell completion for job names by querying
// the Kubernetes cluster. Used by commands that take a job name as their first
// positional argument (get, status, logs, delete, suspend, resume).
// Silently returns an empty list on any error (cluster unreachable, CRD missing, etc.)
// to avoid polluting the terminal during completion.
func completeJobName(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	k8sClient, err := client.NewClient(kubeconfig, kubeContext)
	if err != nil {
		log.Debug("completion: failed to create k8s client", "error", err.Error())
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx, cancel := context.WithTimeout(cmdContext(cmd), completionTimeout)
	defer cancel()

	ns := resolveNS("")

	mpiAvailable := true
	if err := k8sClient.ResolveMPIVersion(ctx); err != nil {
		mpiAvailable = false
	}

	var completions []string
	for _, kind := range supportedJobKinds {
		if kind == constants.KindMPIJob && !mpiAvailable {
			continue
		}
		jobs, err := k8sClient.List(ctx, kind, ns, v2LabelSelector)
		if err != nil {
			log.Debug("completion: failed to list jobs", "kind", kind, "error", err.Error())
			continue
		}
		for _, job := range jobs {
			completions = append(completions, job.GetName())
		}
	}
	if len(completions) == 0 {
		log.Debug("completion: no jobs found", "namespace", ns)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeOutputFormat provides shell completion for the --output/-o flag.
func completeOutputFormat(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"table\tTable format",
		"wide\tWide table",
		"json\tJSON output",
		"yaml\tYAML output",
	}, cobra.ShellCompDirectiveNoFileComp
}

// completeFile delegates to the shell's native file path completion.
func completeFile(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveDefault
}

// completeStaticChoices returns a completion function that suggests a fixed set
// of values with no file completion.
func completeStaticChoices(choices ...string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return choices, cobra.ShellCompDirectiveNoFileComp
	}
}
