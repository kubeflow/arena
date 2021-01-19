// Copyright 2019 The Kubeflow Authors.
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

package v1alpha2

import (
	common "github.com/kubeflow/common/pkg/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MPIJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MPIJobSpec       `json:"spec,omitempty"`
	Status            common.JobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MPIJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []MPIJob `json:"items"`
}

type MPIJobSpec struct {

	// Specifies the number of slots per worker used in hostfile.
	// Defaults to 1.
	// +optional
	SlotsPerWorker *int32 `json:"slotsPerWorker,omitempty"`

	// Specifies the number of retries before marking this job failed.
	// Defaults to 6.
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// Specifies the duration in seconds relative to the start time that
	// the job may be active before the system tries to terminate it.
	// Note that this takes precedence over `BackoffLimit` field.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// CleanPodPolicy defines the policy that whether to kill pods after the job completes.
	// Defaults to None.
	CleanPodPolicy *common.CleanPodPolicy `json:"cleanPodPolicy,omitempty"`

	// `MPIReplicaSpecs` contains maps from `MPIReplicaType` to `ReplicaSpec` that
	// specify the MPI replicas to run.
	MPIReplicaSpecs map[MPIReplicaType]*common.ReplicaSpec `json:"mpiReplicaSpecs"`

	// MainContainer specifies name of the main container which
	// executes the MPI code.
	MainContainer string `json:"mainContainer,omitempty"`

	// `RunPolicy` encapsulates various runtime policies of the distributed training
	// job, for example how to clean up resources and how long the job can stay
	// active. The policies specified in `RunPolicy` take precedence over
	// the following fields: `BackoffLimit` and `ActiveDeadlineSeconds`.
	RunPolicy *common.RunPolicy `json:"runPolicy,omitempty"`

	// MPIDistribution specifies name of the MPI framwork which is used
	// Defaults to "OpenMPI"
	// Options includes "OpenMPI", "IntelMPI" and "MPICH"
	MPIDistribution *MPIDistributionType `json:"mpiDistribution,omitempty"`
}

// MPIReplicaType is the type for MPIReplica.
type MPIReplicaType common.ReplicaType

const (
	// MPIReplicaTypeLauncher is the type for launcher replica.
	MPIReplicaTypeLauncher MPIReplicaType = "Launcher"

	// MPIReplicaTypeWorker is the type for worker replicas.
	MPIReplicaTypeWorker MPIReplicaType = "Worker"
)

// MPIDistributionType is the type for MPIDistribution.
type MPIDistributionType string

const (
	// MPIDistributionTypeOpenMPI is the type for Open MPI.
	MPIDistributionTypeOpenMPI MPIDistributionType = "OpenMPI"

	// MPIDistributionTypeIntelMPI is the type for Intel MPI.
	MPIDistributionTypeIntelMPI MPIDistributionType = "IntelMPI"

	// MPIDistributionTypeMPICH is the type for MPICh.
	MPIDistributionTypeMPICH MPIDistributionType = "MPICH"
)
