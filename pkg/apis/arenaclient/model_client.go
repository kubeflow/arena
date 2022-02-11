package arenaclient

import (
	"github.com/kubeflow/arena/pkg/apis/config"
	apismodel "github.com/kubeflow/arena/pkg/apis/model"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/model"
)

type ModelClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

// NewModelClient creates a ModelClient
func NewModelClient(namespace string, configer *config.ArenaConfiger) *ModelClient {
	return &ModelClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Namespace sets the namespace,this operation does not change the default namespace
func (m *ModelClient) Namespace(namespace string) *ModelClient {
	copyModelJobClient := &ModelClient{
		namespace: namespace,
		configer:  m.configer,
	}
	return copyModelJobClient
}

func (m *ModelClient) Submit(job *apismodel.Job) error {
	switch job.Type() {
	case types.ModelProfileJob:
		args := job.Args().(*types.ModelProfileArgs)
		return model.SubmitModelProfileJob(args.Namespace, args)
	case types.ModelOptimizeJob:
		args := job.Args().(*types.ModelOptimizeArgs)
		return model.SubmitModelOptimizeJob(args.Namespace, args)
	case types.ModelBenchmarkJob:
		args := job.Args().(*types.ModelBenchmarkArgs)
		return model.SubmitModelBenchmarkJob(args.Namespace, args)
	case types.ModelEvaluateJob:
		args := job.Args().(*types.ModelEvaluateArgs)
		return model.SubmitModelEvaluateJob(args.Namespace, args)
	}
	return nil
}

func (m *ModelClient) Get(jobType types.ModelJobType, name string) (*types.ModelJobInfo, error) {
	job, err := model.SearchModelJob(m.namespace, name, jobType)
	if err != nil {
		return nil, err
	}

	jobInfo := job.Convert2JobInfo()
	return &jobInfo, nil
}

func (m *ModelClient) GetAndPrint(jobType types.ModelJobType, name string, format string) error {
	job, err := model.SearchModelJob(m.namespace, name, jobType)
	if err != nil {
		return err
	}

	model.PrintModelJob(job, utils.TransferPrintFormat(format))
	return nil
}

func (m *ModelClient) List(allNamespaces bool, jobType types.ModelJobType) ([]*types.ModelJobInfo, error) {
	jobs, err := model.ListModelJobs(m.namespace, allNamespaces, jobType)
	if err != nil {
		return nil, err
	}

	var jobInfos []*types.ModelJobInfo
	for _, job := range jobs {
		jobInfo := job.Convert2JobInfo()
		jobInfos = append(jobInfos, &jobInfo)
	}
	return jobInfos, nil
}

func (m *ModelClient) ListAndPrint(allNamespaces bool, jobType types.ModelJobType, format string) error {
	jobs, err := model.ListModelJobs(m.namespace, allNamespaces, jobType)
	if err != nil {
		return err
	}

	model.PrintAllModelJobs(jobs, allNamespaces, utils.TransferPrintFormat(format))
	return nil
}

func (m *ModelClient) Delete(jobType types.ModelJobType, jobNames ...string) error {
	for _, jobName := range jobNames {
		err := model.DeleteModelJob(m.namespace, jobName, jobType)
		if err != nil {
			return err
		}
	}
	return nil
}
