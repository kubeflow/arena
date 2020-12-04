package serving

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/podlogs"
)

func AcceptJobLog(name, version string, jobType types.ServingJobType, args *types.LogArgs) error {
	namespace := args.Namespace
	job, err := SearchServingJob(namespace, name, version, jobType)
	if err != nil {
		return err
	}
	jobInfo := job.Convert2JobInfo()
	// 1.if not found instances,return an error
	if len(jobInfo.Instances) == 0 {
		return fmt.Errorf("not found instances of serving job,please use 'arena serve get %v' to get job information", name)
	}
	// 2.if instance name is null and job has more than one instance,return an error
	// push user to slelect one
	if len(jobInfo.Instances) > 1 && args.InstanceName == "" {
		return fmt.Errorf("%v", moreThanOneInstanceHelpInfo(jobInfo.Instances))
	}
	// 3.if user not specifiy the instance name and the job has only one instance name,pick the instance
	if args.InstanceName == "" {
		args.InstanceName = jobInfo.Instances[0].Name
	}
	// 4.if the instance name is invalid,return an error
	exists := map[string]bool{}
	for _, i := range jobInfo.Instances {
		exists[i.Name] = true
	}
	if _, ok := exists[args.InstanceName]; !ok {
		return fmt.Errorf("invalid instance name %v of serving job %v,please use 'arena serve get %v' to get instance names.", args.InstanceName, name, name)
	}
	logger := podlogs.NewPodLogger(args)
	_, err = logger.AcceptLogs()
	return err
}
