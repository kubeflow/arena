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
// limitations under the License
package types

// CommonSubmitArgs defines the common parts of the submitAthd
type CommonSubmitArgs struct {
	// Name stores the job name,match option --name
	Name string `yaml:"-"`

	// Namespace  stores the namespace of job,match option --namespace
	Namespace string `yaml:"-"`

	// NodeSelectors defines the node selectors,match option --selector
	NodeSelectors map[string]string `yaml:"nodeSelectors"`

	// ConfigFiles stores the config file which is existed in client host node
	// and map it to container,match option --config-file
	ConfigFiles map[string]map[string]ConfigFileInfo `yaml:"configFiles"`

	// Tolerations defines the tolerations which tolerates node taints
	// match option --toleration
	Tolerations []string `yaml:"tolerations"`

	// Image stores the docker image of job,match option --image
	Image string `yaml:"image"`

	// GPUCount stores the gpu count of the job needs,match option --gpus
	GPUCount int `yaml:"gpuCount"`

	// Envs stores the envs of container in job, match option --env
	Envs map[string]string `yaml:"envs"`

	// WorkingDir stores the working directory of container in job,match option --working-dir
	WorkingDir string `yaml:"workingDir"`

	// Command stores the command of job
	Command string `yaml:"command"`

	// Mode is used for horovod,match option --sync-mode
	Mode string `yaml:"mode"`

	// WorkerCount stores the count of job worker,match option --workers
	WorkerCount int `yaml:"workers"`

	// Retry defines the retry times
	Retry int `yaml:"retry"`

	// DataSet stores the kubernetes pvc names
	DataSet map[string]string `yaml:"dataset"`

	// DataDirs stores the files(or directories) in k8s node which will map to containers
	// match option --data-dir
	DataDirs []DataDirVolume `yaml:"dataDirs"`

	// EnableRDMA enable rdma or not,match option --rdma
	EnableRDMA bool `yaml:"enableRDMA"`

	// UseENI defines using eni or not
	UseENI bool `yaml:"useENI"`

	// Annotations defines pod annotations of job,match option --annotation
	Annotations map[string]string `yaml:"annotations"`

	// IsNonRoot is root user or not
	IsNonRoot bool `yaml:"isNonRoot"`

	// PodSecurityContext defines the pod security context
	PodSecurityContext LimitedPodSecurityContext `yaml:"podSecurityContext"`

	// PriorityClassName defines the priority class
	PriorityClassName string `yaml:"priorityClassName"`

	// Conscheduling defines using Conscheduling
	Conscheduling bool

	// PodGroupName stores pod group name
	PodGroupName string `yaml:"podGroupName"`

	// PodGroupMinAvailable stores pod group min available
	PodGroupMinAvailable string `yaml:"podGroupMinAvailable"`

	// ImagePullSecrets stores image pull secrets,match option --image-pull-secrets
	ImagePullSecrets []string `yaml:"imagePullSecrets"`

	// HelmOptions stores the helm options
	HelmOptions []string `yaml:"-"`
}

// DataDirVolume defines the volume of kubernetes
type DataDirVolume struct {
	// HostPath defines the host path
	HostPath string `yaml:"hostPath"`
	// ContainerPath defines container path
	ContainerPath string `yaml:"containerPath"`
	// Name defines the volume name
	Name string `yaml:"name"`
}

// LimitedPodSecurityContext defines the kuberntes pod security context
type LimitedPodSecurityContext struct {
	RunAsUser          int64   `yaml:"runAsUser"`
	RunAsNonRoot       bool    `yaml:"runAsNonRoot"`
	RunAsGroup         int64   `yaml:"runAsGroup"`
	SupplementalGroups []int64 `yaml:"supplementalGroups"`
}

// ConfigFileInfo defines the config files which will be mounted to containers
type ConfigFileInfo struct {
	ContainerFileName string `yaml:"containerFileName"`
	HostFile          string `yaml:"hostFile"`
	Key               string `yaml:"key"`
	ContainerFilePath string `yaml:"containerFilePath"`
}

type SubmitSyncCodeArgs struct {
	SyncMode   string `yaml:"syncMode"`            // --syncMode: rsync, hdfs, git
	SyncSource string `yaml:"syncSource"`          // --syncSource
	SyncImage  string `yaml:"syncImage,omitempty"` // --syncImage
	// syncGitProjectName
	SyncGitProjectName string `yaml:"syncGitProjectName,omitempty"` // --syncImage
}
