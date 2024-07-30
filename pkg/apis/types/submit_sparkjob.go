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

type SubmitSparkJobArgs struct {
	Name         string          `yaml:"-"`
	Namespace    string          `yaml:"-"`
	TrainingType TrainingJobType `yaml:"-"`
	Image        string          `yaml:"Image"`
	MainClass    string          `yaml:"MainClass"`
	Jar          string          `yaml:"Jar"`
	SparkVersion string          `yaml:"SparkVersion"`
	Driver       *Driver         `yaml:"Driver"`
	Executor     *Executor       `yaml:"Executor"`
	// Annotations defines pod annotations of job,match option --annotation
	Annotations map[string]string `yaml:"annotations"`
	// Labels specify the job labels and it is work for pods
	Labels map[string]string `yaml:"labels"`
}

type Driver struct {
	CPURequest     int    `yaml:"CPURequest"`
	MemoryRequest  string `yaml:"MemoryRequest"`
	ServiceAccount string `yaml:"ServiceAccount"`
}

type Executor struct {
	Replicas      int    `yaml:"Replicas"`
	CPURequest    int    `yaml:"CPURequest"`
	MemoryRequest string `yaml:"MemoryRequest"`
}
