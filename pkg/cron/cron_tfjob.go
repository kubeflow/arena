package cron

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitCronTFJob(namespace string, submitArgs *types.CronTFJobArgs) (err error) {
	b, err := json.Marshal(submitArgs)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	cron_tfjob_chart := util.GetChartsFolder() + "/cron-tfjob"
	// the master is also considered as a worker
	// submitArgs.WorkerCount = submitArgs.WorkerCount - 1

	//if submitArgs.TFRuntime != nil {
	//	cron_tfjob_chart = util.GetChartsFolder() + "/" + submitArgs.TFRuntime.GetChartName()
	//}
	err = workflow.SubmitJob(submitArgs.Name, string(types.CronTFTrainingJob), namespace, submitArgs, cron_tfjob_chart, submitArgs.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", submitArgs.Name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", submitArgs.Name, submitArgs.TrainingType)

	return nil
}
