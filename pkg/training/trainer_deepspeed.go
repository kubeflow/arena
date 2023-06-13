// Copyright 2023 The Kubeflow Authors
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

package training

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/operators/et-operator/api/v1alpha1"
	"github.com/kubeflow/arena/pkg/operators/et-operator/client/clientset/versioned"
)

const (
	// added key of labels for pods.
	deepspeedLabelGroupName            = "group-name"
	deepspeedLabelTrainingJobName      = "training-job-name"
	deepspeedLabelTrainingJobRole      = "training-job-role"
	deepspeedJobMetaDataAnnotationsKey = "kubectl.kubernetes.io/last-applied-configuration"
)

// DeepSpeedJob Information
type DeepSpeedJob struct {
	*BasicJobInfo
	trainingjob  *v1alpha1.TrainingJob
	pods         []*v1.Pod // all the pods including statefulset and job
	chiefPod     *v1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType // return trainer type
}

func (dsj *DeepSpeedJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
	var urls []string
	return urls, nil
}

func (dsj *DeepSpeedJob) Name() string {
	return dsj.name
}

func (dsj *DeepSpeedJob) Uid() string {
	return string(dsj.trainingjob.UID)
}

// Get the chief Pod of the Job.
func (dsj *DeepSpeedJob) ChiefPod() *v1.Pod {
	return dsj.chiefPod
}

func (dsj *DeepSpeedJob) Trainer() types.TrainingJobType {
	return dsj.trainerType
}

// Get all the pods of the Training Job
func (dsj *DeepSpeedJob) AllPods() []*v1.Pod {
	return dsj.pods
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (dsj *DeepSpeedJob) GetStatus() (status string) {
	status = "UNKNOWN"
	if dsj.trainingjob.Name == "" {
		return status
	}

	if dsj.isSucceeded() {
		status = "SUCCEEDED"
	} else if dsj.isFailed() {
		status = "FAILED"
	} else if dsj.isPending() {
		status = "PENDING"
	} else if dsj.isScaling() {
		status = "SCALING"
	} else if dsj.isMaxWaitTimeExceeded() {
		status = "MAX_WAIT_TIME_EXCEEDED"
	} else {
		status = "RUNNING"
	}

	return status
}

// Get the start time
func (dsj *DeepSpeedJob) StartTime() *metav1.Time {
	return &dsj.trainingjob.CreationTimestamp
}

// Get the Job Age
func (dsj *DeepSpeedJob) Age() time.Duration {
	tj := dsj.trainingjob

	// use creation timestamp
	if tj.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(tj.CreationTimestamp.Time)
}

// Get the Job Training Duration
func (dsj *DeepSpeedJob) Duration() time.Duration {
	trainingjob := dsj.trainingjob

	if trainingjob.CreationTimestamp.IsZero() {
		return 0
	}

	if dsj.isFailed() {
		cond := getPodLatestCondition(dsj.chiefPod)
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Time.Sub(trainingjob.CreationTimestamp.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", dsj.chiefPod.Name)
		}
	}

	return metav1.Now().Sub(trainingjob.CreationTimestamp.Time)
}

// Requested GPU count of the Job
func (dsj *DeepSpeedJob) RequestedGPU() int64 {
	if dsj.requestedGPU > 0 {
		return dsj.requestedGPU
	}
	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(dsj.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}
	for _, pod := range dsj.pods {
		dsj.requestedGPU += gpuInPod(*pod)
	}
	return dsj.requestedGPU
}

// Requested GPU count of the Job
func (dsj *DeepSpeedJob) AllocatedGPU() int64 {
	if dsj.allocatedGPU > 0 {
		return dsj.allocatedGPU
	}
	for _, pod := range dsj.pods {
		dsj.allocatedGPU += gpuInActivePod(*pod)
	}
	return dsj.allocatedGPU
}

// Get the hostIP of the chief Pod
func (dsj *DeepSpeedJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if dsj.GetStatus() == "RUNNING" {
		hostIP = dsj.chiefPod.Status.HostIP
	}

	return hostIP
}

func (dsj *DeepSpeedJob) Namespace() string {
	return dsj.trainingjob.Namespace
}

func (dsj *DeepSpeedJob) GetTrainJob() interface{} {
	return dsj.trainingjob
}

func (dsj *DeepSpeedJob) GetWorkerMaxReplicas(maxWorkers int) interface{} {
	_, worker := parseAnnotations(dsj.trainingjob)
	log.Infof("worker: %v", worker)
	if len(worker) == 0 {
		return maxWorkers
	}
	if _, ok := worker["maxReplicas"]; ok {
		maxWorkers = int(worker["maxReplicas"].(float64))
	}
	return maxWorkers
}

func (dsj *DeepSpeedJob) GetWorkerMinReplicas(minWorkers int) interface{} {
	_, worker := parseAnnotations(dsj.trainingjob)
	log.Infof("worker: %v", worker)
	if len(worker) == 0 {
		return minWorkers
	}
	if _, ok := worker["minReplicas"]; ok {
		minWorkers = int(worker["minReplicas"].(float64))
	}
	return minWorkers
}

func (dsj *DeepSpeedJob) isSucceeded() bool {
	// status.DeepSpeedJobLauncherStatusType
	return dsj.trainingjob.Status.Phase == "Succeeded"
}

func (dsj *DeepSpeedJob) isFailed() bool {
	return dsj.trainingjob.Status.Phase == "Failed" && !dsj.hasMaxWaitTimeExceededAnnotation()
}

func (dsj *DeepSpeedJob) isScaling() bool {
	return dsj.trainingjob.Status.Phase == "Scaling"
}

func (dsj *DeepSpeedJob) isPending() bool {

	if len(dsj.chiefPod.Name) == 0 || dsj.chiefPod.Status.Phase == v1.PodPending {
		log.Debugf("The DeepSpeedJob is pending due to chiefPod is not ready")
		return true
	}

	return false
}

func (dsj *DeepSpeedJob) hasMaxWaitTimeExceededAnnotation() bool {
	if val, ok := dsj.trainingjob.Annotations[spotInstanceJobStatusAnnotation]; ok {
		if val == "timeout" {
			return true
		}
	}
	return false
}

func (dsj *DeepSpeedJob) isMaxWaitTimeExceeded() bool {
	return dsj.trainingjob.Status.Phase == "Failed" && dsj.hasMaxWaitTimeExceededAnnotation()
}

// GetPriorityClass Get PriorityClass
func (dsj *DeepSpeedJob) GetPriorityClass() string {
	launcher, worker := parseAnnotations(dsj.trainingjob)
	if launcher != nil {
		if _, ok := launcher["template"]; ok {
			podTemplate := launcher["template"].(map[string]interface{})
			if _, ok := podTemplate["spec"]; ok {
				podSpec := podTemplate["spec"].(map[string]interface{})
				log.Debugf("podSpec: %v", podSpec)
				if pc, ok := podSpec["priorityClassName"]; ok && pc != "" {
					return pc.(string)
				}
			}
		}
	}

	if worker != nil {
		if _, ok := worker["template"]; ok {
			podTemplate := worker["template"].(map[string]interface{})
			if _, ok := podTemplate["spec"]; ok {
				podSpec := podTemplate["spec"].(map[string]interface{})
				log.Debugf("podSpec: %v", podSpec)
				if pc, ok := podSpec["priorityClassName"]; ok && pc != "" {
					return pc.(string)
				}
			}
		}
	}
	return ""
}

// DeepSpeedJobTrainer DeepSpeed Job trainer
type DeepSpeedJobTrainer struct {
	client      *kubernetes.Clientset
	jobClient   *versioned.Clientset
	trainerType types.TrainingJobType
	// check if it's enabled
	enabled bool
}

// NewDeepSpeedJobTrainer new deepspeed job trainer
func NewDeepSpeedJobTrainer() Trainer {
	enable := false
	jobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	_, err := config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.ETCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("DeepSpeedJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("DeepSpeedJobTrainer is disabled,reason: %v", err)
	}
	log.Debugf("Succeed to init DeepSpeedJobTrainer")
	return &DeepSpeedJobTrainer{
		jobClient:   jobClient,
		client:      config.GetArenaConfiger().GetClientSet(),
		trainerType: types.DeepSpeedTrainingJob,
		enabled:     enable,
	}
}

func (dst *DeepSpeedJobTrainer) IsEnabled() bool {
	return dst.enabled
}

// Get the type
func (dst *DeepSpeedJobTrainer) Type() types.TrainingJobType {
	return dst.trainerType
}

// check if it's et job
func (dst *DeepSpeedJobTrainer) IsSupported(name, ns string) bool {
	if !dst.enabled {
		return false
	}
	_, err := dst.GetTrainingJob(name, ns)
	return err == nil
}

func (dst *DeepSpeedJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	// 1. Get the batchJob of training Job
	deepspeedjob, err := k8saccesser.GetK8sResourceAccesser().GetETJob(dst.jobClient, namespace, name)
	if err != nil {
		return nil, err
	}
	if err := CheckJobIsOwnedByTrainer(deepspeedjob.Labels); err != nil {
		return nil, err
	}
	if !dst.isDeepSpeedJob(name, namespace, deepspeedjob) {
		return nil, types.ErrTrainingJobNotFound
	}
	// 2. Find the pod list, and determine the pod of the job
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("release=%v,app=%v", name, dst.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	jobs, err := k8saccesser.GetK8sResourceAccesser().ListBatchJobs(namespace, deepspeedLabelTrainingJobName)
	if err != nil {
		return nil, err
	}
	statefulsets, err := k8saccesser.GetK8sResourceAccesser().ListStatefulSets(namespace, deepspeedLabelTrainingJobName)
	if err != nil {
		return nil, err
	}
	filterPods, chiefPod := getPodsOfDeepSpeedJob(deepspeedjob, dst, pods)
	return &DeepSpeedJob{
		BasicJobInfo: &BasicJobInfo{
			resources: dst.resources(statefulsets, jobs, name, namespace, filterPods),
			name:      name,
		},
		trainingjob: deepspeedjob,
		chiefPod:    chiefPod,
		pods:        filterPods,
		trainerType: dst.Type(),
	}, nil

}

func (dst *DeepSpeedJobTrainer) isChiefPod(item *v1.Pod) bool {
	if item.Labels[deepspeedLabelTrainingJobRole] != "launcher" {
		return false
	}
	log.Debugf("the pod %s with labels training-job-role=laucher", item.Name)
	return true
}

func (dst *DeepSpeedJobTrainer) isDeepSpeedJob(name, ns string, item *v1alpha1.TrainingJob) bool {
	if item.Labels["release"] != name {
		return false
	}
	if item.Labels["app"] != string(dst.trainerType) {
		return false
	}
	if item.Namespace != ns {
		return false
	}
	return true
}

func (dst *DeepSpeedJobTrainer) isDeepSpeedPod(name, ns string, pod *v1.Pod) bool {
	return utils.IsDeepSpeedPod(name, ns, pod)
}

func (dst *DeepSpeedJobTrainer) resources(statefulsets []*appv1.StatefulSet, batchJobs []*batchv1.Job, name string, namespace string, pods []*v1.Pod) []Resource {
	resources := []Resource{}
	// 2. Find the pod list, and determine the pod of the job
	for _, sts := range statefulsets {
		if sts.Labels[deepspeedLabelTrainingJobName] != name {
			continue
		}
		if sts.Namespace != namespace {
			continue
		}
		resources = append(resources, Resource{
			Name:         sts.Name,
			Uid:          string(sts.UID),
			ResourceType: ResourceTypeStatefulSet,
		})
	}

	for _, job := range batchJobs {
		if job.Namespace != namespace {
			continue
		}
		if job.Labels[deepspeedLabelTrainingJobName] != name {
			continue
		}
		resources = append(resources, Resource{
			Name:         job.Name,
			Uid:          string(job.UID),
			ResourceType: ResourceTypeJob,
		})
	}
	resources = append(resources, podResources(pods)...)
	return resources
}

/**
* List Training jobs
 */

func (dst *DeepSpeedJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	jobLabels := GetTrainingJobLabels(dst.Type())
	deepspeedjobs, err := k8saccesser.GetK8sResourceAccesser().ListETJobs(dst.jobClient, namespace, jobLabels)
	if err != nil {
		return trainingJobs, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", dst.Type()), "", nil)
	if err != nil {
		return trainingJobs, err
	}
	for _, dpjob := range deepspeedjobs {
		// Find the pod list, and determine the pod of the job
		filterPods, chiefPod := getPodsOfDeepSpeedJob(dpjob, dst, pods)
		trainingJobs = append(trainingJobs, &DeepSpeedJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      dpjob.Name,
			},
			trainingjob: dpjob,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: dst.Type(),
		})
	}
	return trainingJobs, nil
}

func getPodsOfDeepSpeedJob(job *v1alpha1.TrainingJob, dst *DeepSpeedJobTrainer, podList []*v1.Pod) ([]*v1.Pod, *v1.Pod) {
	return getPodsOfTrainingJob(job.Name, job.Namespace, podList, dst.isDeepSpeedPod, func(pod *v1.Pod) bool {
		return dst.isChiefPod(pod)
	})
}
