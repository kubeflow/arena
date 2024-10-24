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

type SubmitRayJobArgs struct {
	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// for tensorboard
	SubmitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`

	// ShutdownAfterJobFinishes will determine whether to delete the ray cluster once rayJob succeed or failed.
	// default:=false
	ShutdownAfterJobFinishes bool `yaml:"shutdownAfterJobFinishes,omitempty"`

	// TTLSecondsAfterFinished is the TTL to clean up RayCluster.
	// It's only working when ShutdownAfterJobFinishes set to true.
	// default:=0
	TTLSecondsAfterFinished int32 `yaml:"ttlSecondsAfterFinished,omitempty"`

	// ActiveDeadlineSeconds is the duration in seconds that the RayJob may be active before
	// KubeRay actively tries to terminate the RayJob; value must be positive integer.
	ActiveDeadlineSeconds int32 `yaml:"activeDeadlineSeconds,omitempty"`

	// suspend specifies whether the RayJob controller should create a RayCluster instance
	// If a job is applied with the suspend field set to true,
	// the RayCluster will not be created and will wait for the transition to false.
	// If the RayCluster is already created, it will be deleted.
	// In case of transition to false a new RayCluster will be created.
	Suspend bool `yaml:"suspend,omitempty"`

	RayClusterSpec `yaml:",inline"`

	// ShareMemory Specifies the shared memory size
	ShareMemory string `yaml:"shareMemory"`
}

type RayClusterSpec struct {
	// The version of Ray you are using. Make sure all Ray containers are running this version of Ray.
	RayVersion string `yaml:"rayVersion"`
	// EnableInTreeAutoscaling indicates whether operator should create in tree autoscaling configs
	EnableInTreeAutoscaling bool `yaml:"enableInTreeAutoscaling,omitempty"`
	// AutoscalerOptions specifies optional configuration for the Ray autoscaler.
	AutoscalerOptions AutoscalerOptions `yaml:"autoscalerOptions,omitempty"`

	HeadGroupSpec HeadGroupSpec `yaml:"head"`

	WorkerGroupSpec WorkerGroupSpec `yaml:"worker"`
	// the command that needs to be executed before stopping
	PreStopCmd []string `yaml:"preStopCmd"`
}

// AutoscalerOptions specifies optional configuration for the Ray autoscaler.
type AutoscalerOptions struct {
	// cpu specifies optional resource request and limit overrides for the autoscaler container.
	// Default values: 500m CPU request and limit.
	Cpu string `yaml:"cpu,omitempty"`
	// memory specifies optional resource request and limit overrides for the autoscaler
	//  Default values: 512Mi memory request and limit.
	Memory string `yaml:"memory,omitempty"`
	// Image optionally overrides the autoscaler's container image. This override is for provided for autoscaler testing and development.
	Image string `yaml:"image,omitempty"`
	// ImagePullPolicy optionally overrides the autoscaler container's image pull policy. This override is for provided for autoscaler testing and development.
	ImagePullPolicy string `yaml:"imagePullPolicy,omitempty"`
	// IdleTimeoutSeconds is the number of seconds to wait before scaling down a worker pod which is not using Ray resources.
	// Defaults to 60 (one minute). It is not read by the KubeRay operator but by the Ray autoscaler.
	IdleTimeoutSeconds int32 `yaml:"idleTimeoutSeconds,omitempty"`
	// UpscalingMode is "Conservative", "Default", or "Aggressive."
	// Conservative: Upscaling is rate-limited; the number of pending worker pods is at most the size of the Ray cluster.
	// Default: Upscaling is not rate-limited.
	// Aggressive: An alias for Default; upscaling is not rate-limited.
	// It is not read by the KubeRay operator but by the Ray autoscaler.
	// +kubebuilder:validation:Enum=Default;Aggressive;Conservative
	UpscalingMode string `yaml:"upscalingMode,omitempty"`
}

// HeadGroupSpec are the spec for the head pod
type HeadGroupSpec struct {
	// ServiceType is Kubernetes service type of the head service. it will be used by the workers to connect to the head pod
	ServiceType string `yaml:"serviceType,omitempty"`
	Image       string `yaml:"image"`
	Cpu         string `yaml:"cpu"`
	Memory      string `yaml:"memory"`
	Gpu         int    `yaml:"gpu"`
}

// WorkerGroupSpec are the specs for the worker pods
type WorkerGroupSpec struct {
	Image  string `yaml:"image"`
	Cpu    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
	Gpu    int    `yaml:"gpu"`
	// Replicas is the number of desired Pods for this worker group.
	// +kubebuilder:default:=0
	Replicas int32 `yaml:"replicas,omitempty"`
	// MinReplicas denotes the minimum number of desired Pods for this worker group.
	// +kubebuilder:default:=0
	MinReplicas int32 `yaml:"minReplicas"`
	// MaxReplicas denotes the maximum number of desired Pods for this worker group, and the default value is maxInt32.
	// +kubebuilder:default:=2147483647
	MaxReplicas int32 `yaml:"maxReplicas"`
	// NumOfHosts denotes the number of hosts to create per replica. The default value is 1.
	// +kubebuilder:default:=1
	NumOfHosts int32 `yaml:"numOfHosts,omitempty"`
}
