package evaluate

import (
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitEvaluateJob(namespace string, submitArgs *types.EvaluateJobArgs) (err error) {
	evaluateJobChart := util.GetChartsFolder() + "/evaluatejob"

	err = workflow.SubmitJob(submitArgs.Name, string(types.EvaluateJob), namespace, submitArgs, evaluateJobChart, submitArgs.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The evaluate job %s has been submitted successfully", submitArgs.Name)
	log.Infof("You can run `arena evaluate get %s` to check the evaluate job status", submitArgs.Name)

	return nil
}