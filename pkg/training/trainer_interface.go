// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package training

import (
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// The Training Job can be TensorFlow, MPI and Caffe
type TrainingJob interface {
	// Get the chief Pod of the Job.
	ChiefPod() *corev1.Pod

	// Get the name of the Training Job
	Name() string

	// Get the unique identity of the Training Job
	Uid() string

	// Get the namespace of the Training Job
	Namespace() string

	// Get all the pods of the Training Job
	AllPods() []*corev1.Pod

	// Get all the kubernetes resource of the Training Job
	Resources() []Resource

	// Get the Status of the Job: RUNNING, PENDING,
	GetStatus() string

	// Return trainer Type, match the training job type
	Trainer() types.TrainingJobType

	// Get the Job Age
	Age() time.Duration

	// Get the Job Duration
	Duration() time.Duration

	// Get start time
	StartTime() *metav1.Time

	// Get Dashboard
	GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error)

	// Requested GPU count of the Job
	RequestedGPU() int64

	// Requested GPU count of the Job
	AllocatedGPU() int64

	// the host ip of the chief pod
	HostIPOfChief() string

	// The priority class name of the training job
	GetPriorityClass() string

	GetTrainJob() interface{}

	GetLabels() map[string]string
}

type Trainer interface {
	// IsEnabled is used to check the trainer is enabled or not
	IsEnabled() bool
	// Check if the training job is supported
	IsSupported(name, ns string) bool

	// Get TrainingJob object directly. this method is called when `arena get`
	GetTrainingJob(name, namespace string) (TrainingJob, error)

	// Get the type of trainer
	Type() types.TrainingJobType

	// List all tf training jobs
	ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error)
}
