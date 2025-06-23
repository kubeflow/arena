// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serving

import (
	"fmt"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lws_v1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
	lws_client "sigs.k8s.io/lws/client-go/clientset/versioned"
)

type DistributedServingProcesser struct {
	lwsClient *lws_client.Clientset
	*processer
}

type lwsJob struct {
	lws *lws_v1.LeaderWorkerSet
	*servingJob
}

func NewDistributedServingProcesser() Processer {
	p := &processer{
		processerType:   types.DistributedServingJob,
		client:          config.GetArenaConfiger().GetClientSet(),
		enable:          true,
		useIstioGateway: false,
	}

	lwsClient := lws_client.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	return &DistributedServingProcesser{
		lwsClient: lwsClient,
		processer: p,
	}
}

func SubmitDistributedServingJob(namespace string, args *types.DistributedServingArgs) (err error) {
	nameWithVersion := fmt.Sprintf("%v-%v", args.Name, args.Version)
	args.Namespace = namespace
	processers := GetAllProcesser()
	processer, ok := processers[args.Type]
	if !ok {
		return fmt.Errorf("the processer of %v is not found", args.Type)
	}
	jobs, err := processer.GetServingJobs(args.Namespace, args.Name, args.Version)
	if err != nil {
		return err
	}
	if err := ValidateJobsBeforeSubmiting(jobs, args.Name); err != nil {
		return err
	}
	chart := util.GetChartsFolder() + "/distributed-serving"
	err = workflow.SubmitJob(nameWithVersion, string(types.DistributedServingJob), namespace, args, chart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena serve get %s --type %s -n %s` to check the job status", args.Name, args.Type, args.Namespace)
	return nil
}

func (p *DistributedServingProcesser) ListServingJobs(namespace string, allNamespace bool) ([]ServingJob, error) {
	selector := fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType)
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsIsolateUserInNamespace() {
		selector = fmt.Sprintf("%v,%v=%v", selector, types.UserNameIdLabel, arenaConfiger.GetUser().GetId())
	}
	log.Debugf("filter jobs by labels: %v", selector)
	return p.FilterServingJobs(namespace, allNamespace, selector)
}

func (p *DistributedServingProcesser) GetServingJobs(namespace, name, version string) ([]ServingJob, error) {
	selector := []string{
		fmt.Sprintf("%v=%v", servingNameLabelKey, name),
		fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType),
	}
	log.Debugf("processer %v,filter jobs by labels: %v", p.processerType, selector)
	return p.FilterServingJobs(namespace, false, strings.Join(selector, ","))
}

func (p *DistributedServingProcesser) FilterServingJobs(namespace string, allNamespace bool, label string) ([]ServingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}

	// get leaderworkerset
	lwsList, err := k8saccesser.GetK8sResourceAccesser().ListLWSJobs(p.lwsClient, namespace, label)
	if err != nil {
		return nil, err
	}

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
	for _, lws := range lwsList {
		filterPods := []*corev1.Pod{}
		for _, pod := range pods {
			if lws.Labels[servingNameLabelKey] == pod.Labels[servingNameLabelKey] &&
				lws.Labels[servingTypeLabelKey] == pod.Labels[servingTypeLabelKey] {
				filterPods = append(filterPods, pod)
			}
		}
		version := lws.Labels[servingVersionLabelKey]
		servingJobs = append(servingJobs, &lwsJob{
			lws: lws,
			servingJob: &servingJob{
				name:          lws.Labels[servingNameLabelKey],
				namespace:     lws.Namespace,
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

func (s *lwsJob) Uid() string {
	return string(s.lws.UID)
}

func (s *lwsJob) Age() time.Duration {
	return time.Since(s.lws.CreationTimestamp.Time)
}

func (s *lwsJob) StartTime() *metav1.Time {
	return &s.lws.CreationTimestamp
}

func (s *lwsJob) RequestCPUs() float64 {
	replicas := s.lws.Spec.Replicas
	size := s.lws.Spec.LeaderWorkerTemplate.Size
	masterCPUs := 0.0
	for _, c := range s.lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers {
		if val, ok := c.Resources.Limits[corev1.ResourceName(types.CPUResourceName)]; ok {
			masterCPUs += float64(val.Value())
		}
	}
	result := masterCPUs * float64(*replicas)
	if size != nil && *size > 1 {
		workerCPUs := 0.0
		for _, c := range s.lws.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers {
			if val, ok := c.Resources.Limits[corev1.ResourceName(types.CPUResourceName)]; ok {
				workerCPUs += float64(val.Value())
			}
		}
		workerCPUs *= float64((*size) - 1)
		result += float64(*replicas) * workerCPUs
	}
	return result
}

func (s *lwsJob) RequestGPUs() float64 {
	replicas := s.lws.Spec.Replicas
	size := s.lws.Spec.LeaderWorkerTemplate.Size
	masterGPUs := 0.0
	for _, c := range s.lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers {
		if val, ok := c.Resources.Limits[corev1.ResourceName(types.NvidiaGPUResourceName)]; ok {
			masterGPUs += float64(val.Value())
		}
		if val, ok := c.Resources.Limits[corev1.ResourceName(types.AliyunGPUResourceName)]; ok {
			masterGPUs += float64(val.Value())
		}
	}
	result := masterGPUs * float64(*replicas)
	if size != nil && *size > 1 {
		workerGPUs := 0.0
		for _, c := range s.lws.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers {
			if val, ok := c.Resources.Limits[corev1.ResourceName(types.NvidiaGPUResourceName)]; ok {
				workerGPUs += float64(val.Value())
			}
		}
		for _, c := range s.lws.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers {
			if val, ok := c.Resources.Limits[corev1.ResourceName(types.AliyunGPUResourceName)]; ok {
				workerGPUs += float64(val.Value())
			}
		}
		workerGPUs *= float64((*size) - 1)
		result += float64(*replicas) * workerGPUs
	}
	return result
}

func (s *lwsJob) RequestGPUMemory() int {
	replicas := s.lws.Spec.Replicas
	size := s.lws.Spec.LeaderWorkerTemplate.Size

	masterGpuMemory := 0
	for _, c := range s.lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers {
		if val, ok := c.Resources.Limits[corev1.ResourceName(types.GPUShareResourceName)]; ok {
			masterGpuMemory += int(val.Value())
		}
	}

	result := masterGpuMemory * int(*replicas)
	if size != nil && *size > 1 {
		workerGpuMemory := 0
		for _, c := range s.lws.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers {
			if val, ok := c.Resources.Limits[corev1.ResourceName(types.GPUShareResourceName)]; ok {
				workerGpuMemory += int(val.Value())
			}
		}
		workerGpuMemory *= int((*size) - 1)
		result += int(*replicas) * workerGpuMemory
	}

	return result
}

func (s *lwsJob) RequestGPUCore() int {
	replicas := s.lws.Spec.Replicas
	size := s.lws.Spec.LeaderWorkerTemplate.Size

	masterGpuCore := 0
	for _, c := range s.lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers {
		if val, ok := c.Resources.Limits[corev1.ResourceName(types.GPUCoreShareResourceName)]; ok {
			masterGpuCore += int(val.Value())
		}
	}

	result := masterGpuCore * int(*replicas)
	if size != nil && *size > 1 {
		workerGpuCore := 0
		for _, c := range s.lws.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers {
			if val, ok := c.Resources.Limits[corev1.ResourceName(types.GPUCoreShareResourceName)]; ok {
				workerGpuCore += int(val.Value())
			}
		}
		workerGpuCore *= int((*size) - 1)
		result += int(*replicas) * workerGpuCore
	}

	return result
}

func (s *lwsJob) DesiredInstances() int {
	return int(s.lws.Status.Replicas)
}

func (s *lwsJob) AvailableInstances() int {
	return int(s.lws.Status.ReadyReplicas)
}

func (s *lwsJob) GetLabels() map[string]string {
	return s.lws.Labels
}

func (s *lwsJob) Convert2JobInfo() types.ServingJobInfo {
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
