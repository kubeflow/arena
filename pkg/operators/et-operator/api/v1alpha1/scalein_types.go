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
	"github.com/kubeflow/arena/pkg/operators/et-operator/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScaleInSpec defines the desired state of ScaleIn
type ScaleInSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Script string `json:"script,omitempty"`

	// Optional number of retries to execute script.
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// Optional number of timeout to execute script.
	// +optional
	Timeout *int32 `json:"timeout,omitempty"`

	ToDelete *ToDeleteSpec `json:"toDelete,omitempty"`

	Selector Selector `json:"selector,omitempty"`

	Env []EnvSpec `json:"env,omitempty"`
}

type ToDeleteSpec struct {
	Count    int      `json:"count,omitempty"`
	PodNames []string `json:"podNames,omitempty"`
}

// ScaleInStatus defines the observed state of ScaleIn
type ScaleInStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:subresource:Spec
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:object:root=true

// ScaleIn is the Schema for the scaleins API
type ScaleIn struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScaleInSpec `json:"spec,omitempty"`
	// Most recently observed status of the PyTorchJob.
	// Read-only (modified by the system).
	Status common.JobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScaleInList contains a list of ScaleIn
type ScaleInList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScaleIn `json:"items"`
}
