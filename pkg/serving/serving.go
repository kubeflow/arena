package serving

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
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
		locker := new(sync.RWMutex)
		processers = map[types.ServingJobType]Processer{}
		processerInits := []func() Processer{
			NewCustomServingProcesser,
			NewKFServingProcesser,
			NewTensorflowServingProcesser,
			NewTensorrtServingProcesser,
			NewSeldonServingProcesser,
			NewTritonServingProcesser,
		}
		var wg sync.WaitGroup
		for _, initFunc := range processerInits {
			wg.Add(1)
			f := initFunc
			go func() {
				defer wg.Done()
				p := f()
				locker.Lock()
				processers[p.Type()] = p
				locker.Unlock()
			}()
		}
		wg.Wait()
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

func (s *servingJob) RequestGPUs() float64 {
	replicas := *s.deployment.Spec.Replicas
	podGPUs := 0
	for _, c := range s.deployment.Spec.Template.Spec.Containers {
		if val, ok := c.Resources.Limits[v1.ResourceName(types.NvidiaGPUResourceName)]; ok {
			podGPUs += int(val.Value())
		}
		if val, ok := c.Resources.Limits[v1.ResourceName(types.AliyunGPUResourceName)]; ok {
			podGPUs += int(val.Value())
		}
	}
	return float64(replicas * int32(podGPUs))
}

func (s *servingJob) RequestGPUMemory() int {
	replicas := *s.deployment.Spec.Replicas
	podGPUMemory := 0
	for _, c := range s.deployment.Spec.Template.Spec.Containers {
		if val, ok := c.Resources.Limits[v1.ResourceName(types.GPUShareResourceName)]; ok {
			podGPUMemory += int(val.Value())
		}
	}
	return int(replicas * int32(podGPUMemory))
}

func (s *servingJob) AvailableInstances() int {
	return int(s.deployment.Status.AvailableReplicas)
}

func (s *servingJob) DesiredInstances() int {
	return int(s.deployment.Status.Replicas)
}

func (s *servingJob) Instances() []types.ServingInstance {
	instances := []types.ServingInstance{}
	for index, pod := range s.pods {
		status, totalContainers, restart, readyContainer := utils.DefinePodPhaseStatus(*pod)
		age := util.ShortHumanDuration(time.Now().Sub(pod.ObjectMeta.CreationTimestamp.Time))
		gpuMemory := utils.GPUMemoryCountInPod(pod)
		gpus := getPodGPUs(pod, gpuMemory, index)
		instances = append(instances, types.ServingInstance{
			Name:             pod.Name,
			Status:           status,
			Age:              age,
			NodeIP:           pod.Status.HostIP,
			NodeName:         pod.Spec.NodeName,
			IP:               pod.Status.PodIP,
			ReadyContainer:   readyContainer,
			TotalContainer:   totalContainers,
			RestartCount:     restart,
			RequestGPUs:      gpus,
			RequestGPUMemory: gpuMemory,
		})
	}
	return instances
}

func (s *servingJob) Convert2JobInfo() types.ServingJobInfo {
	servingType := types.ServingTypeMap[s.servingType].Alias
	servingJobInfo := types.ServingJobInfo{
		Name:              s.name,
		Namespace:         s.namespace,
		Version:           s.version,
		Type:              servingType,
		Age:               util.ShortHumanDuration(s.Age()),
		Desired:           s.DesiredInstances(),
		IPAddress:         s.IPAddress(),
		Available:         s.AvailableInstances(),
		RequestGPUs:       s.RequestGPUs(),
		RequestGPUMemory:  s.RequestGPUMemory(),
		Endpoints:         s.Endpoints(),
		Instances:         s.Instances(),
		CreationTimestamp: s.StartTime().Unix(),
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
	selector := fmt.Sprintf("%v=%v,%v=%v", servingNameLabelKey, name, servingTypeLabelKey, p.processerType)
	if version != "" {
		selector = fmt.Sprintf("%v=%v,%v=%v,%v=%v",
			servingNameLabelKey,
			name,
			servingTypeLabelKey,
			p.processerType,
			servingVersionLabelKey,
			version,
		)
	}
	return p.FilterServingJobs(namespace, false, selector)
}

func (p *processer) FilterServingJobs(namespace string, allNamespace bool, label string) ([]ServingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	// 1.get deployment
	deployments, err := k8saccesser.GetK8sResourceAccesser().ListDeployments(namespace, label)
	if err != nil {
		return nil, err
	}
	selector := fmt.Sprintf("%v,%v,%v=%v", servingNameLabelKey, servingVersionLabelKey, servingTypeLabelKey, p.processerType)
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, selector, "", nil)
	if err != nil {
		return nil, err
	}
	services, err := k8saccesser.GetK8sResourceAccesser().ListServices(namespace, selector)
	if err != nil {
		return nil, err
	}
	istioGatewayServices := p.getIstioGatewayService()
	servingJobs := []ServingJob{}
	for _, deployment := range deployments {
		filterPods := []*v1.Pod{}
		filterServices := []*v1.Service{}
		for _, pod := range pods {
			if p.isDeploymentPod(deployment, pod) {
				filterPods = append(filterPods, pod)
			}
		}
		for _, svc := range services {
			if p.isDeploymentService(deployment, svc) {
				filterServices = append(filterServices, svc)
			}
		}
		// 4. get istio gateway
		servingJobs = append(servingJobs, &servingJob{
			name:          deployment.Labels[servingNameLabelKey],
			namespace:     deployment.Namespace,
			servingType:   p.processerType,
			version:       deployment.Labels[servingVersionLabelKey],
			deployment:    deployment,
			pods:          filterPods,
			services:      filterServices,
			istioServices: istioGatewayServices,
		})
	}
	return servingJobs, nil
}

func (p *processer) isDeploymentPod(d *appv1.Deployment, pod *v1.Pod) bool {
	if d.Namespace != pod.Namespace {
		return false
	}
	if pod.Labels[servingNameLabelKey] != d.Labels[servingNameLabelKey] {
		return false
	}
	if pod.Labels[servingVersionLabelKey] != d.Labels[servingVersionLabelKey] {
		return false
	}
	if pod.Labels[servingTypeLabelKey] != string(p.processerType) {
		return false
	}
	return true
}

func (p *processer) isDeploymentService(d *appv1.Deployment, svc *v1.Service) bool {
	if d.Namespace != svc.Namespace {
		return false
	}
	if svc.Labels[servingNameLabelKey] != d.Labels[servingNameLabelKey] {
		return false
	}
	if svc.Labels[servingVersionLabelKey] != d.Labels[servingVersionLabelKey] {
		return false
	}
	if svc.Labels[servingTypeLabelKey] != string(p.processerType) {
		return false
	}
	return true

}

func (p *processer) getIstioGatewayService() []*v1.Service {
	istioServices := []*v1.Service{}
	if !p.useIstioGateway {
		return istioServices
	}
	istioServices, _ = k8saccesser.GetK8sResourceAccesser().ListServices(
		metav1.NamespaceAll,
		"app=istio-ingressgateway,istio=ingressgateway")
	return istioServices
}

func (p *processer) ListServingJobs(namespace string, allNamespace bool) ([]ServingJob, error) {
	return p.FilterServingJobs(namespace, allNamespace, fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType))
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

func getNodeGPUMemory(nodeName string) float64 {
	node, err := k8saccesser.GetK8sResourceAccesser().GetNode(nodeName)
	if err != nil {
		log.Debugf("failed to get node gpu memory,reason: %v", err)
		return float64(0)
	}
	totalGPUs := getResourceOfGPUShareNode(node, types.GPUShareCountName)
	totalGPUMemory := getResourceOfGPUShareNode(node, types.GPUShareResourceName)
	if totalGPUs == 0 {
		return float64(0)
	}
	return totalGPUMemory / totalGPUs
}

func getResourceOfGPUShareNode(node *v1.Node, resourceName string) float64 {
	val, ok := node.Status.Capacity[v1.ResourceName(resourceName)]
	if !ok {
		return 0
	}
	return float64(val.Value())
}

func getPodGPUs(pod *v1.Pod, gpuMemory int, index int) float64 {
	if utils.IsCompletedPod(pod) {
		return float64(0)
	}
	if pod.Status.Phase != v1.PodRunning {
		return float64(0)
	}
	if len(pod.Spec.NodeName) == 0 {
		return float64(0)
	}
	if gpuMemory != 0 {
		nodeGPUMemory := getNodeGPUMemory(pod.Spec.NodeName)
		if index == 0 {
			log.Debugf("node name: %v,single gpu memory: %vGiB\n", pod.Spec.NodeName, nodeGPUMemory)
		}
		if nodeGPUMemory == float64(0) {
			return float64(0)
		}
		return float64(gpuMemory) / nodeGPUMemory
	}
	return float64(utils.GPUCountInPod(pod) + utils.AliyunGPUCountInPod(pod))
}
