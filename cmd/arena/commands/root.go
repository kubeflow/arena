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
	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/cmd/arena/commands/project"

	"github.com/kubeflow/arena/pkg/config"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	loadingRules *clientcmd.ClientConfigLoadingRules
	logLevel     string
)

// NewCommand returns a new instance of an Arena command
func NewCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   config.CLIName,
		Short: "runai is the command line interface to RunAI",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		// Would be run before any child command
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
		},
	}

	// command.AddCommand(NewCompletionCommand())
	// 1. unzip chart
	// command.AddCommand(NewInstallCommand())
	//command.AddCommand(NewListCommand())
	addKubectlFlagsToCmd(command)

	// enable logging
	command.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")

	command.AddCommand(NewRunaiJobCommand())
	// command.AddCommand(NewServeCommand())
	command.AddCommand(NewListCommand())
	// command.AddCommand(NewPruneCommand())
	command.AddCommand(NewGetCommand())
	// command.AddCommand(NewLogViewerCommand())
	command.AddCommand(NewLogsCommand())
	command.AddCommand(NewDeleteCommand())
	command.AddCommand(NewTopCommand())
	command.AddCommand(NewVersionCmd())
	// command.AddCommand(NewDataCommand())
	// command.AddCommand(NewCompletionCommand())
	command.AddCommand(NewUpdateCommand())
	command.AddCommand(NewBashCommand())
	command.AddCommand(NewExecCommand())
	command.AddCommand(NewTemplateCommand())
	command.AddCommand(project.NewProjectCommand())
	// command.AddCommand(NewWaitCommand())
	// command.AddCommand(cmd.NewVersionCmd(CLIName))

	return command
}

func addKubectlFlagsToCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(flags.ProjectFlag, "p", "", "Specifies the Run:AI project to use for this Job.")
}

func createNamespace(client *kubernetes.Clientset, namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ns)
	return err
}

func getNamespace(client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {
	return client.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
}

func ensureNamespace(client *kubernetes.Clientset, namespace string) error {
	_, err := getNamespace(client, namespace)
	if err != nil && errors.IsNotFound(err) {
		return createNamespace(client, namespace)
	}
	return err
}
