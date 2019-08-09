package types

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	app_v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
)

type Serving struct {
	Name      string      `yaml:"name" json:"name"`
	Namespace string      `yaml:"namespace" json:"namespace"`
	ServeType ServingType `yaml:"serving_type" json:"serving_type"`
	Version   string      `yaml:"version" json:"version"`
	pods      []v1.Pod
	svcs      []v1.Service
	deploy    app_v1.Deployment
	client    *kubernetes.Clientset
}

func NewServingJob(client *kubernetes.Clientset, deploy app_v1.Deployment, allPods []v1.Pod) Serving {
	servingTypeLabel := deploy.Labels["servingType"]
	servingVersion := deploy.Labels["servingVersion"]
	servingName := deploy.Labels["servingName"]
	servingType := ServingTF
	if stype := KeyMapServingType(servingTypeLabel); stype != ServingType("") {
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
				service.Labels["servingVersion"] == s.Version {
				// return service.Spec.ClusterIP
				svcs = append(svcs, service)
			}
		}
		s.svcs = svcs
	}
	return s.svcs
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
func DefinePodPhaseStatus(pod v1.Pod) (string, int, int, int) {
	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}
	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			reason = "Running"
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}
	return reason, totalContainers, restarts, readyContainers
}
