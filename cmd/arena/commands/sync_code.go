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
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type submitSyncCodeArgs struct {
	SyncMode   string `yaml:"syncMode"`            // --syncMode: rsync, hdfs, git
	SyncSource string `yaml:"syncSource"`          // --syncSource
	SyncImage  string `yaml:"syncImage,omitempty"` // --syncImage
	// syncGitProjectName
	SyncGitProjectName string `yaml:"syncGitProjectName,omitempty"` // --syncImage
}

func (sc *submitSyncCodeArgs) HandleSyncCode() error {

	switch sc.SyncMode {
	case "":
		log.Debugf("No action for sync Code")
	case "git":
		log.Debugf("Check and prepare sync code with git")
		if sc.SyncSource == "" {
			return fmt.Errorf("--syncSource should be set when syncMode is set")
		}

		// split test.git to test

		parts := strings.Split(strings.Trim(sc.SyncSource, "/"), "/")
		sc.SyncGitProjectName = strings.Split(parts[len(parts)-1], ".git")[0]
		log.Debugf("Try to split %s to get project name %s", sc.SyncSource, sc.SyncGitProjectName)
	case "rsync":
		log.Debugf("Check and prepare sync code with rsync")
		if sc.SyncSource == "" {
			return fmt.Errorf("--syncSource should be set when syncMode is set")
		}

	default:
		log.Fatalf("Unknown sync mode: %s", sc.SyncMode)
		return fmt.Errorf("Unknown sync mode: %s, it should be git or rsync", sc.SyncMode)
	}

	return nil
}

func (sc *submitSyncCodeArgs) addSyncFlags(command *cobra.Command) {
	command.Flags().StringVar(&sc.SyncMode, "syncMode", "", "syncMode: support rsync, hdfs, git")
	// command.MarkFlagRequired("syncMode")
	command.Flags().StringVar(&sc.SyncSource, "syncSource", "", "syncSource: for rsync, it's like 10.88.29.56::backup/data/logoRecoTrain.zip; for git, it's like https://github.com/kubeflow/tf-operator.git")
	command.Flags().StringVar(&sc.SyncImage, "syncImage", "", "the docker image of syncImage")
}
