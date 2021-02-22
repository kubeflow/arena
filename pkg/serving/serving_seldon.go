package serving

import (
	"encoding/json"
	"fmt"

"github.com/kubeflow/arena/pkg/apis/config"
"github.com/kubeflow/arena/pkg/apis/types"
"github.com/kubeflow/arena/pkg/util"
"github.com/kubeflow/arena/pkg/workflow"
log "github.com/sirupsen/logrus"
)

// SeldonServingProcesser use the default processer
type SeldonServingProcesser struct {
	*processer
}

func NewSeldonServingProcesser() Processer {
	p := &processer{
		processerType:   types.SeldonServingJob,
		client:          config.GetArenaConfiger().GetClientSet(),
		enable:          true,
		useIstioGateway: true,
	}
	return &SeldonServingProcesser{
		processer: p,
	}
}

func SubmitSeldonServingJob(namespace string, args *types.SeldonServingArgs) (err error) {
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
	// if job has been existed,skip to create it and return an error
	if len(jobs) != 0 {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena serve delete %s -n seldon'", args.Name, args.Name)
	}
	// the master is also considered as a worker
	chart := util.GetChartsFolder() + "/seldon-core"
	log.Infof("seldon chart path: %s", chart)
	temp, _ := json.Marshal(args)
	log.Infof("seldon args: %s", string(temp))
	err = workflow.SubmitJob(nameWithVersion, string(types.SeldonServingJob), namespace, args, chart)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", args.Name, args.Type)
	return nil
}
