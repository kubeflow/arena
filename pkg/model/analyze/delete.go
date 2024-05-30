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

package analyze

import (
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func DeleteModelJob(namespace, name string, jobType types.ModelJobType) error {
	job, err := SearchModelJob(namespace, name, jobType)
	if err != nil {
		if strings.Contains(err.Error(), "Not found model job") {
			log.Infof("The model job '%v' doest not exist,skip to delete it.", name)
			return nil
		}
		return err
	}
	err = workflow.DeleteJob(name, namespace, string(job.Type()))
	if err != nil {
		return err
	}
	log.Infof("The model job %s has been deleted successfully", job.Name())
	return nil
}
