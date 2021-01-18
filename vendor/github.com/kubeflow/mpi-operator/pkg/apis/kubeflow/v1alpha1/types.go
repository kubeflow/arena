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
	// Note that this is deprecated in favor of `ProcessingUnits` field.
	// +optional
	GPUs *int32 `json:"gpus,omitempty"`

	// The maximum number of GPUs available per node.
	// Note that this will be ignored if the GPU resources are explicitly
	// specified in the MPIJob pod spec.
	// This is deprecated in favor of `ProcessingUnitsPerNode` field.
	GPUsPerNode *int32 `json:"gpusPerNode,omitempty"`

	// Specifies the desired number of processing units the MPIJob should run on.
	// Mutually exclusive with the `Replicas` field.
	// +optional
	ProcessingUnits *int32 `json:"processingUnits,omitempty"`

	// The maximum number of processing units available per node.
	// Note that this will be ignored if the processing resources are explicitly
	// specified in the MPIJob pod spec.
	// +optional
	ProcessingUnitsPerNode *int32 `json:"processingUnitsPerNode,omitempty"`

	// The processing resource type, e.g. 'nvidia.com/gpu' or 'cpu'.
	// Defaults to 'nvidia.com/gpu'
	// +optional
	ProcessingResourceType string `json:"processingResourceType,omitempty"`

	// Specifies the number of slots per worker used in hostfile.
	// Defaults to the number of processing units per worker.
	// +optional
	SlotsPerWorker *int32 `json:"slotsPerWorker,omitempty"`

	// Run the launcher on the master.
	// Defaults to false.
	// +optional
	LauncherOnMaster bool `json:"launcherOnMaster,omitempty"`

	// Specifies the number of retries before marking this job failed.
	// Defaults to 6.
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// Specifies the duration in seconds relative to the start time that
	// the job may be active before the system tries to terminate it.
	// Note that this takes precedence over `BackoffLimit` field.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// Specifies the desired number of replicas the MPIJob should run on.
	// The `PodSpec` should specify the number of processing units.
	// Mutually exclusive with the `GPUs` or `ProcessingUnits` fields.
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

	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}
