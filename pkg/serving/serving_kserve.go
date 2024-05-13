package serving

import (
	"context"
	"fmt"
	"strings"
	"time"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kserveClient "github.com/kserve/kserve/pkg/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
)

// KServeProcesser use the default processer
type KServeProcesser struct {
	kserveClient *kserveClient.Clientset
	*processer
}

type kserveJob struct {
	inferenceService     *kservev1beta1.InferenceService
	inferenceDeployments []*appv1.Deployment
	*servingJob
}

func NewKServeProcesser() Processer {
	p := &processer{
		processerType:   types.KServeJob,
		client:          config.GetArenaConfiger().GetClientSet(),
		enable:          true,
		useIstioGateway: false,
	}

	kc := kserveClient.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	return &KServeProcesser{
		kserveClient: kc,
		processer:    p,
	}
}

func SubmitKServeJob(namespace string, args *types.KServeArgs) (err error) {
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
	chart := util.GetChartsFolder() + "/kserve"
	err = workflow.SubmitJob(args.Name, string(types.KServeJob), namespace, args, chart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena serve get %s --type %s -n %s` to check the job status", args.Name, args.Type, args.Namespace)
	return nil
}

func (p *KServeProcesser) ListServingJobs(namespace string, allNamespace bool) ([]ServingJob, error) {
	selector := fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType)
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsIsolateUserInNamespace() {
		selector = fmt.Sprintf("%v,%v=%v", selector, types.UserNameIdLabel, arenaConfiger.GetUser().GetId())
	}
	log.Debugf("filter jobs by labels: %v", selector)
	return p.FilterServingJobs(namespace, allNamespace, selector)
}

func (p *KServeProcesser) GetServingJobs(namespace, name, version string) ([]ServingJob, error) {
	selector := []string{
		fmt.Sprintf("%v=%v", servingNameLabelKey, name),
		fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType),
	}
	log.Debugf("processer %v,filter jobs by labels: %v", p.processerType, selector)
	return p.FilterServingJobs(namespace, false, strings.Join(selector, ","))
}

func (p *KServeProcesser) FilterServingJobs(namespace string, allNamespace bool, label string) ([]ServingJob, error) {
	ctx := context.TODO()
	if allNamespace {
		namespace = metav1.NamespaceAll
	}

	inferenceServiceList, err := k8saccesser.GetK8sResourceAccesser().ListKServeInferenceService(ctx, p.kserveClient, namespace, label)
	if err != nil {
		return nil, err
	}

	// get deployment
	deployments, err := k8saccesser.GetK8sResourceAccesser().ListDeployments(namespace, label)
	if err != nil {
		return nil, err
	}
	log.Debugf("processer: %v,found target deployments: %v", p.processerType, len(deployments))

	// get pod
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, label, "", nil)
	if err != nil {
		return nil, err
	}

	// get svc
	services, err := k8saccesser.GetK8sResourceAccesser().ListServices(namespace, label)
	if err != nil {
		return nil, err
	}

	servingJobs := []ServingJob{}
	for _, iservice := range inferenceServiceList {
		filterDeployments := []*appv1.Deployment{}
		filterPods := []*v1.Pod{}
		for _, deployment := range deployments {
			if iservice.Labels[servingNameLabelKey] == deployment.Labels[servingNameLabelKey] &&
				iservice.Labels[servingTypeLabelKey] == deployment.Labels[servingTypeLabelKey] {
				filterDeployments = append(filterDeployments, deployment)
			}
		}

		for _, pod := range pods {
			if iservice.Labels[servingNameLabelKey] == pod.Labels[servingNameLabelKey] &&
				iservice.Labels[servingTypeLabelKey] == pod.Labels[servingTypeLabelKey] {
				filterPods = append(filterPods, pod)
			}
		}

		version := iservice.Status.Components["predictor"].LatestCreatedRevision
		if len(version) > 5 {
			version = version[len(version)-5:]
		}

		servingJobs = append(servingJobs, &kserveJob{
			inferenceService:     iservice,
			inferenceDeployments: filterDeployments,
			servingJob: &servingJob{
				name:          iservice.Labels[servingNameLabelKey],
				namespace:     iservice.Namespace,
				servingType:   p.processerType,
				version:       version,
				deployment:    nil,
				pods:          filterPods,
				services:      services,
				istioServices: nil,
			},
		})
	}

	return servingJobs, nil
}

func (s *kserveJob) Uid() string {
	return string(s.inferenceService.UID)
}

func (s *kserveJob) Age() time.Duration {
	return time.Since(s.inferenceService.ObjectMeta.CreationTimestamp.Time)
}

func (s *kserveJob) StartTime() *metav1.Time {
	return &s.inferenceService.ObjectMeta.CreationTimestamp
}

func (s *kserveJob) Endpoints() []types.Endpoint {
	endpoints := []types.Endpoint{}
	if s.inferenceService.Status.URL.String() != "" {
		endpoint := types.Endpoint{
			Port: 80,
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

func (s *kserveJob) IPAddress() string {
	return s.inferenceService.Status.URL.String()
}

func (s *kserveJob) RequestCPUs() float64 {
	var result float64
	for _, dp := range s.inferenceDeployments {
		replicas := *dp.Spec.Replicas
		podCPUs := 0.0
		for _, c := range dp.Spec.Template.Spec.Containers {
			if val, ok := c.Resources.Limits[v1.ResourceName(types.CPUResourceName)]; ok {
				podCPUs += float64(val.Value())
			}
		}
		result = result + float64(replicas)*podCPUs
	}
	return result
}

func (s *kserveJob) RequestGPUs() float64 {
	var result float64
	for _, dp := range s.inferenceDeployments {
		replicas := *dp.Spec.Replicas
		podGPUs := 0
		for _, c := range dp.Spec.Template.Spec.Containers {
			if val, ok := c.Resources.Limits[v1.ResourceName(types.NvidiaGPUResourceName)]; ok {
				podGPUs += int(val.Value())
			}
			if val, ok := c.Resources.Limits[v1.ResourceName(types.AliyunGPUResourceName)]; ok {
				podGPUs += int(val.Value())
			}
		}
		result = result + float64(replicas*int32(podGPUs))
	}
	return result
}

func (s *kserveJob) RequestGPUMemory() int {
	var result int
	for _, dp := range s.inferenceDeployments {
		replicas := *dp.Spec.Replicas
		podGPUMemory := 0
		for _, c := range dp.Spec.Template.Spec.Containers {
			if val, ok := c.Resources.Limits[v1.ResourceName(types.GPUShareResourceName)]; ok {
				podGPUMemory += int(val.Value())
			}
		}
		result = result + int(replicas*int32(podGPUMemory))
	}
	return result
}

func (s *kserveJob) RequestGPUCore() int {
	var result int
	for _, dp := range s.inferenceDeployments {
		replicas := *dp.Spec.Replicas
		podGPUCore := 0
		for _, c := range dp.Spec.Template.Spec.Containers {
			if val, ok := c.Resources.Limits[v1.ResourceName(types.GPUCoreShareResourceName)]; ok {
				podGPUCore += int(val.Value())
			}
		}
		result = result + int(replicas*int32(podGPUCore))
	}
	return result
}

func (s *kserveJob) DesiredInstances() int {
	var desired int32
	for _, dp := range s.inferenceDeployments {
		desired = desired + dp.Status.Replicas
	}
	return int(desired)
}

func (s *kserveJob) AvailableInstances() int {
	var available int32
	for _, dp := range s.inferenceDeployments {
		available = available + dp.Status.AvailableReplicas
	}
	return int(available)
}

func (s *kserveJob) GetLabels() map[string]string {
	return s.inferenceService.Labels
}

func (s *kserveJob) Convert2JobInfo() types.ServingJobInfo {
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
		RequestCPUs:       s.RequestCPUs(),
		RequestGPUs:       s.RequestGPUs(),
		RequestGPUMemory:  s.RequestGPUMemory(),
		RequestGPUCore:    s.RequestGPUCore(),
		Endpoints:         s.Endpoints(),
		Instances:         s.Instances(),
		CreationTimestamp: s.StartTime().Unix(),
	}
	return servingJobInfo
}
