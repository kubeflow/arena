package serving

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/arenacache"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	GPU_RESOURCE_NAME         = "nvidia.com/gpu"
	ALIYUN_GPU_RESOURCE_NAME  = "aliyun.com/gpu"
	GPU_MEM_RESOURCE_NAME     = "aliyun.com/gpu-mem"
	servingNameLabelKey       = "servingName"
	servingTypeLabelKey       = "servingType"
	servingVersionLabelKey    = "servingVersion"
	istioNamespace            = "istio-system"
	grpcServingPortName       = "grpc-serving"
	restfulServingPortName    = "restful-serving"
	istioGatewayHTTPPortName  = "http2"
	istioGatewayHTTPsPortName = "https"
)

var (
	errServingJobNotFound = errors.New("serving job not found")
)

var processers map[types.ServingJobType]Processer

var once sync.Once

func GetAllProcesser() map[types.ServingJobType]Processer {
	once.Do(func() {
		processers = map[types.ServingJobType]Processer{}
		processerInits := []func() Processer{
			NewCustomServingProcesser,
			NewKFServingProcesser,
			NewTensorflowServingProcesser,
			NewTensorrtServingProcesser,
		}
		for _, initFunc := range processerInits {
			p := initFunc()
			processers[p.Type()] = p
		}
	})
	return processers
}

type servingJob struct {
	name          string
	namespace     string
	servingType   types.ServingJobType
	version       string
	pods          []*v1.Pod
	deployment    *appv1.Deployment
	services      []*v1.Service
	istioServices []*v1.Service
}

func (s *servingJob) Name() string {
	return s.name
}

func (s *servingJob) Namespace() string {
	return s.namespace
}

func (s *servingJob) Type() types.ServingJobType {
	return s.servingType
}

func (s *servingJob) Version() string {
	return s.version
}

func (s *servingJob) Pods() []*v1.Pod {
	return s.pods
}

func (s *servingJob) Deployment() *appv1.Deployment {
	return s.deployment
}

func (s *servingJob) Services() []*v1.Service {
	return s.services
}

func (s *servingJob) IPAddress() string {
	// 1.search istio gateway
	for _, svc := range s.istioServices {
		if svc.Spec.Type == v1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) != 0 {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
	}
	// 2.search ingress load balancer
	ipAddress := "N/A"
	for _, svc := range s.services {
		if svc.Spec.Type == v1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) != 0 {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
		// if the service is not we created,skip it
		ports := svc.Spec.Ports
		found := false
		for _, p := range ports {
			if p.Name == grpcServingPortName {
				found = true
				break
			}
			if p.Name == restfulServingPortName {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if ipAddress == "N/A" {
			ipAddress = svc.Spec.ClusterIP
		}
	}
	return ipAddress
}

func (s *servingJob) Age() time.Duration {
	return time.Now().Sub(s.deployment.ObjectMeta.CreationTimestamp.Time)
}

func (s *servingJob) StartTime() *metav1.Time {
	return &s.deployment.ObjectMeta.CreationTimestamp
}

func (s *servingJob) Endpoints() []types.Endpoint {
	endpoints := []types.Endpoint{}
	for _, svc := range s.istioServices {
		if svc.Spec.Type == v1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) != 0 {
			for _, p := range svc.Spec.Ports {
				if p.Name == istioGatewayHTTPPortName {
					endpoint := types.Endpoint{
						Name: "HTTP",
						Port: int(p.Port),
					}
					endpoints = append(endpoints, endpoint)
					continue
				}
				if p.Name == istioGatewayHTTPsPortName {
					endpoint := types.Endpoint{
						Name: "HTTPS",
						Port: int(p.Port),
					}
					endpoints = append(endpoints, endpoint)
					continue
				}
			}

		}
	}
	if len(endpoints) != 0 {
		return endpoints
	}
	for _, svc := range s.services {
		for _, p := range svc.Spec.Ports {
			if p.Name != grpcServingPortName && p.Name != restfulServingPortName {
				log.Debugf("the service %v has no ports which names are %v and %v,skip pick its' ports", svc.Name, grpcServingPortName, restfulServingPortName)
				continue
			}
			name := "restful"
			if strings.Contains(grpcServingPortName, p.Name) {
				name = "grpc"
			}
			endpoint := types.Endpoint{
				Name:     strings.ToUpper(name),
				NodePort: int(p.NodePort),
				Port:     int(p.Port),
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func (s *servingJob) RequestGPUs() int {
	request := uint(0)
	for _, pod := range s.pods {
		containers := pod.Spec.Containers
		for _, container := range containers {
			if val, ok := container.Resources.Limits[v1.ResourceName(GPU_RESOURCE_NAME)]; ok {
				request += uint(val.Value())
			}
			if val, ok := container.Resources.Limits[v1.ResourceName(ALIYUN_GPU_RESOURCE_NAME)]; ok {
				request += uint(val.Value())
			}
		}
	}
	return int(request)
}

func (s *servingJob) RequestGPUMemory() int {
	request := uint(0)
	for _, pod := range s.pods {
		containers := pod.Spec.Containers
		for _, container := range containers {
			if val, ok := container.Resources.Limits[v1.ResourceName(GPU_MEM_RESOURCE_NAME)]; ok {
				request += uint(val.Value())
			}
		}
	}
	return int(request)
}

func (s *servingJob) AvailableInstances() int {
	return int(s.deployment.Status.AvailableReplicas)
}

func (s *servingJob) DesiredInstances() int {
	return int(s.deployment.Status.Replicas)
}

func (s *servingJob) Instances() []types.ServingInstance {
	instances := []types.ServingInstance{}
	for _, pod := range s.pods {
		status, totalContainers, restart, readyContainer := utils.DefinePodPhaseStatus(*pod)
		age := util.ShortHumanDuration(time.Now().Sub(pod.ObjectMeta.CreationTimestamp.Time))
		instances = append(instances, types.ServingInstance{
			Name:           pod.Name,
			Status:         status,
			Age:            age,
			NodeIP:         pod.Status.HostIP,
			NodeName:       pod.Spec.NodeName,
			IP:             pod.Status.PodIP,
			ReadyContainer: readyContainer,
			TotalContainer: totalContainers,
			RestartCount:   restart,
		})
	}
	return instances
}

func (s *servingJob) Convert2JobInfo() types.ServingJobInfo {
	servingJobInfo := types.ServingJobInfo{
		Name:             s.name,
		Namespace:        s.namespace,
		Version:          s.version,
		Type:             s.servingType,
		Age:              util.ShortHumanDuration(s.Age()),
		Desired:          s.DesiredInstances(),
		IPAddress:        s.IPAddress(),
		Available:        s.AvailableInstances(),
		RequestGPU:       s.RequestGPUs(),
		RequestGPUMemory: s.RequestGPUMemory(),
		Endpoints:        s.Endpoints(),
		Instances:        s.Instances(),
	}
	return servingJobInfo
}

// processer defines the default processer
type processer struct {
	processerType   types.ServingJobType
	enable          bool
	useIstioGateway bool
	client          *kubernetes.Clientset
}

func (p *processer) Type() types.ServingJobType {
	return p.processerType
}

func (p *processer) IsEnabled() bool {
	return p.enable
}

func (p *processer) IsSupported(namespace, name, version string) bool {
	if !p.enable {
		return false
	}
	jobs, err := p.GetServingJobs(namespace, name, version)
	return err == nil && len(jobs) != 0
}

func (p *processer) GetServingJobs(namespace, name, version string) ([]ServingJob, error) {
	// if arena is daemon mode,get serving job from cache
	if config.GetArenaConfiger().IsDaemonMode() {
		return p.getServingJobsFromCache(namespace, name, version)
	}
	// get serving job from api server
	return p.getServingJobsFromAPIServer(namespace, name, version)
}

// getServingJobsFromAPIServer is used to get serving jobs from apiserver
// if version is null,return all serving jobs whose name are the same
func (p *processer) getServingJobsFromAPIServer(namespace, name, version string) ([]ServingJob, error) {
	versionLabel := fmt.Sprintf("%v=%v", servingVersionLabelKey, version)
	if version == "" {
		versionLabel = fmt.Sprintf("%v", servingVersionLabelKey)
	}
	selector := fmt.Sprintf("%v=%v,%v=%v,%v",
		servingNameLabelKey, name,
		servingTypeLabelKey, p.processerType,
		versionLabel,
	)
	return p.FilterServingJobs(namespace, false, selector)
}

func (p *processer) FilterServingJobs(namespace string, allNamespace bool, filter string) ([]ServingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	// 1.get deployment
	deploymentList, err := p.client.AppsV1().Deployments(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind: "ListOptions",
		}, LabelSelector: filter,
	})
	if err != nil {
		return nil, err
	}
	servingJobs := []ServingJob{}
	for _, d := range deploymentList.Items {
		deployment := d.DeepCopy()
		selector := fmt.Sprintf("%v=%v,%v=%v,%v=%v",
			servingNameLabelKey, deployment.Labels[servingNameLabelKey],
			servingVersionLabelKey, deployment.Labels[servingVersionLabelKey],
			servingTypeLabelKey, p.processerType,
		)
		// 2.get pods
		podList, err := p.client.CoreV1().Pods(deployment.Namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: selector,
		})
		if err != nil {
			return nil, err
		}
		pods := []*v1.Pod{}
		for _, pod := range podList.Items {
			pods = append(pods, pod.DeepCopy())
		}
		// 3.get services
		serviceList, err := p.client.CoreV1().Services(deployment.Namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: selector,
		})
		if err != nil {
			return nil, err
		}
		services := []*v1.Service{}
		for _, s := range serviceList.Items {
			services = append(services, s.DeepCopy())
		}
		// 4. get istio gateway
		istioServices := p.getIstioGatewayService()
		servingJobs = append(servingJobs, &servingJob{
			name:          deployment.Labels[servingNameLabelKey],
			namespace:     deployment.Namespace,
			servingType:   p.processerType,
			version:       deployment.Labels[servingVersionLabelKey],
			deployment:    deployment,
			pods:          pods,
			services:      services,
			istioServices: istioServices,
		})
	}
	return servingJobs, nil
}

func (p *processer) getIstioGatewayService() []*v1.Service {
	istioServices := []*v1.Service{}
	if !p.useIstioGateway {
		return istioServices
	}
	istioServiceList, err := p.client.CoreV1().Services(metav1.NamespaceAll).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("app=istio-ingressgateway,istio=ingressgateway"),
	})
	if err != nil {
		log.Debugf("failed to get istio gateway,reason: %v,we will use cluster ip to endpoint", err)
		return istioServices
	}
	for _, s := range istioServiceList.Items {
		istioServices = append(istioServices, s.DeepCopy())
	}
	return istioServices
}

func (p *processer) getServingJobsFromCache(namespace, name, version string) ([]ServingJob, error) {
	// 1.find the deployments from the cache
	deployments, allPods := arenacache.GetArenaCache().FilterDeployments(
		func(deployment *appv1.Deployment) bool {
			return p.IsKnownDeployment(namespace, name, version, deployment)
		},
		p.IsDeploymentPod,
	)
	servingJobs := []ServingJob{}
	for key, d := range deployments {
		deployment := d.DeepCopy()
		pods := allPods[key]
		services := arenacache.GetArenaCache().FilterServices(func(s *v1.Service) bool {
			return p.IsKnownService(namespace, name, version, s)
		})
		servingJobs = append(servingJobs, &servingJob{
			name:        deployment.Labels[servingNameLabelKey],
			namespace:   deployment.Namespace,
			servingType: p.processerType,
			version:     deployment.Labels[servingVersionLabelKey],
			deployment:  deployment,
			pods:        pods,
			services:    services,
		})
	}
	return servingJobs, nil
}

func (p *processer) ListServingJobs(namespace string, allNamespace bool) ([]ServingJob, error) {
	// if arena is configured as daemon,getting all serving jobs from cache is corrent
	if config.GetArenaConfiger().IsDaemonMode() {
		return p.listServingJobsFromCache(namespace, allNamespace)
	}
	return p.listServingJobsFromAPIServer(namespace, allNamespace)
}

func (p *processer) listServingJobsFromAPIServer(namespace string, allNamespace bool) ([]ServingJob, error) {
	selector := fmt.Sprintf("%v=%v",
		servingTypeLabelKey, p.processerType,
	)
	return p.FilterServingJobs(namespace, allNamespace, selector)
}

func (p *processer) listServingJobsFromCache(namespace string, allNamespace bool) ([]ServingJob, error) {
	// 1.define filter function
	filter := func(deployment *appv1.Deployment) bool {
		return deployment.Namespace == namespace && deployment.Labels[servingTypeLabelKey] == string(p.processerType)
	}
	// 2.if all namespaces is true,get all mpijobs
	if allNamespace {
		filter = func(deployment *appv1.Deployment) bool {
			return deployment.Labels[servingTypeLabelKey] == string(p.processerType)
		}
	}
	deployments, allPods := arenacache.GetArenaCache().FilterDeployments(
		filter,
		p.IsDeploymentPod,
	)
	servingJobs := []ServingJob{}
	for key, d := range deployments {
		deployment := d.DeepCopy()
		pods := allPods[key]
		services := arenacache.GetArenaCache().FilterServices(func(s *v1.Service) bool {
			return p.IsKnownService(
				deployment.Namespace,
				deployment.Labels[servingNameLabelKey],
				deployment.Labels[servingVersionLabelKey],
				s,
			)
		})
		servingJobs = append(servingJobs, &servingJob{
			name:        deployment.Labels[servingNameLabelKey],
			namespace:   deployment.Namespace,
			servingType: p.processerType,
			version:     deployment.Labels[servingVersionLabelKey],
			deployment:  deployment,
			pods:        pods,
			services:    services,
		})
	}
	return servingJobs, nil
}

func (p *processer) IsDeploymentPod(deployment *appv1.Deployment, pod *v1.Pod) bool {
	if deployment.Namespace != pod.Namespace {
		return false
	}
	if deployment.Labels[servingNameLabelKey] != pod.Labels[servingNameLabelKey] {
		return false
	}
	if deployment.Labels[servingTypeLabelKey] != pod.Labels[servingTypeLabelKey] {
		return false
	}
	if deployment.Labels[servingVersionLabelKey] != pod.Labels[servingVersionLabelKey] {
		return false
	}
	return true
}

func (p *processer) IsKnownDeployment(namespace, name, version string, deployment *appv1.Deployment) bool {
	if deployment.Namespace != namespace {
		return false
	}
	if deployment.Labels[servingNameLabelKey] != name {
		return false
	}
	if deployment.Labels[servingTypeLabelKey] != string(p.processerType) {
		return false
	}
	if version == "" {
		return true
	}
	if deployment.Labels[servingVersionLabelKey] != version {
		return false
	}
	return true
}

func (p *processer) IsKnownService(namespace, name, version string, service *v1.Service) bool {
	if service.Namespace != namespace {
		return false
	}
	if service.Labels[servingNameLabelKey] != name {
		return false
	}
	if service.Labels[servingTypeLabelKey] != string(p.processerType) {
		return false
	}
	if version == "" {
		return true
	}
	if service.Labels[servingVersionLabelKey] != version {
		return false
	}
	return true
}
