package commands

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type RunaiTrainer struct {
	client *kubernetes.Clientset
}

func NewRunaiTrainer(client *kubernetes.Clientset) Trainer {
	return &RunaiTrainer{
		client: client,
	}
}

func fieldSelectorByName(name string) string {
	return fmt.Sprintf("metadata.name=%s", name)
}

func (rt *RunaiTrainer) IsSupported(name, ns string) bool {
	runaiJobList, err := rt.client.Batch().Jobs(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiJobList.Items) > 0 {
		for _, item := range runaiJobList.Items {
			if item.Spec.Template.Spec.SchedulerName == "runai-scheduler" {
				return true
			}
		}
	}

	runaiStatefulSetsList, err := rt.client.Apps().StatefulSets(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiStatefulSetsList.Items) > 0 {
		for _, item := range runaiStatefulSetsList.Items {
			if item.Spec.Template.Spec.SchedulerName == "runai-scheduler" {
				return true
			}
		}
	}

	runaiReplicaSetsList, err := rt.client.Apps().ReplicaSets(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiReplicaSetsList.Items) > 0 {
		for _, item := range runaiReplicaSetsList.Items {
			if item.Spec.Template.Spec.SchedulerName == "runai-scheduler" {
				return true
			}
		}
	}

	return false
}

func (rt *RunaiTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {

	runaiJobList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiJobList.Items) > 0 {
		return rt.getTrainingJob(runaiJobList.Items[0])
	}

	runaiStatufulsetList, err := rt.client.Apps().StatefulSets(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiStatufulsetList.Items) > 0 {
		return rt.getTrainingStatefulset(runaiStatufulsetList.Items[0])
	}

	runaiReplicaSetsList, err := rt.client.Apps().ReplicaSets(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiReplicaSetsList.Items) > 0 {
		return rt.getTrainingReplicaSet(runaiReplicaSetsList.Items[0])
	}

	return nil, fmt.Errorf("Failed to find the job for %s", name)
}

func (rt *RunaiTrainer) Type() string {
	return defaultRunaiTrainingType
}

func (rt *RunaiTrainer) getTrainingReplicaSet(replicaSet appsv1.ReplicaSet) (TrainingJob, error) {
	labels := []string{}
	for key, value := range replicaSet.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: strings.Join(labels, ","),
	})

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	pods := podList.Items
	lastCreatedPod := getLastCreatedPod(pods)
	return NewRunaiJob(pods, *lastCreatedPod, replicaSet.CreationTimestamp, rt.Type(), replicaSet.Name, false, replicaSet.Labels["app"] == "runai", []string{}), nil
}

func (rt *RunaiTrainer) getTrainingStatefulset(statefulset appsv1.StatefulSet) (TrainingJob, error) {
	labels := []string{}
	for key, value := range statefulset.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: strings.Join(labels, ","),
	})

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	pods := podList.Items
	lastCreatedPod := getLastCreatedPod(pods)
	return NewRunaiJob(pods, *lastCreatedPod, statefulset.CreationTimestamp, rt.Type(), statefulset.Name, true, statefulset.Labels["app"] == "runai", []string{}), nil
}

func (rt *RunaiTrainer) getTrainingJob(job batchv1.Job) (TrainingJob, error) {
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
	lastCreatedPod := getLastCreatedPod(pods)
	return NewRunaiJob(pods, *lastCreatedPod, job.CreationTimestamp, rt.Type(), job.Name, true, job.Labels["app"] == "runai", []string{}), nil
}

type RunaiJobInfo struct {
	name              string
	kind              string
	creationTimestamp metav1.Time
	pods              []v1.Pod
	createdByCLI      bool
	interactive       bool
}

func (rt *RunaiTrainer) ListTrainingJobs() ([]TrainingJob, error) {
	runaiJobs := []TrainingJob{}
	services, err := getServicesInNamespace(namespace)

	if err != nil {
		return nil, err
	}

	nodeIp, err := getNodeIp()

	if err != nil {
		return nil, err
	}

	ingressService, err := getIngressService()
	
	if err != nil {
		return nil, err
	}

	ingresses, err := getIngressesForNamespace(namespace)

	if err != nil {
		return nil, err
	}

	// Get all pods running with runai scheduler
	runaiPods, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		FieldSelector: "spec.schedulerName=runai-scheduler",
	})

	if err != nil {
		return nil, err
	}

	jobPodMap := make(map[string]*RunaiJobInfo)

	// Group the pods by their controller
	for _, pod := range runaiPods.Items {
		controller := ""
		kind := ""

		for _, owner := range pod.OwnerReferences {
			if *owner.Controller {
				controller = owner.Name
				kind = owner.Kind
			}
		}

		if jobPodMap[controller] == nil {
			jobPodMap[controller] = &RunaiJobInfo{
				name: controller,
				pods: []v1.Pod{},
				kind: kind,
			}
		}

		// If controller exists for pod than add it to the map
		if controller != "" {
			jobPodMap[controller].pods = append(jobPodMap[controller].pods, pod)
		}
	}

	// Find more info on each of the controllers
	runaiJobList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{})

	for _, job := range runaiJobList.Items {
		if jobPodMap[job.Name] != nil {
			jobPodMap[job.Name].creationTimestamp = job.CreationTimestamp

			if job.Labels["app"] == "runaijob" {
				jobPodMap[job.Name].createdByCLI = true
			}
		}
	}

	runaiStatefulSetsList, err := rt.client.Apps().StatefulSets(namespace).List(metav1.ListOptions{})

	for _, statefulSet := range runaiStatefulSetsList.Items {
		if jobPodMap[statefulSet.Name] != nil {
			jobPodMap[statefulSet.Name].creationTimestamp = statefulSet.CreationTimestamp
			jobPodMap[statefulSet.Name].interactive = true

			if statefulSet.Labels["app"] == "runaijob" {
				jobPodMap[statefulSet.Name].createdByCLI = true
			}
		}
	}

	runaiReplicaSetsList, err := rt.client.Apps().ReplicaSets(namespace).List(metav1.ListOptions{})

	for _, replicaSet := range runaiReplicaSetsList.Items {
		if jobPodMap[replicaSet.Name] != nil {
			jobPodMap[replicaSet.Name].creationTimestamp = replicaSet.CreationTimestamp
			jobPodMap[replicaSet.Name].interactive = true

			if replicaSet.Labels["app"] == "runaijob" {
				jobPodMap[replicaSet.Name].createdByCLI = true
			}
		}
	}

	for _, jobInfo := range jobPodMap {
		lastCreatedPod := getLastCreatedPod(jobInfo.pods)
		serviceOfPod := getServiceOfPod(services, lastCreatedPod)

		serviceUrls := []string{}
		if serviceOfPod != nil {
			serviceUrls = getServiceUrls(ingressService, ingresses, nodeIp, *serviceOfPod)
		}

		runaiJobs = append(runaiJobs, NewRunaiJob(jobInfo.pods, *lastCreatedPod, jobInfo.creationTimestamp, "runai", jobInfo.name, jobInfo.interactive, jobInfo.createdByCLI, serviceUrls))
	}

	return runaiJobs, nil
}

func getNodeIp() (string, error) {
	nodesList, err := clientset.Core().Nodes().List(metav1.ListOptions{})

	if err != nil {
		return "", err
	}

	if len(nodesList.Items) != 0 {
		for _, node := range nodesList.Items {
			addresses := node.Status.Addresses
			for _, address := range addresses {
				if address.Type == v1.NodeInternalIP {
					return address.Address, nil
				}
			}
		}
	}

	return "", nil
}

func getServiceEndpoints(nodeIp string, service v1.Service) []string{
	if service.Status.LoadBalancer.Ingress != nil && len(service.Status.LoadBalancer.Ingress) != 0 {
		urls := []string{}
		for _, port := range service.Spec.Ports {
			serviceIp := service.Status.LoadBalancer.Ingress[0].IP
			var url string
			if port.Port == 80 {
				url = fmt.Sprintf("http://%s", serviceIp)
			} else if port.Port == 443 {
				url = fmt.Sprintf("https://%s", serviceIp)
			} else {
				url = fmt.Sprintf("http://%s:%d", serviceIp, port.Port)
			}
			urls = append(urls, url)
		}

		return urls
	}

	if service.Spec.Type == v1.ServiceTypeLoadBalancer {
		return []string{"<pending>"}
	}

	if service.Spec.Type == v1.ServiceTypeNodePort {
		urls := []string{}
		for _, port := range service.Spec.Ports {
			urls = append(urls, fmt.Sprintf("http://%s:%d", nodeIp, port.NodePort))
		}

		return urls
	}

	return []string{}
}

func getServiceUrls(ingressService *v1.Service, ingresses []extensionsv1.Ingress, nodeIp string, service v1.Service) []string{
	ingressEndpoints := []string{}
	if ingressService != nil {
		ingressEndpoints = getServiceEndpoints(nodeIp, *ingressService) 
	}

	if service.Spec.Type == v1.ServiceTypeNodePort || service.Spec.Type == v1.ServiceTypeLoadBalancer {
		return getServiceEndpoints(nodeIp, service)
	} else {
		urls := []string{}
		for _, servicePortConfig := range service.Spec.Ports {
			servicePort := servicePortConfig.Port
			ingressPathForService := getIngressPathOfService(ingresses, service, servicePort)

			// No path specified
			if ingressPathForService == nil {
				continue;
			}

			if len(ingressEndpoints) > 0 && ingressEndpoints[0] == "<pending>" {
				return []string{"<pending>"}
			}
			for _, ingressEndpoint := range ingressEndpoints {
				urls = append(urls, fmt.Sprintf("%s%s", ingressEndpoint, *ingressPathForService))
			}	
		}

		return urls
	}

	return []string{}
}

func getLastCreatedPod(pods []v1.Pod) *v1.Pod {
	lastCreatedPod := pods[0]
	otherPods := pods[1:]
	for _, item := range otherPods {
		if lastCreatedPod.CreationTimestamp.Before(&item.CreationTimestamp) {
			lastCreatedPod = item
		}
	}

	return &lastCreatedPod
}

func getServiceOfPod(services []v1.Service, pod *v1.Pod) *v1.Service {
	for _, service := range services {

		if service.Spec.Selector == nil {
			continue
		}

		match := true
		for key, value := range service.Spec.Selector {
			if pod.Labels[key] != value {
				match = false
			}
		}

		if match {
			return &service
		}
	}

	return nil
}

func getServicesInNamespace(namespace string) ([]v1.Service, error) {
	servicesList, err := clientset.Core().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return []v1.Service{}, err
	}
	return servicesList.Items, nil
}

func getIngressesForNamespace(namespace string) ([]extensionsv1.Ingress, error){
	ingresses, err := clientset.ExtensionsV1beta1().Ingresses(namespace).List(metav1.ListOptions{
	})

	if err != nil {
		return []extensionsv1.Ingress{}, nil
	}

	ngnixIngresses := []extensionsv1.Ingress{}
	for _, ingress := range ingresses.Items {

		// Support only ngnix ingresses
		if ingress.Annotations["kubernetes.io/ingress.class"] == "nginx" {
			ngnixIngresses = append(ngnixIngresses, ingress)
		}
	}

	return ngnixIngresses, nil
}

func getIngressPathOfService(ingresses []extensionsv1.Ingress, service v1.Service, port int32) *string{
	var ingressPath string

	for _, ingress := range ingresses {
		rules := ingress.Spec.Rules
		for _, rule := range rules {
			paths := rule.HTTP.Paths
			for _, path := range paths {
				if path.Backend.ServiceName == service.Name && path.Backend.ServicePort.IntVal == port {
					ingressPath = path.Path
					return &ingressPath
				}
			}
		}
	}

	return nil
}

func getIngressService() (*v1.Service, error) {
	servicesList, err := clientset.Core().Services("").List(metav1.ListOptions{
		LabelSelector: "app=nginx-ingress",
	})

	if err != nil {
		return nil, err
	}

	if len(servicesList.Items) > 0 {
		return &servicesList.Items[0], nil
	}

	return nil, nil
}
