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
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	apistraining "github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/podexec"
	"github.com/kubeflow/arena/pkg/training"
)

var (
	errJobNotFoundMessage = "Not found training job %s in namespace %s,please use 'arena submit' to create it."
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
func (t *TrainingJobClient) Submit(job *apistraining.Job) error {
	switch job.Type() {
	case types.TFTrainingJob:
		args := job.Args().(*types.SubmitTFJobArgs)
		return training.SubmitTFJob(t.namespace, args)
	case types.PytorchTrainingJob:
		args := job.Args().(*types.SubmitPyTorchJobArgs)
		return training.SubmitPytorchJob(t.namespace, args)
	case types.MPITrainingJob:
		args := job.Args().(*types.SubmitMPIJobArgs)
		return training.SubmitMPIJob(t.namespace, args)
	case types.HorovodTrainingJob:
		args := job.Args().(*types.SubmitHorovodJobArgs)
		return training.SubmitHorovodJob(t.namespace, args)
	case types.VolcanoTrainingJob:
		args := job.Args().(*types.SubmitVolcanoJobArgs)
		return training.SubmitVolcanoJob(t.namespace, args)
	case types.ETTrainingJob:
		args := job.Args().(*types.SubmitETJobArgs)
		return training.SubmitETJob(t.namespace, args)
	case types.SparkTrainingJob:
		args := job.Args().(*types.SubmitSparkJobArgs)
		return training.SubmitSparkJob(t.namespace, args)
	case types.DeepSpeedTrainingJob:
		args := job.Args().(*types.SubmitDeepSpeedJobArgs)
		return training.SubmitDeepSpeedJob(t.namespace, args)
	case types.RayJob:
		args := job.Args().(*types.SubmitRayJobArgs)
		return training.SubmitRayJob(t.namespace, args)
	}
	return nil
}

// ScaleIn scales in job
func (t *TrainingJobClient) ScaleIn(job *apistraining.Job) error {
	switch job.Type() {
	case types.ETTrainingJob:
		args := job.Args().(*types.ScaleInETJobArgs)
		return training.SubmitScaleInETJob(t.namespace, args)
	}
	return nil
}

// ScaleOut scales out job
func (t *TrainingJobClient) ScaleOut(job *apistraining.Job) error {
	switch job.Type() {
	case types.ETTrainingJob:
		args := job.Args().(*types.ScaleOutETJobArgs)
		return training.SubmitScaleOutETJob(t.namespace, args)
	}
	return nil
}

// Get returns a training job information
func (t *TrainingJobClient) Get(jobName string, jobType types.TrainingJobType, showPrometheusMetric bool) (*types.TrainingJobInfo, error) {
	job, err := training.SearchTrainingJob(jobName, t.namespace, jobType)
	if err != nil {
		return nil, err
	}
	services, nodes := training.PrepareServicesAndNodesForTensorboard([]training.TrainingJob{job}, false)
	jobInfo := training.BuildJobInfo(job, showPrometheusMetric, services, nodes)
	return jobInfo, nil
}

// GetAndPrint print training job information
func (t *TrainingJobClient) GetAndPrint(jobName string, jobType types.TrainingJobType, format string, showEvent bool, showGPU bool) error {
	if utils.TransferPrintFormat(format) == types.UnknownFormat {
		return fmt.Errorf("unknown output format,only support:[wide|json|yaml]")
	}
	job, err := training.SearchTrainingJob(jobName, t.namespace, jobType)
	if err != nil {
		if err == types.ErrTrainingJobNotFound {
			return fmt.Errorf(errJobNotFoundMessage, jobName, t.namespace)
		}
		return err
	}

	// Search model version associated with the job
	jobLabels := job.GetLabels()
	mv := searchModelVersionByJobLabels(t.namespace, t.configer, jobLabels)
	training.PrintTrainingJob(job, mv, format, showEvent, showGPU)
	return nil
}

// List returns all training jobs
func (t *TrainingJobClient) List(allNamespaces bool, trainingType types.TrainingJobType, showPrometheusMetric bool) ([]*types.TrainingJobInfo, error) {
	jobs, err := training.ListTrainingJobs(t.namespace, allNamespaces, trainingType)
	if err != nil {
		return nil, err
	}
	jobInfos := []*types.TrainingJobInfo{}
	services, nodes := training.PrepareServicesAndNodesForTensorboard(jobs, allNamespaces)
	for _, job := range jobs {
		jobInfos = append(jobInfos, training.BuildJobInfo(job, showPrometheusMetric, services, nodes))
	}
	return jobInfos, nil
}

// ListAndPrint lists and prints the job informations
func (t *TrainingJobClient) ListAndPrint(allNamespaces bool, format string, trainingType types.TrainingJobType) error {
	if utils.TransferPrintFormat(format) == types.UnknownFormat {
		return fmt.Errorf("unknown output format,only support:[wide|json|yaml]")
	}
	jobs, err := training.ListTrainingJobs(t.namespace, allNamespaces, trainingType)
	if err != nil {
		return err
	}
	training.DisplayTrainingJobList(jobs, format, allNamespaces)
	return nil
}

// Logs returns the training job log
func (t *TrainingJobClient) Logs(jobName string, jobType types.TrainingJobType, args *types.LogArgs) error {
	args.Namespace = t.namespace
	args.JobName = jobName
	return training.AcceptJobLog(jobName, jobType, args)
}

func (t *TrainingJobClient) Attach(jobName string, jobType types.TrainingJobType, args *podexec.AttachPodArgs) error {
	job, err := t.Get(jobName, jobType, false)
	if err != nil {
		return err
	}
	if len(job.Instances) == 0 {
		return fmt.Errorf("can not attach the training job %v, because it has no instances", job.Name)
	}
	if args.Options.PodName == "" {
		args.Options.PodName = job.ChiefName
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
func (t *TrainingJobClient) Delete(jobType types.TrainingJobType, jobNames ...string) error {
	for _, jobName := range jobNames {
		err := training.DeleteTrainingJob(jobName, t.namespace, jobType)
		if err != nil {
			if err == types.ErrTrainingJobNotFound {
				return nil
			}
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

// Prune cleans the not running training jobs
func (t *TrainingJobClient) Prune(allNamespaces bool, since time.Duration) error {
	return training.PruneTrainingJobs(t.namespace, allNamespaces, since)
}

func (t *TrainingJobClient) Top(args []string, allNamespaces bool, jobType types.TrainingJobType, instanceName string, notStop bool, format types.FormatStyle) error {
	return training.TopTrainingJobs(args, t.namespace, allNamespaces, jobType, instanceName, notStop, format)
}
