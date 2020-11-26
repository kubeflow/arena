package arenaclient

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/config"
	apistraining "github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/training"
)

// TrainingJobClient provides some operators for managing training jobs.
type TrainingJobClient struct {
	// namespace store the namespace
	namespace            string
	arenaSystemNamespace string
	configer             *config.ArenaConfiger
}

// NewTrainingJobClient creates a TrainingJobClient
func NewTrainingJobClient(namespace, arenaSystemNamespace string, configer *config.ArenaConfiger) *TrainingJobClient {
	return &TrainingJobClient{
		namespace:            namespace,
		arenaSystemNamespace: arenaSystemNamespace,
		configer:             configer,
	}
}

// Namespace sets the namespace
func (t *TrainingJobClient) Namespace(namespace string) *TrainingJobClient {
	copyTrainingJobClient := &TrainingJobClient{
		namespace:            namespace,
		arenaSystemNamespace: t.arenaSystemNamespace,
		configer:             t.configer,
	}
	return copyTrainingJobClient
}

// Submit submits a training job
func (t *TrainingJobClient) Submit(job apistraining.Job) error {
	switch job.Type() {
	case types.TFTrainingJob:
		args := job.Args().(*types.SubmitTFJobArgs)
		return training.SubmitTFJob(t.namespace, args)
	case types.PytorchTrainingJob:
		args := job.Args().(*types.SubmitPyTorchJobArgs)
		return training.SubmitPytorchJob(t.namespace, args)
	}
	return nil
}

// Get returns a training job information
func (t *TrainingJobClient) Get(jobName string, jobType types.TrainingJobType) (*types.TrainingJobInfo, error) {
	job, err := training.SearchTrainingJob(jobName, t.namespace, jobType)
	if err != nil {
		return nil, err
	}
	jobInfo := training.BuildJobInfo(job)
	return jobInfo, nil
}

// GetAndPrint print training job information
func (t *TrainingJobClient) GetAndPrint(jobName string, jobType types.TrainingJobType, format string, showEvent bool) error {
	if utils.TransferPrintFormat(format) == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide,json,yaml]")
	}
	job, err := training.SearchTrainingJob(jobName, t.namespace, jobType)
	if err != nil {
		return err
	}
	training.PrintTrainingJob(job, format, showEvent)
	return nil
}

// List returns all training jobs
func (t *TrainingJobClient) List(allNamespaces bool) ([]*types.TrainingJobInfo, error) {
	jobs, err := training.ListTrainingJobs(t.namespace, allNamespaces)
	if err != nil {
		return nil, err
	}
	jobInfos := []*types.TrainingJobInfo{}
	for _, job := range jobs {
		jobInfos = append(jobInfos, training.BuildJobInfo(job))
	}
	return jobInfos, nil
}

func (t *TrainingJobClient) ListAndPrint(allNamespaces bool, format string) error {
	if utils.TransferPrintFormat(format) == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide,json,yaml]")
	}
	jobs, err := training.ListTrainingJobs(t.namespace, allNamespaces)
	if err != nil {
		return err
	}
	training.DisplayTrainingJobList(jobs, format)
	return nil
}

// Logs returns the training job log
func (t *TrainingJobClient) Logs(jobName string, jobType types.TrainingJobType, args *types.LogArgs) error {
	args.Namespace = t.namespace
	args.JobName = jobName
	fmt.Println(args)
	return training.AcceptJobLog(jobName, jobType, args)
}

// Delete deletes the target training job
func (t *TrainingJobClient) Delete(jobType types.TrainingJobType, jobNames ...string) error {
	for _, jobName := range jobNames {
		err := training.DeleteTrainingJob(jobName, t.namespace, jobType)
		if err != nil {
			return err
		}
	}
	return nil
}

// LogViewer returns the log viewer
func (t *TrainingJobClient) LogViewer(jobName string, jobType types.TrainingJobType) ([]string, error) {
	job, err := training.SearchTrainingJob(jobName, t.namespace, jobType)
	if err != nil {
		return nil, err
	}
	return job.GetJobDashboards(t.configer.GetClientSet(), t.namespace, t.arenaSystemNamespace)
}
