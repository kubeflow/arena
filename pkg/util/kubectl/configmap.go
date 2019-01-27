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

package kubectl

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const JOB_CONFIG_LABEL = "createdBy=arena"

/**
*
* list configMaps by using namespace
**/
func ListAppConfigMaps(clientset *kubernetes.Clientset, namespace string, trainingTypes []string) (jobs []types.TrainingJobInfo, err error) {

	jobs = []types.TrainingJobInfo{}
	cmList, err := clientset.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{LabelSelector: JOB_CONFIG_LABEL})
	if err != nil {
		return nil, err
	}

	log.Debugf("The cmList is %v", cmList.Items)

	for _, cm := range cmList.Items {
		found := false

		job := types.TrainingJobInfo{}

	innerLoop:
		for _, trainingType := range trainingTypes {
			if strings.HasSuffix(cm.Name, fmt.Sprintf("-%s", trainingType)) {
				found = true
				job.Name = strings.TrimSuffix(cm.Name, fmt.Sprintf("-%s", trainingType))
				job.Type = trainingType
				job.Namespace = cm.Namespace
				break innerLoop
			}
		}

		if found {
			jobs = append(jobs, job)
		} else {
			log.Debugf("drop %s in training configmap", job)
		}

	}

	log.Debugf("the job training configmap: %q", jobs)

	return jobs, nil
}
