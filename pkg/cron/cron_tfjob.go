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

package cron

import (
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitCronTFJob(namespace string, submitArgs *types.CronTFJobArgs) (err error) {
	cronTFJobChart := util.GetChartsFolder() + "/cron-tfjob"

	err = workflow.SubmitJob(submitArgs.Name, string(types.CronTFTrainingJob), namespace, submitArgs, cronTFJobChart, submitArgs.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The cron tfjob %s has been submitted successfully", submitArgs.Name)
	log.Infof("You can run `arena cron get %s` to check the cron status", submitArgs.Name)

	return nil
}
