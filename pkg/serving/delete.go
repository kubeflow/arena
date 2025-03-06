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

package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func DeleteServingJob(namespace, name, version string, jobType types.ServingJobType) error {
	job, err := SearchServingJob(namespace, name, version, jobType)
	if err != nil {
		if strings.Contains(err.Error(), "Not found serving job") {
			log.Infof("The serving job '%v' does not exist,skip to delete it.", name)
			return nil
		}
		return err
	}
	nameWithVersion := fmt.Sprintf("%v-%v", job.Name(), job.Version())
	if job.Type() == types.KServeJob {
		nameWithVersion = job.Name()
	}
	servingType := string(job.Type())
	err = workflow.DeleteJob(nameWithVersion, namespace, servingType)
	if err != nil {
		return err
	}
	log.Infof("The serving job %s with version %s has been deleted successfully", job.Name(), job.Version())
	return nil
}
