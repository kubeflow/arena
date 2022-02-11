package model

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitModelEvaluateJob(namespace string, args *types.ModelEvaluateArgs) error {
	args.Namespace = namespace

	if args.Command == "" {
		args.Command = fmt.Sprintf("python easy_inference/main.py evaluate --model-path=%s --model-platform=%s --dataset-path=%s "+
			"--batch-size=%d --report-path=%s", args.ModelPath, args.ModelPlatform, args.DatasetPath, args.BatchSize, args.ReportPath)
	}

	modelJobChart := util.GetChartsFolder() + "/modeljob"
	err := workflow.SubmitJob(args.Name, string(types.ModelEvaluateJob), namespace, args, modelJobChart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The model evaluate job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena model get %s` to check the job status", args.Name)
	return nil
}
