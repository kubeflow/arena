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

package training

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/pkg/util"
)

func PruneTrainingJobs(namespace string, allNamespaces bool, since time.Duration) error {
	jobs := []TrainingJob{}
	trainers := GetAllTrainers()
	for _, trainer := range trainers {
		if !trainer.IsEnabled() {
			continue
		}
		trainingJobs, err := trainer.ListTrainingJobs(namespace, allNamespaces)
		if err != nil {
			log.Debugf("failed to list jobs of tainer %v,reason: %v", trainer.Type(), err)
			continue
		}
		jobs = append(jobs, trainingJobs...)
	}
	deleted := false
	if since == -1 {
		return fmt.Errorf("Your need to specify the relative duration live time of the job that need to be cleaned by --since. Like --since 10h")
	}
	for _, job := range jobs {
		if GetJobRealStatus(job) == "RUNNING" {
			continue
		}
		if job.Age() < since {
			continue
		}
		deleted = true
		fmt.Printf("Delete %s %s with Age %s \n", job.Trainer(), job.Name(), util.ShortHumanDuration(job.Age()))
		err := DeleteTrainingJob(job.Name(), job.Namespace(), job.Trainer())
		if err != nil {
			fmt.Printf("Failed to delete %s %s, err: %++v", job.Trainer(), job.Name(), err)
		}
	}
	if !deleted {
		fmt.Println("No job need to be deleted")
	}
	return nil
}
