package commands

import
(
	v1 "k8s.io/api/core/v1"

	log "github.com/sirupsen/logrus"
	app_v1 "k8s.io/api/apps/v1"
	"fmt"
)

type ServingJob struct {
	Name string
	ServeType string
	Namespace string
	Version string
	pods []v1.Pod
	deploy app_v1.Deployment
}


func NewServingJob(deploy app_v1.Deployment) ServingJob {
	chart := deploy.Labels["chart"]
	serviceVersion := deploy.Labels["serviceVersion"]
	servingName := deploy.Labels["servingName"]
	servingType := "Tensorflow"
	if serveType, ok := serving_charts[chart]; ok {
		servingType = serveType
	}
	job := ServingJob{
		Name: servingName,
		ServeType: servingType,
		Namespace: deploy.Namespace,
		Version: serviceVersion,
		deploy: deploy,
	}
	for _, pod := range allPods {
		if IsPodControllerByDeploment(pod, deploy) {
			job.pods = append(job.pods, pod)
		}
	}
	return job
}

func (s ServingJob) GetName() string {
	if s.Version != "" {
		return fmt.Sprintf("%s-%s", s.Name, s.Version)
	}
	return s.Name
}

func (s ServingJob) AllPods() []v1.Pod {
	return s.pods
}

func (s ServingJob) GetClusterIP() string {
	serviceName := fmt.Sprintf("%s-%s", s.deploy.Labels["release"], s.deploy.Labels["app"])
	for _, service := range allServices {
		if service.Name == serviceName {
			return service.Spec.ClusterIP
		}
	}
	return "N/A"
}

func (s ServingJob) GetStatus() string {
	hasPendingPod := false
	for _, pod := range s.pods {
		if pod.Status.Phase == v1.PodPending {
			log.Debugf("pod %s is pending", pod.Name)
			hasPendingPod = true
			break
		}
		if hasPendingPod {
			return "PENDING"
		}
	}
	return "RUNNING"
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