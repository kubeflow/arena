package analyze

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitModelBenchmarkJob(namespace string, args *types.ModelBenchmarkArgs) error {
	args.Namespace = namespace

	if args.Command == "" {
		args.Command = fmt.Sprintf("python easy_inference/main.py benchmark --model-config-file=%s "+
			"--report-path=%s --concurrency=%d --requests=%d --duration=%d",
			args.ModelConfigFile, args.ReportPath, args.Concurrency, args.Requests, args.Duration)
	}

	modelJobChart := util.GetChartsFolder() + "/modeljob"
	err := workflow.SubmitJob(args.Name, string(types.ModelBenchmarkJob), namespace, args, modelJobChart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The model benchmark job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena model analyze get %s` to check the job status", args.Name)
	return nil
}
