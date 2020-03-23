package flags

import (
	"github.com/kubeflow/arena/pkg/client"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetProjectFlag(cmd *cobra.Command, kubeClient *client.Client) string {
	flagValue := getFlagValue(cmd, ProjectFlag)
	if flagValue != "" {
		return flagValue
	}

	return kubeClient.GetDefaultNamespace()
}

func GetProjectFlagIncludingAll(cmd *cobra.Command, kubeClient *client.Client, allFlag bool) string {
	if allFlag {
		return metav1.NamespaceAll
	} else {
		return GetProjectFlag(cmd, kubeClient)
	}
}

func getFlagValue(cmd *cobra.Command, name string) string {
	return cmd.Flags().Lookup(name).Value.String()
}
