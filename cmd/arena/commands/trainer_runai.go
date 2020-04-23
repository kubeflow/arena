package commands

import (
	"fmt"
	"strings"

	cmdTypes "github.com/kubeflow/arena/cmd/arena/types"
	"github.com/kubeflow/arena/pkg/client"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	runaiTrainType                  = "Train"
	runaiInteractiveType            = "Interactive"
	runaiPreemptibleInteractiveType = "Interactive-Preemptible"
)

type RunaiTrainer struct {
	client kubernetes.Interface
}

func NewRunaiTrainer(client client.Client) Trainer {
	return &RunaiTrainer{
		client: client.GetClientset(),
	}
}

func fieldSelectorByName(name string) string {
	return fmt.Sprintf("metadata.name=%s", name)
}

func (rt *RunaiTrainer) IsSupported(name, ns string) bool {
	runaiJobList, err := rt.client.BatchV1().Jobs(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiJobList.Items) > 0 {
		for _, item := range runaiJobList.Items {
			if item.Spec.Template.Spec.SchedulerName == SchedulerName {
				return true
			}
		}
	}

	runaiStatefulSetsList, err := rt.client.AppsV1().StatefulSets(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiStatefulSetsList.Items) > 0 {
		for _, item := range runaiStatefulSetsList.Items {
			return rt.isRunaiPodObject(item.ObjectMeta, item.Spec.Template)
		}
	}

	runaiReplicaSetsList, err := rt.client.AppsV1().ReplicaSets(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiReplicaSetsList.Items) > 0 {
		for _, item := range runaiReplicaSetsList.Items {
			if item.Spec.Template.Spec.SchedulerName == SchedulerName {
				return true
			}
		}
	}

	return false
}

func (rt *RunaiTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {

	runaiJobList, err := rt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiJobList.Items) > 0 {
		podSpecJob := cmdTypes.PodTemplateJobFromJob(runaiJobList.Items[0])
		result, err := rt.getRunaiTrainingJob(*podSpecJob, namespace)
		if err != nil {
			log.Debugf("failed to get job %s in namespace %s due to %v", name, namespace, err)
		}

		if result != nil {
			return result, nil
		}
	}

	runaiStatufulsetList, err := rt.client.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiStatufulsetList.Items) > 0 {
		podSpecJob := cmdTypes.PodTemplateJobFromStatefulSet(runaiStatufulsetList.Items[0])
		result, err := rt.getRunaiTrainingJob(*podSpecJob, namespace)
		if err != nil {
			log.Debugf("failed to get job %s in namespace %s due to %v", name, namespace, err)
		}

		if result != nil {
			return result, nil
		}
	}

	runaiReplicaSetsList, err := rt.client.AppsV1().ReplicaSets(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiReplicaSetsList.Items) > 0 {
		podSpecJob := cmdTypes.PodTemplateJobFromReplicaSet(runaiReplicaSetsList.Items[0])
		result, err := rt.getRunaiTrainingJob(*podSpecJob, namespace)
		if err != nil {
			log.Debugf("failed to get job %s in namespace %s due to %v", name, namespace, err)
		}

		if result != nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("Failed to find the job for %s", name)
}

func (rt *RunaiTrainer) Type() string {
	return defaultRunaiTrainingType
}

func (rt *RunaiTrainer) getRunaiTrainingJob(podSpecJob cmdTypes.PodTemplateJob, namespace string) (TrainingJob, error) {
	if podSpecJob.Template.Spec.SchedulerName != SchedulerName {
		return nil, nil
	}

	labels := []string{}
	for key, value := range podSpecJob.Selector.MatchLabels {
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

	filteredPods := []v1.Pod{}
	for _, pod := range podList.Items {
		if pod.OwnerReferences[0].UID == podSpecJob.UID {
			filteredPods = append(filteredPods, pod)
		}
	}

	lastCreatedPod := getLastCreatedPod(filteredPods)
	ownerResource := cmdTypes.Resource{
		Uid:          string(podSpecJob.UID),
		ResourceType: podSpecJob.Type,
		Name:         podSpecJob.Name,
	}

	jobType := rt.getJobType(&podSpecJob)
	return NewRunaiJob(filteredPods, lastCreatedPod, podSpecJob.CreationTimestamp, jobType, podSpecJob.Name, podSpecJob.Labels["app"] == "runai", []string{}, false, podSpecJob.Template.Spec, podSpecJob.Template.ObjectMeta, podSpecJob.ObjectMeta, namespace, ownerResource), nil
}

func (rt *RunaiTrainer) isRunaiPodObject(metadata metav1.ObjectMeta, template v1.PodTemplateSpec) bool {
	if template.Spec.SchedulerName != SchedulerName {
		return false
	}

	if _, ok := metadata.Labels["mpi_job_name"]; ok {
		return false
	}

	return true
}

type PodTemplateJob struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Selector *metav1.LabelSelector
	Template v1.PodTemplateSpec
}

type RunaiJobInfo struct {
	name              string
	jobType           string
	creationTimestamp metav1.Time
	pods              []v1.Pod
	createdByCLI      bool
	deleted           bool
	podSpec           v1.PodSpec
	podMetadata       metav1.ObjectMeta
	ObjectMeta        metav1.ObjectMeta
	owner             cmdTypes.Resource
}

type RunaiOwnerInfo struct {
	Name string
	Type string
	Uid  string
}

func (rt *RunaiTrainer) IsEnabled() bool {
	return true
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
		FieldSelector: fmt.Sprintf("spec.schedulerName=%s", SchedulerName),
	})

	if err != nil {
		return nil, err
	}

	jobPodMap := make(map[types.UID]*RunaiJobInfo)

	// Group the pods by their controller
	for _, pod := range runaiPods.Items {
		if IsMPIPod(pod) {
			continue
		}
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
				deleted:      true,
				podSpec:      pod.Spec,
				podMetadata:  pod.ObjectMeta,
				createdByCLI: pod.Labels["app"] == "runaijob",
			}
		}

		// If controller exists for pod than add it to the map
		if controller != "" {
			jobPodMap[uid].pods = append(jobPodMap[uid].pods, pod)
		}
	}

	// Get all different job stypes to one general job type with pod spec
	jobsForListCommand := []*cmdTypes.PodTemplateJob{}
	runaiJobList, err := rt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{})

	for _, job := range runaiJobList.Items {
		podTemplateJob := cmdTypes.PodTemplateJobFromJob(job)
		jobsForListCommand = append(jobsForListCommand, podTemplateJob)
	}

	runaiStatefulSetsList, err := rt.client.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{})

	for _, statefulSet := range runaiStatefulSetsList.Items {
		podTemplateJob := cmdTypes.PodTemplateJobFromStatefulSet(statefulSet)
		jobsForListCommand = append(jobsForListCommand, podTemplateJob)
	}

	replicasetJobs, err := rt.client.AppsV1().ReplicaSets(namespace).List(metav1.ListOptions{})

	for _, replicaSet := range replicasetJobs.Items {
		podTemplateJob := cmdTypes.PodTemplateJobFromReplicaSet(replicaSet)
		jobsForListCommand = append(jobsForListCommand, podTemplateJob)
	}

	for _, job := range jobsForListCommand {
		if !rt.isRunaiPodObject(job.ObjectMeta, job.Template) {
			continue
		}

		var jobInfo *RunaiJobInfo
		if jobPodMap[job.UID] != nil {
			jobInfo = jobPodMap[job.UID]
		} else {
			// Create the job even if it does not have any pods currently
			jobInfo = &RunaiJobInfo{}
			jobPodMap[job.UID] = jobInfo
			jobInfo.name = job.Name
			jobInfo.podSpec = job.Template.Spec
			jobInfo.podMetadata = job.Template.ObjectMeta
		}

		jobInfo.ObjectMeta = job.ObjectMeta
		jobInfo.creationTimestamp = job.CreationTimestamp
		jobInfo.deleted = false
		jobInfo.owner = cmdTypes.Resource{
			Name:         job.Name,
			ResourceType: job.Type,
			Uid:          string(job.UID),
		}

		if job.Labels["app"] == "runaijob" {
			jobInfo.createdByCLI = true
		}

		jobInfo.jobType = rt.getJobType(job)
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

		runaiJobs = append(runaiJobs, NewRunaiJob(jobInfo.pods, lastCreatedPod, jobInfo.creationTimestamp, jobInfo.jobType, jobInfo.name, jobInfo.createdByCLI, serviceUrls, jobInfo.deleted, jobInfo.podSpec, jobInfo.podMetadata, jobInfo.ObjectMeta, namespace, jobInfo.owner))
	}

	return runaiJobs, nil
}

func (rt *RunaiTrainer) getJobType(job *cmdTypes.PodTemplateJob) string {
	if job.Type == cmdTypes.ResourceTypeStatefulSet || job.Type == cmdTypes.ResourceTypeReplicaset {
		if job.Template.ObjectMeta.Labels["priorityClassName"] == "interactive-preemptible" {
			return runaiPreemptibleInteractiveType
		}
		return runaiInteractiveType
	}
	return runaiTrainType
}

func (rt *RunaiTrainer) getNodeIp() (string, error) {
	nodesList, err := rt.client.CoreV1().Nodes().List(metav1.ListOptions{})

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

func getServiceEndpoints(nodeIp string, service v1.Service) []string {
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

func getServiceUrls(ingressService *v1.Service, ingresses []extensionsv1.Ingress, nodeIp string, service v1.Service) []string {
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
				continue
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

func (rt *RunaiTrainer) getIngressesForNamespace(namespace string) ([]extensionsv1.Ingress, error) {
	ingresses, err := rt.client.ExtensionsV1beta1().Ingresses(namespace).List(metav1.ListOptions{})

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

func getIngressPathOfService(ingresses []extensionsv1.Ingress, service v1.Service, port int32) *string {
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
	servicesList, err := rt.client.CoreV1().Services("").List(metav1.ListOptions{
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
