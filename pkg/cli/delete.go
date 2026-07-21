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

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a training job",
	Long:  `Delete a training job by name or YAML file (similar to kubectl delete -f).`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string

		var yamlNS string
		if deleteFile != "" {
			// Load from file
			t, err := task.LoadFromFile(deleteFile)
			if err != nil {
				return fmt.Errorf("failed to load file %q: %w", deleteFile, err)
			}
			name = t.Name
			if name == "" {
				return fmt.Errorf("file %q does not specify a job name", deleteFile)
			}
			yamlNS = t.Namespace
		} else if len(args) > 0 {
			name = args[0]
		} else {
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
		fmt.Printf("%s/%s deleted\n", strings.ToLower(jobType), name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteFile, "file", "f", "", "path to YAML file")
	jobCmd.AddCommand(deleteCmd)
}
