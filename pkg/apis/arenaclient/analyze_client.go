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

package arenaclient

import (
	"github.com/kubeflow/arena/pkg/apis/config"
	apisanalyze "github.com/kubeflow/arena/pkg/apis/model/analyze"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/model/analyze"
)

type AnalyzeClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

func NewAnalyzeClient(namespace string, configer *config.ArenaConfiger) *AnalyzeClient {
	return &AnalyzeClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Namespace sets the namespace,this operation does not change the default namespace
func (m *AnalyzeClient) Namespace(namespace string) *AnalyzeClient {
	copyModelJobClient := &AnalyzeClient{
		namespace: namespace,
		configer:  m.configer,
	}
	return copyModelJobClient
}

func (m *AnalyzeClient) Submit(job *apisanalyze.Job) error {
	switch job.Type() {
	case types.ModelProfileJob:
		args := job.Args().(*types.ModelProfileArgs)
		return analyze.SubmitModelProfileJob(args.Namespace, args)
	case types.ModelOptimizeJob:
		args := job.Args().(*types.ModelOptimizeArgs)
		return analyze.SubmitModelOptimizeJob(args.Namespace, args)
	case types.ModelBenchmarkJob:
		args := job.Args().(*types.ModelBenchmarkArgs)
		return analyze.SubmitModelBenchmarkJob(args.Namespace, args)
	case types.ModelEvaluateJob:
		args := job.Args().(*types.ModelEvaluateArgs)
		return analyze.SubmitModelEvaluateJob(args.Namespace, args)
	}
	return nil
}

func (m *AnalyzeClient) Get(jobType types.ModelJobType, name string) (*types.ModelJobInfo, error) {
	job, err := analyze.SearchModelJob(m.namespace, name, jobType)
	if err != nil {
		return nil, err
	}

	jobInfo := job.Convert2JobInfo()
	return &jobInfo, nil
}

func (m *AnalyzeClient) GetAndPrint(jobType types.ModelJobType, name string, format string) error {
	job, err := analyze.SearchModelJob(m.namespace, name, jobType)
	if err != nil {
		return err
	}

	analyze.PrintModelJob(job, utils.TransferPrintFormat(format))
	return nil
}

func (m *AnalyzeClient) List(allNamespaces bool, jobType types.ModelJobType) ([]*types.ModelJobInfo, error) {
	jobs, err := analyze.ListModelJobs(m.namespace, allNamespaces, jobType)
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

func (m *AnalyzeClient) ListAndPrint(allNamespaces bool, jobType types.ModelJobType, format string) error {
	jobs, err := analyze.ListModelJobs(m.namespace, allNamespaces, jobType)
	if err != nil {
		return err
	}

	analyze.PrintAllModelJobs(jobs, allNamespaces, utils.TransferPrintFormat(format))
	return nil
}

func (m *AnalyzeClient) Delete(jobType types.ModelJobType, jobNames ...string) error {
	for _, jobName := range jobNames {
		err := analyze.DeleteModelJob(m.namespace, jobName, jobType)
		if err != nil {
			return err
		}
	}
	return nil
}
