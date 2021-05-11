/*
Copyright 2021 The Alibaba Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// JobConditionType defines all kinds of types of JobStatus.
type JobConditionType string

const (
	// JobCreated means the job has been accepted by the system,
	// but one or more of the pods/services has not been started.
	// This includes time before pods being scheduled and launched.
	JobCreated JobConditionType = "Created"

	// JobRunning means all sub-resources (e.g. services/pods) of this job
	// have been successfully scheduled and launched.
	// The training is running without error.
	JobRunning JobConditionType = "Running"

	// JobRestarting means one or more sub-resources (e.g. services/pods) of this job
	// reached phase failed but maybe restarted according to it's restart policy
	// which specified by user in v1.PodTemplateSpec.
	// The training is freezing/pending.
	JobRestarting JobConditionType = "Restarting"

	// JobSucceeded means all sub-resources (e.g. services/pods) of this job
	// reached phase have terminated in success.
	// The training is complete without error.
	JobSucceeded JobConditionType = "Succeeded"

	// JobFailed means one or more sub-resources (e.g. services/pods) of this job
	// reached phase failed with no restarting.
	// The training has failed its execution.
	JobFailed JobConditionType = "Failed"
)

// CronTemplateSpec describes a template for launching a specific job.
type CronTemplateSpec struct {
	metav1.TypeMeta `json:",inline"`

	// Workload is the specification of the desired cron job with specific types.
	// +kubebuilder:pruning:PreserveUnknownFields
	Workload *runtime.RawExtension `json:"workload,omitempty"`
}

// CronSpec defines the desired state of Cron
type CronSpec struct {
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Specifies the job that will be created when executing a CronTask.
	CronTemplate CronTemplateSpec `json:"template"`

	// Specifies how to treat concurrent executions of a Task.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// Deadline is the timestamp that a cron job can keep scheduling util then.
	Deadline *metav1.Time `json:"deadline,omitempty"`

	// The number of finished job history to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	HistoryLimit *int32 `json:"historyLimit,omitempty"`
}

// CronStatus defines the observed state of Cron
type CronStatus struct {
	// A list of currently running jobs.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// History is a list of scheduled cron job with its digest records.
	// +optional
	History []CronHistory `json:"history,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

type CronHistory struct {
	// Object is the reference of the historical scheduled cron job.
	Object corev1.TypedLocalObjectReference `json:"object"`
	// Status is the final status when job finished.
	Status JobConditionType `json:"status"`
	// Created is the creation timestamp of job.
	Created *metav1.Time `json:"created,omitempty"`
	// Finished is the failed or succeeded timestamp of job.
	Finished *metav1.Time `json:"finished,omitempty"`
}

// ConcurrencyPolicy describes how the job will be handled.
// Only one of the following concurrent policies may be specified.
// If none of the following policies is specified, the default one
// is AllowConcurrent.
type ConcurrencyPolicy string

const (
	// AllowConcurrent allows CronJobs to run concurrently.
	AllowConcurrent ConcurrencyPolicy = "Allow"

	// ForbidConcurrent forbids concurrent runs, skipping next run if previous
	// hasn't finished yet.
	ForbidConcurrent ConcurrencyPolicy = "Forbid"

	// ReplaceConcurrent cancels currently running job and replaces it with a new one.
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// +genclient
// +kubebuilder:object:root=true

// Cron is the Schema for the crons API
type Cron struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronSpec   `json:"spec,omitempty"`
	Status CronStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronList contains a list of Cron
type CronList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cron `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cron{}, &CronList{})
}
