/*

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
	"github.com/kubeflow/arena/dependency/operators/et-operator/api/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TrainingJobSpec defines the desired state of TrainingJob
type TrainingJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// CleanPodPolicy defines the policy that whether to kill pods after the job completes.
	// Defaults to None.
	CleanPodPolicy *common.CleanPodPolicy `json:"cleanPodPolicy,omitempty"`

	// `ETReplicaSpecs` contains maps from `ETReplicaType` to `ReplicaSpec` that
	// specify the ET replicas to run.
	ETReplicaSpecs ETReplicaSpecs `json:"etReplicaSpecs"`

	// Specifies the number of slots per worker used in hostfile.
	// Defaults to 1.
	// +optional
	SlotsPerWorker *int32 `json:"slotsPerWorker,omitempty"`
}

type ETReplicaSpecs struct {
	Launcher *common.ReplicaSpec `json:"Launcher"`
	Worker   *ETReplicaSpec      `json:"Worker"`
}

type ETReplicaSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	Replicas *int32 `json:"Replicas,omitempty"`

	// MaxReplicas is the desired max number of replicas of the given template.
	// If unspecified, MaxReplicas defaults to infinite.
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// MinReplicas is the desired min number of replicas of the given template.
	// If unspecified, MinReplicas defaults to InitReplicas
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// Template is the object that describes the pod that
	// will be created for this replica. RestartPolicy in PodTemplateSpec
	// will be overide by RestartPolicy in ReplicaSpec
	Template v1.PodTemplateSpec `json:"template,omitempty"`

	// Restart policy for all replicas within the job.
	// One of Always, OnFailure, Never and ExitCode.
	// Default to Never.
	RestartPolicy *common.RestartPolicy `json:"restartPolicy,omitempty"`
}

// ETReplicaType is the type for ETReplica.
type ETReplicaType common.ReplicaType

const (
	// ETReplicaTypeLauncher is the type for launcher replica.
	ETReplicaTypeLauncher ETReplicaType = "Launcher"

	// ETReplicaTypeWorker is the type for worker replicas.
	ETReplicaTypeWorker ETReplicaType = "Worker"
)

// TrainingJobStatus defines the observed state of TrainingJob
type TrainingJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=TypeMeta
// +resource:path=trainingjob
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:object:root=true

// TrainingJob is the Schema for the trainingjobs API
type TrainingJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TrainingJobSpec `json:"spec,omitempty"`

	// Most recently observed status of the PyTorchJob.
	// Read-only (modified by the system).
	Status common.JobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TrainingJobList contains a list of TrainingJob
type TrainingJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrainingJob `json:"items"`
}
