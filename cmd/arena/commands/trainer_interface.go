package commands

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// The Training Job can be TensorFlow, MPI and Caffe
type TrainingJob interface {
	// Get the chief Pod of the Job.
	ChiefPod() v1.Pod

	// Get the name of the Training Job
	Name() string

	// Get all the pods of the Training Job
	AllPods() []v1.Pod

	// Get the Status of the Job: RUNNING, PENDING,
	GetStatus() string

	// Return trainer Type, support MPI, standalone, tensorflow
	Trainer() string

	// Get the Job Age
	Age() string

	// Get start time
	StartTime() *metav1.Time

	// Get Dashboard
	GetJobDashboards(client *kubernetes.Clientset) ([]string, error)

	// Requested GPU count of the Job
	RequestedGPU() int64

	// Requested GPU count of the Job
	AllocatedGPU() int64

	// the host ip of the chief pod
	HostIPOfChief() string
}

type Trainer interface {
	// Check if the training job is supported
	IsSupported(name, ns string) bool

	// Get TrainingJob object directly. this method is called when `arena get`
	GetTrainingJob(name, namespace string) (TrainingJob, error)

	// Get the type of trainer
	Type() string
}
