// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
)

func NewModelGetCommand() *cobra.Command {
	var name, version string

	var command = &cobra.Command{
		Use:   "get",
		Short: "Get a registered model or model version",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_ = viper.BindPFlags(cmd.Flags())
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
				Kubeconfig:     viper.GetString("config"),
				LogLevel:       viper.GetString("loglevel"),
				Namespace:      viper.GetString("namespace"),
				ArenaNamespace: viper.GetString("arena-namespace"),
				IsDaemonMode:   false,
			})
			if err != nil {
				return fmt.Errorf("failed to create arena client: %v", err)
			}

			modelClient, err := client.Model()
			if err != nil {
				return fmt.Errorf("failed to create arena model client: %v", err)
			}

			if version == "" {
				// Get a registered model
				registeredModel, err := modelClient.GetRegisteredModel(name)
				if err != nil {
					return err
				}
				filter := fmt.Sprintf("name='%s'", registeredModel.Name)
				orderBy := []string{
					"name",
					"version_number",
				}
				modelVersions, err := modelClient.SearchModelVersions(filter, 100, orderBy)
				if err != nil {
					return err
				}
				utils.PrintRegisteredModel(registeredModel)
				utils.PrintModelVersions(modelVersions)
				return nil
			} else {
				// Get a model version
				modelVersion, err := modelClient.GetModelVersion(name, version)
				if err != nil {
					return err
				}
				utils.PrintModelVersion(modelVersion)
			}

			return nil
		},
	}

	command.Flags().StringVar(&name, "name", "", "model name")
	_ = command.MarkFlagRequired("name")
	command.Flags().StringVar(&version, "version", "", "model version")

	return command
}
