package training

import (
	"github.com/kubeflow/arena/pkg/apis/types"
)

type Job interface {
	Name() string
	Type() types.TrainingJobType
	Args() interface{}
}

type job struct {
	name    string
	jobType types.TrainingJobType
	args    interface{}
}

func NewJob(name string, jobType types.TrainingJobType, args interface{}) Job {
	return &job{
		name:    name,
		jobType: jobType,
		args:    args,
	}

}

func (j *job) Name() string {
	return j.name
}

func (j *job) Type() types.TrainingJobType {
	return j.jobType
}

func (j *job) Args() interface{} {
	return j.args
}
