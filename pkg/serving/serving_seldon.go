package serving

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	seldonApp                    = "seldon-app"
	seldonGrpcServingPortName    = "grpc"
	seldonRestfulServingPortName = "http"
)

// SeldonServingProcesser use the default processer
type SeldonServingProcesser struct {
	*processer
}

type seldonServingJob struct {
	*servingJob
}

func NewSeldonServingProcesser() Processer {
	p := &processer{
		processerType:   types.SeldonServingJob,
		client:          config.GetArenaConfiger().GetClientSet(),
		enable:          true,
		useIstioGateway: true,
	}
	return &SeldonServingProcesser{
		processer: p,
	}
}

func (p *SeldonServingProcesser) GetServingJobs(namespace, name, version string) ([]ServingJob, error) {
	selector := []string{
		fmt.Sprintf("%v=%v", servingNameLabelKey, name),
		fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType),
	}
	if version != "" {
		selector = append(selector, fmt.Sprintf("%v=%v", servingVersionLabelKey, version))
	}
	return p.FilterServingJobs(namespace, false, strings.Join(selector, ","))
}

func (p *SeldonServingProcesser) ListServingJobs(namespace string, allNamespace bool) ([]ServingJob, error) {
	selector := fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType)
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsIsolateUserInNamespace() {
		selector = fmt.Sprintf("%v,%v=%v", selector, types.UserNameIdLabel, arenaConfiger.GetUser().GetId())
	}
	return p.FilterServingJobs(namespace, allNamespace, selector)
}

func (s *seldonServingJob) Convert2JobInfo() types.ServingJobInfo {
	servingType := types.ServingTypeMap[s.servingType].Alias
	servingJobInfo := types.ServingJobInfo{
		UUID:              s.Uid(),
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

func (s *seldonServingJob) IPAddress() string {
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
			if p.Name == seldonGrpcServingPortName {
				found = true
				break
			}
			if p.Name == seldonRestfulServingPortName {
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

func (s *seldonServingJob) Endpoints() []types.Endpoint {
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
			if p.Name != seldonGrpcServingPortName && p.Name != seldonRestfulServingPortName {
				log.Debugf("the service %v has no ports which names are %v and %v,skip pick its' ports", svc.Name, grpcServingPortName, restfulServingPortName)
				continue
			}
			name := "restful"
			if strings.Contains(seldonGrpcServingPortName, p.Name) {
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

func (p *SeldonServingProcesser) FilterServingJobs(namespace string, allNamespace bool, label string) ([]ServingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}

	deployments, err := k8saccesser.GetK8sResourceAccesser().ListDeployments(namespace, label)
	if err != nil {
		return nil, err
	}
	selector := fmt.Sprintf("%v,%v,%v=%v", servingNameLabelKey, servingVersionLabelKey, servingTypeLabelKey, p.processerType)
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, selector, "", nil)
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

		serviceSelector := fmt.Sprintf("%v=%v", seldonApp, deployment.Labels[seldonApp])
		services, err := k8saccesser.GetK8sResourceAccesser().ListServices(namespace, serviceSelector)
		if err != nil {
			continue
		}

		for _, svc := range services {
			filterServices = append(filterServices, svc)
		}

		job := &servingJob{
			name:          deployment.Labels[servingNameLabelKey],
			namespace:     deployment.Namespace,
			servingType:   p.processerType,
			version:       deployment.Labels[servingVersionLabelKey],
			deployment:    deployment,
			pods:          filterPods,
			services:      filterServices,
			istioServices: istioGatewayServices,
		}

		servingJobs = append(servingJobs, &seldonServingJob{
			servingJob: job,
		})
	}

	return servingJobs, nil
}

func SubmitSeldonServingJob(namespace string, args *types.SeldonServingArgs) (err error) {
	nameWithVersion := fmt.Sprintf("%v-%v", args.Name, args.Version)
	args.Namespace = namespace
	processers := GetAllProcesser()
	processer, ok := processers[args.Type]
	if !ok {
		return fmt.Errorf("not found processer whose type is %v", args.Type)
	}
	jobs, err := processer.GetServingJobs(args.Namespace, args.Name, args.Version)
	if err != nil {
		return err
	}
	if err := ValidateJobsBeforeSubmiting(jobs, args.Name); err != nil {
		return err
	}
	// the master is also considered as a worker
	chart := util.GetChartsFolder() + "/seldon-core"
	log.Infof("seldon chart path: %s", chart)
	temp, _ := json.Marshal(args)
	log.Infof("seldon args: %s", string(temp))
	err = workflow.SubmitJob(nameWithVersion, string(types.SeldonServingJob), namespace, args, chart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena serve get %s --type %s` to check the job status", args.Name, args.Type)
	return nil
}
