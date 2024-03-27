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
