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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Global variables
var (
	restConfig   *rest.Config
	clientConfig clientcmd.ClientConfig
	clientset    *kubernetes.Clientset
	// To reduce client-go API call, for 'arena list' scenario
	allPods        []v1.Pod
	allJobs        []batchv1.Job
	useCache       bool
	name           string
	namespace      string
	arenaNamespace string // the system namespace of arena
)

func initKubeClient() (*kubernetes.Clientset, error) {
	loadArenaConifg()
	if clientset != nil {
		return clientset, nil
	}
	var err error
	restConfig, err = clientConfig.ClientConfig()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return clientset, nil
}

// Update namespace if it's not set by user
func updateNamespace(cmd *cobra.Command) (err error) {
	found := false
	// Update the namespace
	if !cmd.Flags().Changed("namespace") {
		namespace, found = arenaConfigs["namespace"]
		if !found {
			namespace, _, err = clientConfig.Namespace()
			log.Debugf("auto detect namespace %s", namespace)
			if err != nil {
				return err
			}
		} else {
			log.Debugf("use arena config namespace %s", namespace)
		}
	} else {
		log.Debugf("force to use namespace %s", namespace)
	}
	return nil
}

func getAllTrainingTypes(clientset *kubernetes.Clientset) []string {
	trainers := NewTrainers(clientset)
	trainerNames := []string{}
	for _, trainer := range trainers {
		trainerNames = append(trainerNames, trainer.Type())
	}

	return trainerNames
}
