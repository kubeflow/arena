package commands

import (
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type RunaiJob struct {
	*BasicJobInfo
	job         batchv1.Job
	trainerType string
	chiefPod    v1.Pod
}

// // Get the chief Pod of the Job.
func (rj *RunaiJob) ChiefPod() v1.Pod {
	return rj.chiefPod
}

// Get the name of the Training Job
func (rj *RunaiJob) Name() string {
	return rj.name
}

// Get the namespace of the Training Job
func (rj *RunaiJob) Namespace() string {
	return rj.job.Namespace
}

// Get all the pods of the Training Job
func (rj *RunaiJob) AllPods() []v1.Pod {
	return []v1.Pod{rj.chiefPod}
}

// Get all the kubernetes resource of the Training Job
func (rj *RunaiJob) Resources() []Resource {
	return rj.resources
}

// Get the Status of the Job: RUNNING, PENDING,
func (rj *RunaiJob) GetStatus() string {
	return string(rj.chiefPod.Status.Phase)
}

// Return trainer Type, support MPI, standalone, tensorflow
func (rj *RunaiJob) Trainer() string {
	return rj.trainerType
}

// Get the Job Age
func (rj *RunaiJob) Age() time.Duration {
	job := rj.job
	if job.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.CreationTimestamp.Time)
}

// TODO
// Get the Job Duration
func (rj *RunaiJob) Duration() time.Duration {
	return 0
}

// TODO
// Get start time
func (rj *RunaiJob) StartTime() *metav1.Time {
	return &rj.job.CreationTimestamp
}

// Get Dashboard
func (rj *RunaiJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	return []string{}, nil
}

// Requested GPU count of the Job
func (rj *RunaiJob) RequestedGPU() int64 {
	val, ok := rj.job.Spec.Template.Spec.Containers[0].Resources.Limits[NVIDIAGPUResourceName]
	if !ok {
		return 0
	}

	return val.Value()
}

// Requested GPU count of the Job
func (rj *RunaiJob) AllocatedGPU() int64 {
	pod := rj.chiefPod

	if pod.Status.Phase == v1.PodRunning {
		return rj.RequestedGPU()
	}

	return 0
}

// the host ip of the chief pod
func (rj *RunaiJob) HostIPOfChief() string {
	return ""
}

// The priority class name of the training job
func (rj *RunaiJob) GetPriorityClass() string {
	return ""
}
