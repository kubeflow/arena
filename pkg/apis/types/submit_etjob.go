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

package types

type SubmitETJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`
	// SubmitTensorboardArgs stores tensorboard information
	SubmitTensorboardArgs `yaml:",inline"`
	// SubmitSyncCodeArgs stores syncing code information
	SubmitSyncCodeArgs  `yaml:",inline"`
	MaxWorkers          int               `yaml:"maxWorkers"`
	MinWorkers          int               `yaml:"minWorkers"`
	LauncherSelectors   map[string]string `yaml:"launcherSelectors"`   // --launcher-selector
	JobRestartPolicy    string            `yaml:"jobRestartPolicy"`    // --job-restart-policy
	WorkerRestartPolicy string            `yaml:"workerRestartPolicy"` // --worker-restart-policy
	JobBackoffLimit     int               `yaml:"jobBackoffLimit"`     // --job-backoff-limit
	// SSHSecret enables create secret for job.
	SSHSecret  string            `yaml:"sshSecret"`
	SecretData map[string]string `yaml:"secretData"`
	// Annotations defines launcher pod annotations of job,match option --launcher-annotation
	LauncherAnnotations map[string]string `yaml:"launcherAnnotations"`
	// Annotations defines worker pod annotations of job,match option --worker-annotation
	WorkerAnnotations map[string]string `yaml:"workerAnnotations"`
}

type ScaleETJobArgs struct {
	//--name string     required, et job name
	Name string `yaml:"etName"`
	// TrainingType stores the trainingType
	JobType TrainingJobType `yaml:"-"`
	// Namespace  stores the namespace of job,match option --namespace
	Namespace string `yaml:"-"`
	//--timeout int     timeout of callback scaler script.
	Timeout int `yaml:"timeout"`
	//--retry int       retry times.
	Retry int `yaml:"retry"`
	//--count int       the nums of you want to add or delete worker.
	Count int `yaml:"count"`
	//--script string        script of scaling.
	Script string `yaml:"script"`
	//-e, --env stringArray      the environment variables
	Envs map[string]string `yaml:"envs"`
}

type ScaleInETJobArgs struct {
	// common args
	ScaleETJobArgs `yaml:",inline"`
}

type ScaleOutETJobArgs struct {
	// common args
	ScaleETJobArgs `yaml:",inline"`
}
