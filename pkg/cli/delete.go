package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/task"
)

var deleteFile string

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a training job",
		Long:  `Delete a training job by name or YAML file (similar to kubectl delete -f).`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string

			var yamlNS string
			switch {
			case deleteFile != "":
				t, err := task.LoadFromFile(deleteFile)
				if err != nil {
					return fmt.Errorf("failed to load file %q: %w", deleteFile, err)
				}
				name = t.Name
				if name == "" {
					return fmt.Errorf("file %q does not specify a job name", deleteFile)
				}
				yamlNS = t.Namespace
			case len(args) > 0:
				name = args[0]
			default:
				return errors.New("either job name or -f flag is required")
			}

			k8sClient, err := client.NewClient(kubeconfig, kubeContext)
			if err != nil {
				return fmt.Errorf("failed to create K8s client: %w", err)
			}

			ns := resolveNS(yamlNS)
			jobType, err := detectJobType(cmdContext(cmd), k8sClient, ns, name)
			if err != nil {
				return err
			}

			err = k8sClient.Delete(cmdContext(cmd), jobType, ns, name)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s/%s deleted\n", strings.ToLower(jobType), name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&deleteFile, "file", "f", "", "path to YAML file")
	cmd.ValidArgsFunction = completeJobName
	_ = cmd.RegisterFlagCompletionFunc("file", completeFile)
	return cmd
}
