package serving

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	seldonApp = "seldon-app"
	seldonGrpcServingPortName = "grpc"
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
	selector := map[string]string{
		servingNameLabelKey: name,
		servingTypeLabelKey: string(p.processerType),
	}
	if version != "" {
		selector[servingVersionLabelKey] = version
	}
	return p.FilterServingJobs(namespace, false, selector)
}

func (p *SeldonServingProcesser) ListServingJobs(namespace string, allNamespace bool) ([]ServingJob, error) {
	selector := map[string]string{
		servingTypeLabelKey: string(p.processerType),
	}

	return p.FilterServingJobs(namespace, allNamespace, selector)
}

func (s *seldonServingJob) Convert2JobInfo() types.ServingJobInfo {
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
		RequestGPU:        s.RequestGPUs(),
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

func (p *SeldonServingProcesser) FilterServingJobs(namespace string, allNamespace bool, labels map[string]string) ([]ServingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	// 1.get deployment
	deploymentList, err := listDeployments(p.client, namespace, labels)
	if err != nil {
		return nil, err
	}
	servingJobs := []ServingJob{}
	for _, d := range deploymentList.Items {
		deployment := d.DeepCopy()
		selector := map[string]string{
			servingNameLabelKey:    deployment.Labels[servingNameLabelKey],
			servingVersionLabelKey: deployment.Labels[servingVersionLabelKey],
			servingTypeLabelKey:    string(p.processerType),
		}
		// 2.get pods
		podList, err := listJobPods(p.client, deployment.Namespace, selector)
		if err != nil {
			return nil, err
		}
		pods := []*v1.Pod{}
		for _, pod := range podList.Items {
			pods = append(pods, pod.DeepCopy())
		}
		// 3.get services
		serviceSelector := map[string]string {
			seldonApp: deployment.Labels[seldonApp],
		}
		serviceList, err := listJobServices(p.client, deployment.Namespace, serviceSelector)
		if err != nil {
			return nil, err
		}
		services := []*v1.Service{}
		for _, s := range serviceList.Items {
			services = append(services, s.DeepCopy())
		}
		// 4. get istio gateway
		istioServices := p.getIstioGatewayService()

		job := &servingJob{
			name:          deployment.Labels[servingNameLabelKey],
			namespace:     deployment.Namespace,
			servingType:   p.processerType,
			version:       deployment.Labels[servingVersionLabelKey],
			deployment:    deployment,
			pods:          pods,
			services:      services,
			istioServices: istioServices,
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
	// if job has been existed,skip to create it and return an error
	if len(jobs) != 0 {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena serve delete %s -n seldon'", args.Name, args.Name)
	}
	// the master is also considered as a worker
	chart := util.GetChartsFolder() + "/seldon-core"
	log.Infof("seldon chart path: %s", chart)
	temp, _ := json.Marshal(args)
	log.Infof("seldon args: %s", string(temp))
	err = workflow.SubmitJob(nameWithVersion, string(types.SeldonServingJob), namespace, args, chart)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", args.Name, args.Type)
	return nil
}
