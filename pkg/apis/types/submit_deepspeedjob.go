// Copyright 2023 The Kubeflow Authors
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

type SubmitDeepSpeedJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`
	// SubmitTensorboardArgs stores tensorboard information
	SubmitTensorboardArgs `yaml:",inline"`
	// SubmitSyncCodeArgs stores syncing code information
	SubmitSyncCodeArgs `yaml:",inline"`
	LauncherSelectors  map[string]string `yaml:"launcherSelectors"` // --launcher-selector
	JobRestartPolicy   string            `yaml:"jobRestartPolicy"`  // --job-restart-policy
	JobBackoffLimit    int               `yaml:"jobBackoffLimit"`   // --job-backoff-limit
	// Annotations defines launcher pod annotations of job,match option --launcher-annotation
	LauncherAnnotations map[string]string `yaml:"launcherAnnotations"`
	// Annotations defines worker pod annotations of job,match option --worker-annotation
	WorkerAnnotations map[string]string `yaml:"workerAnnotations"`
}
