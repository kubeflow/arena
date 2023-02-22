package evaluate

import (
	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/workflow"
)

func DeleteEvaluateJob(name, namespace string) error {
	log.Infof("delete evaluate job, %s-%s", name, namespace)
	return workflow.DeleteJob(name, namespace, string(types.EvaluateJob))
}
