// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analyze

import "github.com/kubeflow/arena/pkg/apis/types"

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

func NewJob(name string, jobType types.ModelJobType, args interface{}) *Job {
	return &Job{
		baseJob: newBaseJob(name, string(jobType), args),
	}
}

func (j *Job) Type() types.ModelJobType {
	return types.ModelJobType(j.baseJob.Type())
}
