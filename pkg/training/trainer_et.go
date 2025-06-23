// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package training

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/operators/et-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/operators/et-operator/api/v1alpha1"
)

const (
	// et-operator added key of labels for pods.
	etLabelGroupName            = "group-name"
	etLabelTrainingJobName      = "training-job-name"
	etLabelTrainingJobRole      = "training-job-role"
	etJobMetaDataAnnotationsKey = "kubectl.kubernetes.io/last-applied-configuration"
)

// ET Job Information
type ETJob struct {
	*BasicJobInfo
	trainingjob  *v1alpha1.TrainingJob
	pods         []*corev1.Pod // all the pods including statefulset and job
	chiefPod     *corev1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType // return trainer type
}

func (ej *ETJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
	var urls []string
	return urls, nil
}

func (ej *ETJob) Name() string {
	return ej.name
}

func (ej *ETJob) Uid() string {
	return string(ej.trainingjob.UID)
}

// Get the chief Pod of the Job.
func (ej *ETJob) ChiefPod() *corev1.Pod {
	return ej.chiefPod
}

func (ej *ETJob) Trainer() types.TrainingJobType {
	return ej.trainerType
}

// Get all the pods of the Training Job
func (ej *ETJob) AllPods() []*corev1.Pod {
	return ej.pods
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (ej *ETJob) GetStatus() (status string) {
	status = "UNKNOWN"
	if ej.trainingjob.Name == "" {
		return status
	}

	if ej.isSucceeded() {
		status = "SUCCEEDED"
	} else if ej.isFailed() {
		status = "FAILED"
	} else if ej.isPending() {
		status = "PENDING"
	} else if ej.isScaling() {
		status = "SCALING"
	} else if ej.isMaxWaitTimeExceeded() {
		status = "MAX_WAIT_TIME_EXCEEDED"
	} else {
		status = "RUNNING"
	}

	return status
}

// Get the start time
func (ej *ETJob) StartTime() *metav1.Time {
	return &ej.trainingjob.CreationTimestamp
}

// Get the Job Age
func (ej *ETJob) Age() time.Duration {
	tj := ej.trainingjob

	// use creation timestamp
	if tj.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(tj.CreationTimestamp.Time)
}

// Get the Job Training Duration
func (ej *ETJob) Duration() time.Duration {
	trainingjob := ej.trainingjob

	if trainingjob.CreationTimestamp.IsZero() {
		return 0
	}

	if ej.isFailed() {
		cond := getPodLatestCondition(ej.chiefPod)
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Sub(trainingjob.CreationTimestamp.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", ej.chiefPod.Name)
		}
	}

	return metav1.Now().Sub(trainingjob.CreationTimestamp.Time)
}

// Requested GPU count of the Job
func (ej *ETJob) RequestedGPU() int64 {
	if ej.requestedGPU > 0 {
		return ej.requestedGPU
	}
	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(ej.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}
	for _, pod := range ej.pods {
		ej.requestedGPU += gpuInPod(*pod)
	}
	return ej.requestedGPU
}

// Requested GPU count of the Job
func (ej *ETJob) AllocatedGPU() int64 {
	if ej.allocatedGPU > 0 {
		return ej.allocatedGPU
	}
	for _, pod := range ej.pods {
		ej.allocatedGPU += gpuInActivePod(*pod)
	}
	return ej.allocatedGPU
}

// Get the hostIP of the chief Pod
func (ej *ETJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if ej.GetStatus() == "RUNNING" {
		hostIP = ej.chiefPod.Status.HostIP
	}

	return hostIP
}

func (ej *ETJob) Namespace() string {
	return ej.trainingjob.Namespace
}

func (ej *ETJob) GetTrainJob() interface{} {
	return ej.trainingjob
}

func (ej *ETJob) GetLabels() map[string]string {
	return ej.trainingjob.Labels
}

func (ej *ETJob) GetWorkerMaxReplicas(maxWorkers int) interface{} {
	_, worker := parseAnnotations(ej.trainingjob)
	log.Infof("worker: %v", worker)
	if len(worker) == 0 {
		return maxWorkers
	}
	if _, ok := worker["maxReplicas"]; ok {
		maxWorkers = int(worker["maxReplicas"].(float64))
	}
	return maxWorkers
}

func (ej *ETJob) GetWorkerMinReplicas(minWorkers int) interface{} {
	_, worker := parseAnnotations(ej.trainingjob)
	log.Infof("worker: %v", worker)
	if len(worker) == 0 {
		return minWorkers
	}
	if _, ok := worker["minReplicas"]; ok {
		minWorkers = int(worker["minReplicas"].(float64))
	}
	return minWorkers
}

func (ej *ETJob) isSucceeded() bool {
	// status.ETJobLauncherStatusType
	return ej.trainingjob.Status.Phase == "Succeeded"
}

func (ej *ETJob) isFailed() bool {
	return ej.trainingjob.Status.Phase == "Failed" && !ej.hasMaxWaitTimeExceededAnnotation()
}

func (ej *ETJob) isScaling() bool {
	return ej.trainingjob.Status.Phase == "Scaling"
}

func (ej *ETJob) isPending() bool {

	if len(ej.chiefPod.Name) == 0 || ej.chiefPod.Status.Phase == corev1.PodPending {
		log.Debugf("The ETJob is pending due to chiefPod is not ready")
		return true
	}

	return false
}

func (ej *ETJob) hasMaxWaitTimeExceededAnnotation() bool {
	if val, ok := ej.trainingjob.Annotations[spotInstanceJobStatusAnnotation]; ok {
		if val == "timeout" {
			return true
		}
	}
	return false
}

func (ej *ETJob) isMaxWaitTimeExceeded() bool {
	return ej.trainingjob.Status.Phase == "Failed" && ej.hasMaxWaitTimeExceededAnnotation()
}

// ET Job trainer
type ETJobTrainer struct {
	client      *kubernetes.Clientset
	jobClient   *versioned.Clientset
	trainerType types.TrainingJobType
	// check if it's enabled
	enabled bool
}

// NewETJobTrainer
func NewETJobTrainer() Trainer {
	enable := false
	jobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	_, err := config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.ETCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("ETJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("ETJobTrainer is disabled,reason: %v", err)
	}
	log.Debugf("Succeed to init ETJobTrainer")
	return &ETJobTrainer{
		jobClient:   jobClient,
		client:      config.GetArenaConfiger().GetClientSet(),
		trainerType: types.ETTrainingJob,
		enabled:     enable,
	}
}

func (ejt *ETJobTrainer) IsEnabled() bool {
	return ejt.enabled
}

// Get the type
func (ejt *ETJobTrainer) Type() types.TrainingJobType {
	return ejt.trainerType
}

// check if it's et job
func (ejt *ETJobTrainer) IsSupported(name, ns string) bool {
	if !ejt.enabled {
		return false
	}
	_, err := ejt.GetTrainingJob(name, ns)
	return err == nil
}

func (ejt *ETJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	// 1. Get the batchJob of training Job
	etjob, err := k8saccesser.GetK8sResourceAccesser().GetETJob(ejt.jobClient, namespace, name)
	if err != nil {
		return nil, err
	}
	if err := CheckJobIsOwnedByTrainer(etjob.Labels); err != nil {
		return nil, err
	}
	if !ejt.isETJob(name, namespace, etjob) {
		return nil, types.ErrTrainingJobNotFound
	}
	// 2. Find the pod list, and determine the pod of the job
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("release=%v,app=%v", name, ejt.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	jobs, err := k8saccesser.GetK8sResourceAccesser().ListBatchJobs(namespace, etLabelTrainingJobName)
	if err != nil {
		return nil, err
	}
	statefulsets, err := k8saccesser.GetK8sResourceAccesser().ListStatefulSets(namespace, etLabelTrainingJobName)
	if err != nil {
		return nil, err
	}
	filterPods, chiefPod := getPodsOfETJob(etjob, ejt, pods)
	return &ETJob{
		BasicJobInfo: &BasicJobInfo{
			resources: ejt.resources(statefulsets, jobs, name, namespace, filterPods),
			name:      name,
		},
		trainingjob: etjob,
		chiefPod:    chiefPod,
		pods:        filterPods,
		trainerType: ejt.Type(),
	}, nil

}

func (ejt *ETJobTrainer) isChiefPod(item *corev1.Pod) bool {
	if item.Labels[etLabelTrainingJobRole] != "launcher" {
		return false
	}
	log.Debugf("the pod %s with labels training-job-role=launcher", item.Name)
	return true
}

func (ejt *ETJobTrainer) isETJob(name, ns string, item *v1alpha1.TrainingJob) bool {
	if item.Labels["release"] != name {
		return false
	}
	if item.Labels["app"] != string(ejt.trainerType) {
		return false
	}
	if item.Namespace != ns {
		return false
	}
	return true
}

func (ejt *ETJobTrainer) isETPod(name, ns string, pod *corev1.Pod) bool {
	return utils.IsETPod(name, ns, pod)
}

func (ejt *ETJobTrainer) resources(statefulsets []*appsv1.StatefulSet, batchJobs []*batchv1.Job, name string, namespace string, pods []*corev1.Pod) []Resource {
	resources := []Resource{}
	// 2. Find the pod list, and determine the pod of the job
	for _, sts := range statefulsets {
		if sts.Labels[etLabelTrainingJobName] != name {
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
		if job.Labels[etLabelTrainingJobName] != name {
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

func (ejt *ETJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	jobLabels := GetTrainingJobLabels(ejt.Type())
	etjobs, err := k8saccesser.GetK8sResourceAccesser().ListETJobs(ejt.jobClient, namespace, jobLabels)
	if err != nil {
		return trainingJobs, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", ejt.Type()), "", nil)
	if err != nil {
		return trainingJobs, err
	}
	for _, etjob := range etjobs {
		// Find the pod list, and determine the pod of the job
		filterPods, chiefPod := getPodsOfETJob(etjob, ejt, pods)
		trainingJobs = append(trainingJobs, &ETJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      etjob.Name,
			},
			trainingjob: etjob,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: ejt.Type(),
		})
	}
	return trainingJobs, nil
}

func parseAnnotations(trainingjob *v1alpha1.TrainingJob) (map[string]interface{}, map[string]interface{}) {
	jobName := trainingjob.Name
	jobNamespace := trainingjob.Namespace
	launcherSpec := map[string]interface{}{}
	workerSpec := map[string]interface{}{}
	raw := trainingjob.Annotations
	if raw == nil {
		log.Warnf("get trainingjob: %v/%v annotations failed.", jobNamespace, jobName)
		return launcherSpec, workerSpec
	}
	var annotations map[string]interface{}
	val, ok := raw[etJobMetaDataAnnotationsKey]
	if !ok {
		return launcherSpec, workerSpec
	}
	err := json.Unmarshal([]byte(val), &annotations)
	if err != nil {
		log.Debugf("failed to parse etjob annotations,reason: %v", err)
		return launcherSpec, workerSpec
	}
	obj, ok := annotations["spec"]
	if !ok {
		log.Debugf("parse trainingjob(%v/%v) specs failed.", jobNamespace, jobName)
		return launcherSpec, workerSpec
	}
	spec := obj.(map[string]interface{})

	replicaSpecItems, ok := spec["etReplicaSpecs"]
	if !ok {
		log.Debugf("parse trainingjob(%v/%v) etReplicaSpecs failed.", jobNamespace, jobName)
		return launcherSpec, workerSpec
	}
	etReplicaSpecs := replicaSpecItems.(map[string]interface{})
	launcherSpecItem, ok := etReplicaSpecs["launcher"]
	if !ok {
		log.Debugf("parse trainingjob(%v/%v) launcherSpec failed.", jobNamespace, jobName)
		return launcherSpec, workerSpec
	}
	launcherSpec = launcherSpecItem.(map[string]interface{})
	workerSpecItem, ok := etReplicaSpecs["worker"]
	if !ok {
		log.Debugf("parse trainingjob(%v/%v) workerSpec failed.", jobNamespace, jobName)
		return map[string]interface{}{}, workerSpec
	}
	workerSpec = workerSpecItem.(map[string]interface{})
	return launcherSpec, workerSpec
}

// Get PriorityClass
func (ej *ETJob) GetPriorityClass() string {
	//log.Debugf("trainingjob: %v", ej.trainingjob)
	// can not get addr of TrainingJob.Spec.ETReplicaSpecs
	//log.Debugf("spec addr: %v", ej.trainingjob.Spec)

	launcher, worker := parseAnnotations(ej.trainingjob)
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

func getPodsOfETJob(job *v1alpha1.TrainingJob, ejt *ETJobTrainer, podList []*corev1.Pod) ([]*corev1.Pod, *corev1.Pod) {
	return getPodsOfTrainingJob(job.Name, job.Namespace, podList, ejt.isETPod, func(pod *corev1.Pod) bool {
		return ejt.isChiefPod(pod)
	})
}
