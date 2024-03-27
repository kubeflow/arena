package analyze

import (
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ModelJob interface {
	// Name return the job name
	Name() string
	// Namespace return the namespace
	Namespace() string
	// Type return the type
	Type() types.ModelJobType
	// Pods return the job pods
	Pods() []*v1.Pod
	// Job return the job
	Job() *batchv1.Job
	// Age return the job age
	Age() time.Duration
	// Duration return the job duration
	Duration() time.Duration
	// Status return the job status
	Status() string
	// StartTime return start time
	StartTime() *metav1.Time
	// RequestCPUs return the cpus which model job owned
	RequestCPUs() int64
	// RequestGPUs return the gpus which model job owned
	RequestGPUs() int64
	// RequestGPUMemory return the gpu memory,only for gpushare
	RequestGPUMemory() int64
	// RequestGPUCore return the gpu core,only for gpushare
	RequestGPUCore() int64
	// Params return the job parameters
	Params() map[string]string
	// Convert2JobInfo convert to ModelJobInfo
	Convert2JobInfo() types.ModelJobInfo
}

type Processor interface {
	// Type returns the processor type
	Type() types.ModelJobType
	// GetModelJob is used to get a model job
	GetModelJob(namespace, name string) (ModelJob, error)
	// ListModelJobs is used to list all model jobs
	ListModelJobs(namespace string, allNamespace bool) ([]ModelJob, error)
}
