package flags

import (
	"github.com/kubeflow/arena/cmd/arena/commands/util"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamespaceToUseFromProjectFlag(cmd *cobra.Command, kubeClient *client.Client) (string, error) {
	flagValue := getFlagValue(cmd, ProjectFlag)
	if flagValue != "" {
		return util.GetNamespaceFromProjectName(flagValue, kubeClient)
	}

	return kubeClient.GetDefaultNamespace(), nil
}

func GetNamespaceToUseFromProjectFlagIncludingAll(cmd *cobra.Command, kubeClient *client.Client, allFlag bool) (string, error) {
	if allFlag {
		return metav1.NamespaceAll, nil
	} else {
		return GetNamespaceToUseFromProjectFlag(cmd, kubeClient)
	}
}

func getFlagValue(cmd *cobra.Command, name string) string {
	return cmd.Flags().Lookup(name).Value.String()
}
