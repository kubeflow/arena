package commands

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type RunaiJob struct {
	*BasicJobInfo
	trainerType       string
	chiefPod          v1.Pod
	creationTimestamp metav1.Time
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
	return rj.chiefPod.Namespace
}

// Get all the pods of the Training Job
func (rj *RunaiJob) AllPods() []v1.Pod {
	return []v1.Pod{rj.chiefPod}
}

// Get all the kubernetes resource of the Training Job
func (rj *RunaiJob) Resources() []Resource {
	return rj.resources
}

func (rj *RunaiJob) getStatus() v1.PodPhase {
	return rj.chiefPod.Status.Phase
}

// Get the Status of the Job: RUNNING, PENDING,
func (rj *RunaiJob) GetStatus() string {
	return string(rj.getStatus())
}

// Return trainer Type, support MPI, standalone, tensorflow
func (rj *RunaiJob) Trainer() string {
	return rj.trainerType
}

// Get the Job Age
func (rj *RunaiJob) Age() time.Duration {
	if rj.creationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(rj.creationTimestamp.Time)
}

// TODO
// Get the Job Duration
func (rj *RunaiJob) Duration() time.Duration {
	status := rj.getStatus()
	startTime := rj.StartTime()

	if startTime == nil {
		return 0
	}

	var finishTime metav1.Time = metav1.Now()

	if status == v1.PodSucceeded || status == v1.PodFailed {
		// The transition time of ready will be when the pod finished executing
		for _, condition := range rj.ChiefPod().Status.Conditions {
			if condition.Type == v1.PodReady {
				finishTime = condition.LastTransitionTime
			}
		}
	}

	return finishTime.Sub(startTime.Time)
}

// TODO
// Get start time
func (rj *RunaiJob) StartTime() *metav1.Time {
	pod := rj.ChiefPod()
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodInitialized && condition.Status == v1.ConditionTrue {
			return &condition.LastTransitionTime
		}
	}

	return nil
}

// Get Dashboard
func (rj *RunaiJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	return []string{}, nil
}

// Requested GPU count of the Job
func (rj *RunaiJob) RequestedGPU() int64 {
	val, ok := rj.chiefPod.Spec.Containers[0].Resources.Limits[NVIDIAGPUResourceName]
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
