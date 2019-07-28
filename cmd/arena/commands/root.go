// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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
	enableTrace  bool
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
	command.PersistentFlags().BoolVar(&enableTrace, "trace", false, "enable trace")
	command.PersistentFlags().StringVar(&arenaNamespace, "arenaNamespace", "arena-system", "The namespace of arena system service, like tf-operator")
	command.PersistentFlags().MarkDeprecated("arenaNamespace", "please use --arena-namespace")
	command.PersistentFlags().StringVar(&arenaNamespace, "arena-namespace", "arena-system", "The namespace of arena system service, like tf-operator")

	command.AddCommand(NewSubmitCommand())
	command.AddCommand(NewServeCommand())
	command.AddCommand(NewListCommand())
	command.AddCommand(NewPruneCommand())
	command.AddCommand(NewGetCommand())
	command.AddCommand(NewLogViewerCommand())
	command.AddCommand(NewLogsCommand())
	command.AddCommand(NewDeleteCommand())
	command.AddCommand(NewTopCommand())
	command.AddCommand(NewVersionCmd(CLIName))
	command.AddCommand(NewDataCommand())
	command.AddCommand(NewCompletionCommand())
	command.AddCommand(NewDiagnoseCommand())
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
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "the namespace of the job")
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
