package client

import "github.com/kubeflow/arena/pkg/task"

// JobStatus represents the status of a training job CRD in the cluster.
type JobStatus struct {
	Name         string `json:"name" yaml:"name"`
	Namespace    string `json:"namespace" yaml:"namespace"`
	Status       string `json:"status" yaml:"status"` // Running, Succeeded, Failed, Unknown
	APIVersion   string `json:"apiVersion" yaml:"apiVersion"`
	Framework    string `json:"framework" yaml:"framework"`
	Replicas     int    `json:"replicas" yaml:"replicas"`
	Ready        int    `json:"ready" yaml:"ready"`
	Age          string `json:"age" yaml:"age"`
	GPURequested int    `json:"gpuRequested" yaml:"gpuRequested"`
}

// JobInfo contains detailed information about a training job, including its status and pod information.
type JobInfo struct {
	Status        JobStatus  `json:"status" yaml:"status"`
	Pods          []PodInfo  `json:"pods" yaml:"pods"`
	Configuration *task.Task `json:"configuration,omitempty" yaml:"configuration,omitempty"`
}

// PodInfo represents information about a pod belonging to a training job.
type PodInfo struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	IP     string `json:"ip" yaml:"ip"`
	Node   string `json:"node" yaml:"node"`
}
