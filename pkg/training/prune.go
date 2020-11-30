package training

import (
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
)

type PruneArgs struct {
	since time.Duration
}

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
