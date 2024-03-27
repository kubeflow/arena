package analyze

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitModelProfileJob(namespace string, args *types.ModelProfileArgs) error {
	args.Namespace = namespace

	if args.Command == "" {
		args.Command = fmt.Sprintf("python easy_inference/main.py profile --model-config-file=%s --report-path=%s",
			args.ModelConfigFile, args.ReportPath)
	}

	modelJobChart := util.GetChartsFolder() + "/modeljob"
	err := workflow.SubmitJob(args.Name, string(types.ModelProfileJob), namespace, args, modelJobChart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The model profile job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena model analyze get %s` to check the job status", args.Name)
	return nil
}
