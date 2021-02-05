package arenaclient

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	apiserving "github.com/kubeflow/arena/pkg/apis/serving"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/podexec"
	"github.com/kubeflow/arena/pkg/serving"
)

// ServingJobClient provides some operators for managing serving jobs.
type ServingJobClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

// NewServingJobClient creates a ServingJobClient
func NewServingJobClient(namespace string, configer *config.ArenaConfiger) *ServingJobClient {
	return &ServingJobClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Namespace sets the namespace,this operation does not change the default namespace
func (t *ServingJobClient) Namespace(namespace string) *ServingJobClient {
	copyServingJobClient := &ServingJobClient{
		namespace: namespace,
		configer:  t.configer,
	}
	return copyServingJobClient
}

// Submit submits a training job
func (t *ServingJobClient) Submit(job *apiserving.Job) error {
	switch job.Type() {
	case types.TFServingJob:
		args := job.Args().(*types.TensorFlowServingArgs)
		return serving.SubmitTensorflowServingJob(args.Namespace, args)
	case types.TRTServingJob:
		args := job.Args().(*types.TensorRTServingArgs)
		return serving.SubmitTensorRTServingJob(args.Namespace, args)
	case types.CustomServingJob:
		args := job.Args().(*types.CustomServingArgs)
		return serving.SubmitCustomServingJob(args.Namespace, args)
	case types.KFServingJob:
		args := job.Args().(*types.KFServingArgs)
		return serving.SubmitKFServingJob(args.Namespace, args)
	}
	return nil
}

// Get returns a training job information
func (t *ServingJobClient) Get(jobName, version string, jobType types.ServingJobType) (*types.ServingJobInfo, error) {
	job, err := serving.SearchServingJob(t.namespace, jobName, version, jobType)
	if err != nil {
		return nil, err
	}
	jobInfo := job.Convert2JobInfo()
	return &jobInfo, nil
}

// GetAndPrint print training job information
func (t *ServingJobClient) GetAndPrint(jobName, version string, jobType types.ServingJobType, format string) error {
	if utils.TransferPrintFormat(format) == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}
	job, err := serving.SearchServingJob(t.namespace, jobName, version, jobType)
	if err != nil {
		return err
	}
	serving.PrintServingJob(job, utils.TransferPrintFormat(format))
	return nil
}

// List returns all training jobs
func (t *ServingJobClient) List(allNamespaces bool, servingType types.ServingJobType) ([]*types.ServingJobInfo, error) {
	jobs, err := serving.ListServingJobs(t.namespace, allNamespaces, servingType)
	if err != nil {
		return nil, err
	}
	jobInfos := []*types.ServingJobInfo{}
	for _, job := range jobs {
		jobInfo := job.Convert2JobInfo()
		jobInfos = append(jobInfos, &jobInfo)
	}
	return jobInfos, nil
}

// ListAndPrint lists and prints the job informations
func (t *ServingJobClient) ListAndPrint(allNamespaces bool, servingType types.ServingJobType, format string) error {
	if utils.TransferPrintFormat(format) == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}
	jobs, err := serving.ListServingJobs(t.namespace, allNamespaces, servingType)
	if err != nil {
		return err
	}
	serving.DisplayAllServingJobs(jobs, allNamespaces, utils.TransferPrintFormat(format))
	return nil
}

// Logs returns the training job log
func (t *ServingJobClient) Logs(jobName, version string, jobType types.ServingJobType, args *types.LogArgs) error {
	args.Namespace = t.namespace
	args.JobName = jobName
	return serving.AcceptJobLog(jobName, version, jobType, args)
}

func (t *ServingJobClient) Attach(jobName, version string, jobType types.ServingJobType, args *podexec.AttachPodArgs) error {
	job, err := t.Get(jobName, version, jobType)
	if err != nil {
		return err
	}
	if len(job.Instances) == 0 {
		return fmt.Errorf("can not attach the training job %v, because it has no instances", job.Name)
	}
	if args.Options.PodName == "" {
		if len(job.Instances) > 1 {
			return fmt.Errorf("%v", moreThanOneInstanceHelpInfo(job.Instances))
		}
		if len(job.Instances) == 0 {
			return fmt.Errorf("not found instances of serving job %v", job.Name)
		}
		args.Options.PodName = job.Instances[0].Name
	}
	command := []string{job.Name}
	command = append(command, args.Command...)
	if err := args.Options.Complete(command, t.namespace, args.CmdArgsLenAtDash); err != nil {
		return err
	}
	if err := args.Options.Validate(); err != nil {
		return err
	}
	return args.Options.Run()
}

// Delete deletes the target training job
func (t *ServingJobClient) Delete(jobType types.ServingJobType, version string, jobNames ...string) error {
	for _, jobName := range jobNames {
		err := serving.DeleteServingJob(t.namespace, jobName, version, jobType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ServingJobClient) TrafficRouterSplit(args *types.TrafficRouterSplitArgs) error {
	return serving.RunTrafficRouterSplit(args.Namespace, args)
}

func moreThanOneInstanceHelpInfo(instances []types.ServingInstance) string {
	header := fmt.Sprintf("There is %d instances have been found:", len(instances))
	lines := []string{}
	footer := fmt.Sprintf("please use '-i' or '--instance' to filter.")
	for _, i := range instances {
		lines = append(lines, fmt.Sprintf("%v", i.Name))
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(lines, "\n"), footer)

}
