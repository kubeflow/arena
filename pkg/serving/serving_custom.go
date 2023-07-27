package serving

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

// CustomServingProcesser use the default processer
type CustomServingProcesser struct {
	*processer
}

func NewCustomServingProcesser() Processer {
	p := &processer{
		processerType:   types.CustomServingJob,
		client:          config.GetArenaConfiger().GetClientSet(),
		enable:          true,
		useIstioGateway: false,
	}
	return &CustomServingProcesser{
		processer: p,
	}
}

func SubmitCustomServingJob(namespace string, args *types.CustomServingArgs) (err error) {
	nameWithVersion := fmt.Sprintf("%v-%v", args.Name, args.Version)
	args.Namespace = namespace
	processers := GetAllProcesser()
	processer, ok := processers[args.Type]
	if !ok {
		return fmt.Errorf("not found processer whose type is %v", args.Type)
	}
	jobs, err := processer.GetServingJobs(args.Namespace, args.Name, args.Version)
	if err != nil {
		return err
	}
	if err := ValidateJobsBeforeSubmiting(jobs, args.Name); err != nil {
		return err
	}
	// the master is also considered as a worker
	customChart := util.GetChartsFolder() + "/custom-serving"
	err = workflow.SubmitJob(nameWithVersion, string(types.CustomServingJob), namespace, args, customChart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena serve get %s --type %s` to check the job status", args.Name, args.Type)
	return nil
}
