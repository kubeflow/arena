package serving

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	app_v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
)

type Serving struct {
	Name      string            `yaml:"name" json:"name"`
	Namespace string            `yaml:"namespace" json:"namespace"`
	ServeType types.ServingType `yaml:"serving_type" json:"serving_type"`
	Version   string            `yaml:"version" json:"version"`
	pods      []v1.Pod
	svcs      []v1.Service
	deploy    app_v1.Deployment
	client    *kubernetes.Clientset
}

func NewServingJob(client *kubernetes.Clientset, deploy app_v1.Deployment, allPods []v1.Pod) Serving {
	servingTypeLabel := deploy.Labels["servingType"]
	servingVersion := deploy.Labels["servingVersion"]
	servingName := deploy.Labels["servingName"]
	servingType := types.ServingTF
	if stype := KeyMapServingType(servingTypeLabel); stype != types.ServingType("") {
		servingType = stype
	}
	serving := Serving{
		Name:      servingName,
		client:    client,
		ServeType: servingType,
		Namespace: deploy.Namespace,
		Version:   servingVersion,
		deploy:    deploy,
	}
	for _, pod := range allPods {
		if IsPodControllerByDeploment(pod, deploy) {
			serving.pods = append(serving.pods, pod)
		}
	}
	return serving
}

func (s Serving) GetName() string {
	return s.Name
}

func (s Serving) AllPods() []v1.Pod {
	return s.pods
}

func (s Serving) AllSvcs() (svcs []v1.Service) {
	svcs = []v1.Service{}
	if len(s.svcs) == 0 {
		allServices, err := util.AcquireServingServices(s.Namespace, s.client)
		if err != nil {
			log.Errorf("failed to list serving services, err: %++v", err)
			return svcs
		}

		log.Debugf("try to get Endpoint IP for name %s and %s", s.Name, s.Version)
		for _, service := range allServices {
			if service.Labels["servingName"] == s.Name &&
				KeyMapServingType(service.Labels["servingType"]) == s.ServeType &&
				service.Labels["servingVersion"] == s.Version {
				// return service.Spec.ClusterIP
				svcs = append(svcs, service)
			}
		}
		s.svcs = svcs
	}
	return s.svcs
}

/** Get the endpoint IP
 * Try to get load balancer id first, if failed, turn to
 */
func (s Serving) GetEndpointIP() string {

	if len(s.AllSvcs()) > 0 {
		service := s.AllSvcs()[0]

		// 1.Get Address for loadbalancer
		if service.Spec.Type == v1.ServiceTypeLoadBalancer {
			if len(service.Status.LoadBalancer.Ingress) > 0 {
				return service.Status.LoadBalancer.Ingress[0].IP
			}
		}

		// 2.Get SVC endpoint address
		return service.Spec.ClusterIP
	}

	return "N/A"
}

// Cluster IP
func (s Serving) GetClusterIP() string {
	if len(s.AllSvcs()) > 0 {
		return s.AllSvcs()[0].Spec.ClusterIP
	}

	return "N/A"
}

func (s Serving) GetPorts() string {
	portList := []string{}

	if len(s.AllSvcs()) > 0 {
		// return s.AllSvcs().Spec.ClusterIP
		svc := s.AllSvcs()[0]
		ports := svc.Spec.Ports
		for _, port := range ports {
			portList = append(portList, fmt.Sprintf("%s:%d", port.Name, port.Port))
		}
		return strings.Join(portList, ",")
	}

	return "N/A"
}

// Available instances
func (s Serving) AvailableInstances() int32 {
	return s.deploy.Status.AvailableReplicas
}

// Desired Instances
func (s Serving) DesiredInstances() int32 {
	return s.deploy.Status.Replicas
}

func (s Serving) GetStatus() string {
	for _, pod := range s.pods {
		if pod.Status.Phase == v1.PodPending {
			log.Debugf("pod %s is pending", pod.Name)
			return "PENDING"
		}
	}
	return "RUNNING"
}

// GetAge returns the time string for serving job is running
func (s Serving) GetAge() string {
	return util.ShortHumanDuration(time.Now().Sub(s.deploy.ObjectMeta.CreationTimestamp.Time))
}

func (s Serving) IsMatchedGivenCondition(target string, targetType string) bool {
	switch {
	case target == "":
		return true
	case targetType == "NAMESPACE" && target == s.Namespace:
		return true
	case targetType == "VERSION" && target == s.Version:
		return true
	case targetType == "TYPE" && KeyMapServingType(target) == s.ServeType:
		return true
	default:
		return false
	}
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
