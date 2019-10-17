package exec 

import (
	"github.com/pkg/errors"
	serving "github.com/kubeflow/arena/pkg/jobs/servingjob"
	"github.com/kubeflow/arena/pkg/podexec"
	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	restclient "k8s.io/client-go/rest"
)

type AcceptExecArgs struct {
	Option *podexec.ExecOptions
	Config *restclient.Config
	Namespace string
	ArgsIn []string
	ArgsLenAtDash int
	Version string
	Type string
}

func ServingJobExecCommand(args AcceptExecArgs) (error) {
	// check the serving job type
	err := types.CheckServingTypeIsOk(args.Type)
	if err != nil {
		return err
	}
	// get clientset from rest config
	clientset := util.CreateK8sClientSetWithConfig(args.Config)
	manager,err := serving.NewServingJobManager(clientset , args.Namespace)
	if err != nil {
		return err
	}
	// create filter args 
	filterArgs := serving.ServingJobFilterArgs{
		Namespace: args.Namespace,
		Type: args.Type,
		Version: args.Version, 
		Name: args.ArgsIn[0],
	}
	// get target jobs under the filter args
	jobs := manager.GetTargetServingJob(filterArgs)
	// if job is not found or jobs accout more than one,print error message 
	if len(jobs) == 0 || len(jobs) > 1{
		helpString,err := manager.GetHelpInfo(jobs)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, helpString)
		w.Flush()
		return nil
	}
	job := jobs[0]
	// get the target pod,if pod assigned by users is null and there is 
	// only one pod in job,return the pod name
	podName,err := job.GetTargetPod(args.Option.PodName)
	if err != nil {
		if err == types.ErrTooManyPods {
			info,_ := job.GetHelpInfo()
			fmt.Println(info)
			return nil
		}
		return err
	}
	// reset pod name
	args.Option.PodName = podName
	// complete the option
	if err := args.Options.Complete(args.Config, args.Namespace,args.ArgsIn, args.ArgsLenAtDash); err != nil {
		if err == types.ErrInvalidUsage {
			return fmt.Errorf(podexec.ExecUsageStr)
		}
		return errors.Wrap(err,"complete exec options failed,reason:")
	}
	// validate the option
	if err := args.Options.Validate(); err != nil {
		return errors.Wrap(err,"validate failed,reason: ")
	}
	// exec in container
	if err := args.Options.Run(); err != nil {
		return errors.Wrap(err,"exec in container failed,reason: ")
	} 
	return nil
}