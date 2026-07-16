package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/client"
)

var suspendCmd = &cobra.Command{
	Use:   "suspend <name>",
	Short: "Suspend a running training job",
	Long:  `Suspend a running training job by setting spec.runPolicy.suspend to true.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		fmt.Printf("%s/%s suspended\n", strings.ToLower(jobType), name)
		return nil
	},
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
		return "", fmt.Errorf("failed to suspend %s %s: %w", jobType, name, err)
	}

	return jobType, nil
}

func init() {
	jobCmd.AddCommand(suspendCmd)
}
