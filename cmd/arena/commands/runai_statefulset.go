package commands

import (
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type RunaiStatefulSet struct {
	*BasicJobInfo
	statefulSet appv1.StatefulSet
	trainerType string
	chiefPod    v1.Pod
}

// // Get the chief Pod of the Job.
func (rs *RunaiStatefulSet) ChiefPod() v1.Pod {
	return rs.chiefPod
}

// Get the name of the Training Job
func (rs *RunaiStatefulSet) Name() string {
	return rs.name
}

// Get the namespace of the Training Job
func (rs *RunaiStatefulSet) Namespace() string {
	return rs.statefulSet.Namespace
}

// Get all the pods of the Training Job
func (rs *RunaiStatefulSet) AllPods() []v1.Pod {
	return []v1.Pod{rs.chiefPod}
}

// Get all the kubernetes resource of the Training Job
func (rs *RunaiStatefulSet) Resources() []Resource {
	return rs.resources
}

// Get the Status of the Job: RUNNING, PENDING,
func (rs *RunaiStatefulSet) GetStatus() string {
	return string(rs.chiefPod.Status.Phase)
}

// Return trainer Type, support MPI, standalone, tensorflow
func (rs *RunaiStatefulSet) Trainer() string {
	return rs.trainerType
}

// Get the Job Age
func (rs *RunaiStatefulSet) Age() time.Duration {
	statefulSet := rs.statefulSet
	if statefulSet.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(statefulSet.CreationTimestamp.Time)
}

// TODO
// Get the Job Duration
func (rs *RunaiStatefulSet) Duration() time.Duration {
	return 0
}

// TODO
// Get start time
func (rs *RunaiStatefulSet) StartTime() *metav1.Time {
	return &rs.statefulSet.CreationTimestamp
}

// Get Dashboard
func (rs *RunaiStatefulSet) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	return []string{}, nil
}

// Requested GPU count of the Job
func (rs *RunaiStatefulSet) RequestedGPU() int64 {
	val, ok := rs.statefulSet.Spec.Template.Spec.Containers[0].Resources.Limits[NVIDIAGPUResourceName]
	if !ok {
		return 0
	}

	return val.Value()
}

// Requested GPU count of the Job
func (rs *RunaiStatefulSet) AllocatedGPU() int64 {
	pod := rs.chiefPod

	if pod.Status.Phase == v1.PodRunning {
		return rs.RequestedGPU()
	}

	return 0
}

// the host ip of the chief pod
func (rs *RunaiStatefulSet) HostIPOfChief() string {
	return ""
}

// The priority class name of the training job
func (rs *RunaiStatefulSet) GetPriorityClass() string {
	return ""
}
