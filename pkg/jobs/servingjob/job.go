package serving

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	app_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServingJob interface {
	GetName() string
	GetNamespace() string
	GetType() types.ServingType
	GetVersion() string
	GetAllPods() []v1.Pod
	GetAllPodNames() []string
	GetAllServices() []v1.Service
	GetInfoByJsonString() string
	GetClusterIP() string
	GetPorts() string
	AvailableInstances() int32
	DesiredInstances() int32
	GetAge() string
	GetStatus() string
	GetTargetPod(podName string) (string, error)
	PodInJob(podName string) bool
	IsMatchedTargetType(typ string) bool
	IsMatchedTargetVersion(version string) bool
	IsMatchedTargetNamespace(ns string) bool
	IsMatchedTargetName(name string) bool
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
		job.svcs = append(job.svcs, svc)
	}
	job.printer = NewServingJobPrinter(job)
	return job
}

func (s *servingJobImpl) GetHelpInfo(obj ...interface{}) (string, error) {
	return s.printer.GetHelpInfo()
}

func (s *servingJobImpl) GetWidePrintString() (string, error) {
	return s.printer.GetWidePrintString()
}

func (s *servingJobImpl) GetYamlPrintString() (string, error) {
	return s.printer.GetYamlPrintString()
}

func (s *servingJobImpl) GetJsonPrintString() (string, error) {
	return s.printer.GetJsonPrintString()
}

func (s *servingJobImpl) PodInJob(podName string) bool {
	for _, pod := range s.pods {
		if podName == pod.Name {
			return true
		}
	}
	return false
}

func (s *servingJobImpl) IsMatchedTargetName(name string) bool {
	if name == "" {
		return true
	}
	return s.name == name
}

func (s *servingJobImpl) IsMatchedTargetNamespace(ns string) bool {
	if ns == "" || ns == metav1.NamespaceAll {
		return true
	}
	return s.namespace == ns

}
func (s *servingJobImpl) IsMatchedTargetVersion(version string) bool {
	if version == "" {
		return true
	}
	return s.version == version
}
func (s *servingJobImpl) IsMatchedTargetType(typ string) bool {
	if typ == "" {
		return true
	}
	return types.KeyMapServingType(typ) == s.serveType
}

func (s *servingJobImpl) GetTargetPod(podName string) (string, error) {
	if podName == "" {
		if len(s.pods) == 1 {
			return s.pods[0].Name,nil
		}
		return "", types.ErrTooManyPods
	}
	for _, pod := range s.pods {
		if podName == pod.Name {
			return podName, nil
		}
	}
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

func (s *servingJobImpl) GetInfoByJsonString() string {
	info := map[string]string{}
	info["name"] = s.name
	info["namespace"] = s.namespace
	info["version"] = s.version
	info["serving_type"] = string(s.serveType)
	jsonBytes, _ := json.Marshal(info)
	return string(jsonBytes)
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
