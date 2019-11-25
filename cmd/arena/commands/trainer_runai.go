package commands

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		return 0
	}

	return rj.RequestedGPU()
}

// the host ip of the chief pod
func (rj *RunaiJob) HostIPOfChief() string {
	return ""
}

// The priority class name of the training job
func (rj *RunaiJob) GetPriorityClass() string {
	return ""
}

type RunaiTrainer struct {
	client *kubernetes.Clientset
}

func NewRunaiTrainer(client *kubernetes.Clientset) Trainer {
	return &RunaiTrainer{
		client: client,
	}
}

func (rt *RunaiTrainer) IsSupported(name, ns string) bool {
	runaiList, err := rt.client.Batch().Jobs(ns).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s,app=runaijob", name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiList.Items) > 0 {
		return true
	} else {
		return false
	}
}

func (rt *RunaiTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		job batchv1.Job
	)

	runaiList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s,app=runaijob", name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		job = runaiList.Items[0]
	}

	return rt.getTrainingJob(job)

}

func (rt *RunaiTrainer) Type() string {
	return defaultRunaiTrainingType
}

func (rt *RunaiTrainer) getTrainingJob(job batchv1.Job) (TrainingJob, error) {
	var (
		lastCreatedPod v1.Pod
	)

	podList, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
	})

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	pods := podList.Items
	lastCreatedPod = pods[0]
	otherPods := pods[1:]
	for _, item := range otherPods {
		if lastCreatedPod.CreationTimestamp.Before(&item.CreationTimestamp) {
			lastCreatedPod = item
		}
	}

	return &RunaiJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(pods),
			name:      job.Name,
		},
		chiefPod:    lastCreatedPod,
		job:         job,
		trainerType: rt.Type(),
	}, nil
}

func (rt *RunaiTrainer) ListTrainingJobs() ([]TrainingJob, error) {
	runaiJobs := []TrainingJob{}

	runaiList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", "runaijob"),
	})

	if err != nil {
		return nil, err
	}

	for _, item := range runaiList.Items {
		runaiJob, err := rt.getTrainingJob(item)

		if err != nil {
			return nil, err
		}

		runaiJobs = append(runaiJobs, runaiJob)
	}

	return runaiJobs, nil
}
