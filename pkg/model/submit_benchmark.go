package model

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitModelBenchmarkJob(namespace string, args *types.ModelBenchmarkArgs) error {
	args.Namespace = namespace

	b, _ := json.Marshal(args)
	log.Debugf("args: %s", string(b))

	if args.Command == "" {
		if args.ModelConfigFile != "" {
			args.Command = fmt.Sprintf("python easy_inference/main.py benchmark --model-config-file=%s "+
				"--report-path=%s --concurrency=%d --requests=%d --duration=%d",
				args.ModelConfigFile, args.ReportPath, args.Concurrency, args.Requests, args.Duration)
		} else {
			args.Command = fmt.Sprintf("python easy_inference/main.py benchmark --model-name=%s --model-path=%s "+
				"--inputs=%s --outputs=%s --report-path=%s --concurrency=%d --requests=%d --duration=%d",
				args.ModelName, args.ModelPath, args.Inputs, args.Outputs, args.ReportPath, args.Concurrency, args.Requests, args.Duration)
		}
	}

	modelJobChart := util.GetChartsFolder() + "/modeljob"
	err := workflow.SubmitJob(args.Name, string(types.ModelBenchmarkJob), namespace, args, modelJobChart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The model benchmark job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena model get %s` to check the job status", args.Name)
	return nil
}
