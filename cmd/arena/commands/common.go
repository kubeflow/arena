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
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Global variables
var (
	// To reduce client-go API call, for 'arena list' scenario
	allPods        []v1.Pod
	allJobs        []batchv1.Job
	useCache       bool
	name           string
	namespace      string
	arenaNamespace string // the system namespace of arena
)

func getAllTrainingTypes(clientset *kubernetes.Clientset) []string {
	trainers := NewTrainers(clientset)
	trainerNames := []string{}
	for _, trainer := range trainers {
		trainerNames = append(trainerNames, trainer.Type())
	}

	return trainerNames
}
