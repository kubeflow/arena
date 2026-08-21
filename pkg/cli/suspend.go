package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/client"
)

func newSuspendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend <name>",
		Short: "Suspend a running training job",
		Long:  `Suspend a running training job by setting spec.runPolicy.suspend to true.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFormat(); err != nil {
				return err
			}
			name := args[0]

			k8sClient, err := client.NewClient(kubeconfig, kubeContext)
			if err != nil {
				return fmt.Errorf("failed to create K8s client: %w", err)
			}

			ns := resolveNS("")

			jobType, err := suspendJob(cmdContext(cmd), k8sClient, ns, name)
			if err != nil {
				return err
			}

			return printActionResult(cmd.OutOrStdout(), name, ns, jobType, "suspended")
		},
	}

	registerOutputFlag(cmd)
	cmd.ValidArgsFunction = completeJobName
	return cmd
}

func suspendJob(ctx context.Context, k8sClient *client.Client, namespace, name string) (string, error) {
	jobType, err := detectJobType(ctx, k8sClient, namespace, name)
	if err != nil {
		return "", err
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"runPolicy": map[string]interface{}{
				"suspend": true,
			},
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return "", fmt.Errorf("failed to marshal patch: %w", err)
	}

	err = k8sClient.Patch(ctx, jobType, namespace, name, patchBytes)
	if err != nil {
		return "", fmt.Errorf("failed to suspend %s %q: %w", jobType, name, err)
	}

	return jobType, nil
}
