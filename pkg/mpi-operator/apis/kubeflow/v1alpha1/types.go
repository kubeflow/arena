// Copyright 2018 The Kubeflow Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MPIJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MPIJobSpec   `json:"spec,omitempty"`
	Status            MPIJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MPIJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []MPIJob `json:"items"`
}

type MPIJobSpec struct {
	// Specifies the desired number of GPUs the MPIJob should run on.
	// Mutually exclusive with the `Replicas` field.
	// +optional
	GPUs *int32 `json:"gpus,omitempty"`

	// Run the launcher on the master.
	// Optional: Default to false
	// +optional
	LauncherOnMaster bool `json:"launcherOnMaster,omitempty"`

	// Optional number of retries before marking this job failed.
	// Defaults to 6
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// Specifies the desired number of replicas the MPIJob should run on.
	// The `PodSpec` should specify the number of GPUs.
	// Mutually exclusive with the `GPUs` field.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Describes the pod that will be created when executing an MPIJob.
	Template corev1.PodTemplateSpec `json:"template,omitempty"`
}

type MPIJobLauncherStatusType string

// These are valid launcher statuses of an MPIJob.
const (
	// LauncherActive means the MPIJob launcher is actively running.
	LauncherActive MPIJobLauncherStatusType = "Active"
	// LauncherSucceeded means the MPIJob launcher has succeeded.
	LauncherSucceeded MPIJobLauncherStatusType = "Succeeded"
	// LauncherFailed means the MPIJob launcher has failed its execution.
	LauncherFailed MPIJobLauncherStatusType = "Failed"
)

type MPIJobStatus struct {
	// Current status of the launcher job.
	// +optional
	LauncherStatus MPIJobLauncherStatusType `json:"launcherStatus,omitempty"`

	// The number of available worker replicas.
	// +optional
	WorkerReplicas int32 `json:"workerReplicas,omitempty"`
}
