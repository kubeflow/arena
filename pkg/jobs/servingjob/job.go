// this package is define a serving job
package serving

import (
	"fmt"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	app_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
// ServingJob define a serving job
type ServingJob interface {
	// GetName() returns the serving job name
	GetName() string
	// GetNamespace() returns the namespace of serving job
	GetNamespace() string
	// GetType() returns the serving job type
	GetType() types.ServingType
	// GetVersion() returns the serving job version
	GetVersion() string
	// return all pods which are belong to the serving job
	GetAllPods() []v1.Pod
	// get all pods names
	GetAllPodNames() []string
	// get all services which are belong to the serving job
	GetAllServices() []v1.Service
	// return the cluster ip of job service
	GetClusterIP() string
	// return the port of job service listening
	GetPorts() string
	// return available pods for the job
	AvailableInstances() int32
	// return desired pods for the job
	DesiredInstances() int32
	// return the duration time of the job is running
	GetAge() string
	// return the job status
	GetStatus() string
	// return a pod name
	GetTargetPod(podName string) (string, error)
	// check pod is in job 
	PodInJob(podName string) bool
	// check the job type is matched the given type
	IsMatchedTargetType(typ string) bool
	// check the job version is matched the given version
	IsMatchedTargetVersion(version string) bool
	// check the job namespace is matched the given namespace
	IsMatchedTargetNamespace(ns string) bool
	// check the job name is matched the given name
	IsMatchedTargetName(name string) bool
	// printer is used to print the job information
	Printer
}

type servingJobImpl struct {
	name      string
	namespace string
	serveType types.ServingType
	version   string
	pods      []v1.Pod
	svcs      []v1.Service
	deploy    app_v1.Deployment
	printer   Printer
}

func NewServingJob(deploy app_v1.Deployment, pods []v1.Pod, svcs []v1.Service) ServingJob {
	job := &servingJobImpl{}
	job.namespace = deploy.Namespace
	job.version = deploy.Labels["servingVersion"]
	job.name = deploy.Labels["servingName"]
	job.serveType = types.KeyMapServingType(deploy.Labels["servingType"])
	job.pods = []v1.Pod{}
	for _, pod := range pods {
		// if pod is in job,add it to the job.pods
		if IsPodControllerByDeploment(pod, deploy) {
			job.pods = append(job.pods, pod)
		}
	}
	job.svcs = []v1.Service{}
	for _, svc := range svcs {
		switch {
		case svc.Labels["servingName"] != job.name:
		case types.KeyMapServingType(svc.Labels["servingType"]) != job.serveType:
		case svc.Labels["servingVersion"] != job.version:
			continue
		}
		// if service belongs to the job,add it
		job.svcs = append(job.svcs, svc)
	}
	job.printer = NewServingJobPrinter(job)
	return job
}

// print help information,eg: there is more than one pod and user does not assign pod name
// we will print the help information
func (s *servingJobImpl) GetHelpInfo(obj ...interface{}) (string, error) {
	return s.printer.GetHelpInfo()
}
// print the job information with wide format
func (s *servingJobImpl) GetWidePrintString() (string, error) {
	return s.printer.GetWidePrintString()
}
// print the job information with yaml format 
func (s *servingJobImpl) GetYamlPrintString() (string, error) {
	return s.printer.GetYamlPrintString()
}
// print the job information with json format
func (s *servingJobImpl) GetJsonPrintString() (string, error) {
	return s.printer.GetJsonPrintString()
}

// check pod is existed in job
func (s *servingJobImpl) PodInJob(podName string) bool {
	for _, pod := range s.pods {
		if podName == pod.Name {
			return true
		}
	}
	return false
}
// check the job name is matched the given name,if given name is null,
// we see it as matched
func (s *servingJobImpl) IsMatchedTargetName(name string) bool {
	if name == "" {
		return true
	}
	return s.name == name
}
// check the job namespace is matched the given namespace,if given namespace is null or is AllNamespace
// we see it as matched 
func (s *servingJobImpl) IsMatchedTargetNamespace(ns string) bool {
	if ns == "" || ns == metav1.NamespaceAll {
		return true
	}
	return s.namespace == ns

}
// check the job version is matched the given version,if the given version is null,we
// see it as matched
func (s *servingJobImpl) IsMatchedTargetVersion(version string) bool {
	if version == "" {
		return true
	}
	return s.version == version
}
// check the job type is matched the given type
func (s *servingJobImpl) IsMatchedTargetType(typ string) bool {
	if typ == "" {
		return true
	}
	return types.KeyMapServingType(typ) == s.serveType
}

// this function is used to check the pod is we want or not, 
// 1.if given pod name is null and there is only one pod in job,we will return the pod
// 2.if given pod name is not null,check the pod is in job or not
func (s *servingJobImpl) GetTargetPod(podName string) (string, error) {
	if podName == "" {
		// if there is only one pod in job,we return the pod name
		if len(s.pods) == 1 {
			return s.pods[0].Name,nil
		}
		// return too many pods error,this will tell user to pick one
		return "", types.ErrTooManyPods
	}
	for _, pod := range s.pods {
		// if given pod is matched the pod in job,return true,we think the pod name given by user is ok
		if podName == pod.Name {
			return podName, nil
		}
	}
	// user gives an error pod name
	return "", types.ErrNotFoundTargetPod
}

// GetAge returns the time string for serving job is running
func (s *servingJobImpl) GetAge() string {
	return util.ShortHumanDuration(time.Now().Sub(s.deploy.ObjectMeta.CreationTimestamp.Time))
}

func (s *servingJobImpl) GetStatus() string {
	for _, pod := range s.pods {
		if pod.Status.Phase == v1.PodPending {
			return "PENDING"
		}
	}
	return "RUNNING"
}

func (s *servingJobImpl) DesiredInstances() int32 {
	return s.deploy.Status.Replicas
}

func (s *servingJobImpl) AvailableInstances() int32 {
	return s.deploy.Status.AvailableReplicas
}

func (s *servingJobImpl) GetPorts() string {
	portList := []string{}
	if len(s.svcs) > 0 {
		svc := s.svcs[0]
		ports := svc.Spec.Ports
		for _, port := range ports {
			portList = append(portList, fmt.Sprintf("%s:%d", port.Name, port.Port))
		}
		return strings.Join(portList, ",")
	}
	return "N/A"
}

func (s *servingJobImpl) GetClusterIP() string {
	if len(s.svcs) > 0 {
		return s.svcs[0].Spec.ClusterIP
	}
	return "N/A"
}

func (s *servingJobImpl) GetAllServices() []v1.Service {
	return s.svcs
}

func (s *servingJobImpl) GetAllPodNames() []string {
	names := []string{}
	for _, pod := range s.pods {
		names = append(names, pod.Name)
	}
	return names
}

func (s *servingJobImpl) GetAllPods() []v1.Pod {
	return s.pods
}

func (s *servingJobImpl) GetVersion() string {
	return s.version
}
func (s *servingJobImpl) GetType() types.ServingType {
	return s.serveType
}

func (s *servingJobImpl) GetName() string {
	return s.name
}

func (s *servingJobImpl) GetNamespace() string {
	return s.namespace
}

func IsPodControllerByDeploment(pod v1.Pod, deploy app_v1.Deployment) bool {
	if len(pod.OwnerReferences) == 0 {
		return false
	}
	podLabel := pod.GetLabels()
	if len(podLabel) == 0 {
		return false
	}
	for key, value := range deploy.Spec.Selector.MatchLabels {
		if podLabel[key] != value {
			return false
		}
	}
	return true
}
