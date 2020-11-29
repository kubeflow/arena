package training

import (
	"github.com/kubeflow/arena/pkg/apis/types"
)

type baseJob struct {
	name    string
	jobType string
	args    interface{}
}

func newBaseJob(name string, jobType string, args interface{}) baseJob {
	return baseJob{
		name:    name,
		jobType: jobType,
		args:    args,
	}
}

func (b *baseJob) Name() string {
	return b.name
}

func (b *baseJob) Type() string {
	return b.jobType
}

func (b *baseJob) Args() interface{} {
	return b.args
}

// Job defines the base job
type Job struct {
	baseJob
}

func NewJob(name string, jobType types.TrainingJobType, args interface{}) *Job {
	return &Job{
		baseJob: newBaseJob(name, string(jobType), args),
	}
}

func (j *Job) Type() types.TrainingJobType {
	return types.TrainingJobType(j.baseJob.Type())
}
