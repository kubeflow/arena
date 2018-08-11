package commands

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// CLIName is the name of the CLI
	CLIName = "arena"
)

var (
	loadingRules *clientcmd.ClientConfigLoadingRules
	logLevel     string
	enablePProf  bool
)

// NewCommand returns a new instance of an Arena command
func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   CLIName,
		Short: "arena is the command line interface to Arena",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	// command.AddCommand(NewCompletionCommand())
	// 1. unzip chart
	// command.AddCommand(NewInstallCommand())
	//command.AddCommand(NewListCommand())
	addKubectlFlagsToCmd(command)

	// enable logging
	command.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")
	command.PersistentFlags().BoolVar(&enablePProf, "pprof", false, "enable cpu profile")
	command.PersistentFlags().StringVar(&arenaNamespace, "arenaNamespace", "arena-system", "The namespace of arena system service, like TFJob")

	command.AddCommand(NewSubmitCommand())
	command.AddCommand(NewSubmitServingCommand())
	command.AddCommand(NewListCommand())
	command.AddCommand(NewGetCommand())
	command.AddCommand(NewLogViewerCommand())
	command.AddCommand(NewLogsCommand())
	command.AddCommand(NewDeleteCommand())
	command.AddCommand(NewTopCommand())
	command.AddCommand(NewVersionCmd(CLIName))
	// command.AddCommand(NewWaitCommand())
	// command.AddCommand(cmd.NewVersionCmd(CLIName))

	return command
}

func addKubectlFlagsToCmd(cmd *cobra.Command) {
	// The "usual" clientcmd/kubectl flags
	loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	// kflags := clientcmd.RecommendedConfigOverrideFlags("")
	cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "config", "", "Path to a kube config. Only required if out-of-cluster")
	cmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "namespace")
	// clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}

func createNamespace(client *kubernetes.Clientset, namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.Core().Namespaces().Create(ns)
	return err
}

func getNamespace(client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return client.Core().Namespaces().Get(namespace, metav1.GetOptions{})
}

func ensureNamespace(client *kubernetes.Clientset, namespace string) error {
	_, err := getNamespace(client, namespace)
	if err != nil && errors.IsNotFound(err) {
		return createNamespace(client, namespace)
	}
	return err
}
