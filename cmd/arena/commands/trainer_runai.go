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
	"k8s.io/apimachinery/pkg/types"
)

type RunaiTrainer struct {
	client kubernetes.Interface
}

func NewRunaiTrainer(client kubernetes.Interface) Trainer {
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

	filteredPods := []v1.Pod{}
	for _, pod := range podList.Items {
		if pod.OwnerReferences[0].UID == replicaSet.UID{
			filteredPods = append(filteredPods, pod)
		}
	}

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	lastCreatedPod := getLastCreatedPod(filteredPods)
	ownerResource := Resource{
		Uid: string(replicaSet.UID),
		ResourceType: ResourceTypeReplicaset,
		Name: replicaSet.Name,
	}
	return NewRunaiJob(filteredPods, lastCreatedPod, replicaSet.CreationTimestamp, rt.Type(), replicaSet.Name, false, replicaSet.Labels["app"] == "runai", []string{},false, replicaSet.Spec.Template.Spec, replicaSet.Spec.Template.ObjectMeta, namespace, ownerResource), nil
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

	filteredPods := []v1.Pod{}
	for _, pod := range podList.Items {
		if pod.OwnerReferences[0].UID == statefulset.UID{
			filteredPods = append(filteredPods, pod)
		}
	}

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	lastCreatedPod := getLastCreatedPod(filteredPods)
	ownerResource := Resource{
		Uid: string(statefulset.UID),
		ResourceType: ResourceTypeStatefulSet,
		Name: statefulset.Name,
	}
	return NewRunaiJob(filteredPods, lastCreatedPod, statefulset.CreationTimestamp, rt.Type(), statefulset.Name, true, statefulset.Labels["app"] == "runai", []string{},false, statefulset.Spec.Template.Spec, statefulset.Spec.Template.ObjectMeta, namespace, ownerResource), nil
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

	filteredPods := []v1.Pod{}
	for _, pod := range podList.Items {
		if pod.OwnerReferences[0].UID == job.UID{
			filteredPods = append(filteredPods, pod)
		}
	}

	// Last created pod will be the chief pod
	lastCreatedPod := getLastCreatedPod(filteredPods)
	ownerResource := Resource{
		Uid: string(job.UID),
		ResourceType: ResourceTypeJob,
		Name: job.Name,
	}
	return NewRunaiJob(filteredPods, lastCreatedPod, job.CreationTimestamp, rt.Type(), job.Name, true, job.Labels["app"] == "runai", []string{}, false, job.Spec.Template.Spec, job.Spec.Template.ObjectMeta, namespace, ownerResource), nil
}

type RunaiJobInfo struct {
	name              string
	creationTimestamp metav1.Time
	pods              []v1.Pod
	createdByCLI      bool
	interactive       bool
	deleted	bool
	podSpec v1.PodSpec
	podMetadata metav1.ObjectMeta
	owner Resource
}

type RunaiOwnerInfo struct {
	Name string
	Type string
	Uid string
}

func (rt *RunaiTrainer) ListTrainingJobs(namespace string) ([]TrainingJob, error) {
	runaiJobs := []TrainingJob{}
	services, err := rt.getServicesInNamespace(namespace)

	if err != nil {
		return nil, err
	}

	nodeIp, err := rt.getNodeIp()

	if err != nil {
		return nil, err
	}

	ingressService, err := rt.getIngressService()
	
	if err != nil {
		return nil, err
	}

	ingresses, err := rt.getIngressesForNamespace(namespace)

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

	jobPodMap := make(map[types.UID]*RunaiJobInfo)

	// Group the pods by their controller
	for _, pod := range runaiPods.Items {
		controller := ""
		var uid types.UID = ""

		for _, owner := range pod.OwnerReferences {
			if owner.Controller != nil && *owner.Controller {
				controller = owner.Name
				uid = owner.UID
			}
		}

		if jobPodMap[uid] == nil {
			jobPodMap[uid] = &RunaiJobInfo{
				name: controller,
				pods: []v1.Pod{},
				// Mark all jobs as deleted unless we find them at the next stage
				deleted: true,
				podSpec: pod.Spec,
				podMetadata: pod.ObjectMeta,
			}
		}

		// If controller exists for pod than add it to the map
		if controller != "" {
			jobPodMap[uid].pods = append(jobPodMap[uid].pods, pod)
		}
	}

	// Find more info on each of the controllers
	runaiJobList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{})

	for _, job := range runaiJobList.Items {
		var jobInfo *RunaiJobInfo
		if jobPodMap[job.UID] != nil {
			jobInfo = jobPodMap[job.UID]
		} else {
			// Create the job even if it does not have any pods currently
			jobInfo = &RunaiJobInfo{}
			jobPodMap[job.UID] = jobInfo
			jobInfo.name = job.Name
			jobInfo.podSpec = job.Spec.Template.Spec
			jobInfo.podMetadata = job.Spec.Template.ObjectMeta
		}

		jobInfo.creationTimestamp = job.CreationTimestamp
		jobInfo.deleted = false
		jobInfo.owner = Resource{
			Name:  job.Name,
			ResourceType: ResourceTypeJob,
			Uid: string(job.UID),
		}

		if job.Labels["app"] == "runaijob" {
			jobInfo.createdByCLI = true
		}
	}

	runaiStatefulSetsList, err := rt.client.Apps().StatefulSets(namespace).List(metav1.ListOptions{})

	for _, statefulSet := range runaiStatefulSetsList.Items {
		var jobInfo *RunaiJobInfo
		if jobPodMap[statefulSet.UID] != nil {
			jobInfo = jobPodMap[statefulSet.UID]
		} else {
			// Create the job even if it does not have any pods currently
			jobInfo = &RunaiJobInfo{}
			jobPodMap[statefulSet.UID] = jobInfo
			jobInfo.name = statefulSet.Name
			jobInfo.podSpec = statefulSet.Spec.Template.Spec
			jobInfo.podMetadata = statefulSet.Spec.Template.ObjectMeta
		}
		jobInfo.creationTimestamp = statefulSet.CreationTimestamp
		jobInfo.interactive = true
		jobInfo.deleted = false
		jobInfo.owner = Resource{
			Name:  statefulSet.Name,
			ResourceType: ResourceTypeStatefulSet,
			Uid: string(statefulSet.UID),
		}

		if statefulSet.Labels["app"] == "runaijob" {
			jobInfo.createdByCLI = true
		}
	}

	runaiReplicaSetsList, err := rt.client.Apps().ReplicaSets(namespace).List(metav1.ListOptions{})

	for _, replicaSet := range runaiReplicaSetsList.Items {
		var jobInfo *RunaiJobInfo
		if	jobPodMap[replicaSet.UID] != nil {
			jobInfo = jobPodMap[replicaSet.UID]
		} else {
			// Create the job even if it does not have any pods currently
			jobInfo = &RunaiJobInfo{}
			jobPodMap[replicaSet.UID] = jobInfo
			jobInfo.name = replicaSet.Name
			jobInfo.podSpec = replicaSet.Spec.Template.Spec
			jobInfo.podMetadata = replicaSet.Spec.Template.ObjectMeta
		}
		jobInfo.creationTimestamp = replicaSet.CreationTimestamp
		jobInfo.interactive = true
		jobInfo.deleted = false
		jobInfo.owner = Resource{
			Name:  replicaSet.Name,
			ResourceType: ResourceTypeReplicaset,
			Uid: string(replicaSet.UID),
		}

		if replicaSet.Labels["app"] == "runaijob" {
			jobInfo.createdByCLI = true
		}
	}

	for _, jobInfo := range jobPodMap {
		lastCreatedPod := getLastCreatedPod(jobInfo.pods)

		serviceUrls := []string{}
		if lastCreatedPod != nil {
			serviceOfPod := getServiceOfPod(services, lastCreatedPod)
			if serviceOfPod != nil {
				serviceUrls = getServiceUrls(ingressService, ingresses, nodeIp, *serviceOfPod)
			}
		}

		runaiJobs = append(runaiJobs, NewRunaiJob(jobInfo.pods, lastCreatedPod, jobInfo.creationTimestamp, "runai", jobInfo.name, jobInfo.interactive, jobInfo.createdByCLI, serviceUrls, jobInfo.deleted, jobInfo.podSpec, jobInfo.podMetadata, namespace, jobInfo.owner))
	}

	return runaiJobs, nil
}

func (rt *RunaiTrainer) getNodeIp() (string, error) {
	nodesList, err := rt.client.Core().Nodes().List(metav1.ListOptions{})

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
	if len(pods) == 0 {
		 return nil
	}
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

func (rt *RunaiTrainer) getServicesInNamespace(namespace string) ([]v1.Service, error) {
	servicesList, err := rt.client.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return []v1.Service{}, err
	}
	return servicesList.Items, nil
}

func (rt *RunaiTrainer) getIngressesForNamespace(namespace string) ([]extensionsv1.Ingress, error){
	ingresses, err := rt.client.ExtensionsV1beta1().Ingresses(namespace).List(metav1.ListOptions{
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

func (rt *RunaiTrainer) getIngressService() (*v1.Service, error) {
	servicesList, err := rt.client.Core().Services("").List(metav1.ListOptions{
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
