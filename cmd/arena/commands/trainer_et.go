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

package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AliyunContainerService/et-operator/pkg/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/AliyunContainerService/et-operator/pkg/api/v1alpha1"
)

const (
	// et-operator added key of labels for pods.
	etLabelGroupName       = "group-name"
	etLabelTrainingJobName = "training-job-name"
	etLabelTrainingJobRole = "training-job-role"

	etJobMetaDataAnnotationsKey = "kubectl.kubernetes.io/last-applied-configuration"
)

var (
	allEtJobs []v1alpha1.TrainingJob
)

func initETJobClient() (jobClientset *versioned.Clientset, err error) {
	if restConfig == nil {
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	// create the jobClientset
	jobClientset = versioned.NewForConfigOrDie(restConfig)
	return jobClientset, nil
}

// ET Job Information
type ETJob struct {
	*BasicJobInfo
	trainingjob  v1alpha1.TrainingJob
	pods         []v1.Pod // all the pods including statefulset and job
	chiefPod     v1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  string // return trainer type: TENSORFLOW
}

func (ej *ETJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
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
func (ej *ETJob) ChiefPod() v1.Pod {
	return ej.chiefPod
}

func (ej *ETJob) Trainer() string {
	return ej.trainerType
}

// Get all the pods of the Training Job
func (ej *ETJob) AllPods() []v1.Pod {
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
			return cond.LastTransitionTime.Time.Sub(trainingjob.CreationTimestamp.Time)
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
	for _, pod := range ej.pods {
		ej.requestedGPU += gpuInPod(pod)
	}
	return ej.requestedGPU
}

// Requested GPU count of the Job
func (ej *ETJob) AllocatedGPU() int64 {
	if ej.allocatedGPU > 0 {
		return ej.allocatedGPU
	}
	for _, pod := range ej.pods {
		ej.allocatedGPU += gpuInActivePod(pod)
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

func (ej *ETJob) GetWorkerMaxReplicas(maxWorkers int) interface{} {
	_, worker := parseAnnotations(ej.trainingjob)
	log.Infof("worker: %v", worker)
	if worker != nil {
		if _, ok := worker["maxReplicas"]; ok {
			maxWorkers = int(worker["maxReplicas"].(float64))
		}
	}
	return maxWorkers
}

func (ej *ETJob) GetWorkerMinReplicas(minWorkers int) interface{} {
	_, worker := parseAnnotations(ej.trainingjob)
	log.Infof("worker: %v", worker)
	if worker != nil {
		if _, ok := worker["minReplicas"]; ok {
			minWorkers = int(worker["minReplicas"].(float64))
		}
	}
	return minWorkers
}

// ET Job trainer
type ETJobTrainer struct {
	client      *kubernetes.Clientset
	jobClient   *versioned.Clientset
	trainerType string
	// check if it's enabled
	enabled bool
}

// NewETJobTrainer
func NewETJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init ET job trainer")
	jobClient, err := initETJobClient()
	if err != nil {
		log.Debugf("unsupported etjob due to %v", err)
		return &ETJobTrainer{
			trainerType: "etjob",
			enabled:     false,
		}
	}
	// allPods have been cached, we do the same to allEtJobs
	if useCache {
		ns := namespace
		if allNamespaces {
			ns = metav1.NamespaceAll
		}

		jobList, err := jobClient.EtV1alpha1().TrainingJobs(ns).List(metav1.ListOptions{})
		if err != nil {
			log.Debugf("unsupported etjob due to %v", err)
			return &ETJobTrainer{
				trainerType: "etjob",
				enabled:     false,
			}
		}

		for _, job := range jobList.Items {
			allEtJobs = append(allEtJobs, job)
		}
	}

	return &ETJobTrainer{
		jobClient:   jobClient,
		client:      client,
		trainerType: "etjob",
		enabled:     true,
	}
}

// Get the type
func (ejt *ETJobTrainer) Type() string {
	return ejt.trainerType
}

// check if it's TensorFlow job
func (ejt *ETJobTrainer) IsSupported(name, ns string) bool {
	if !ejt.enabled {
		return false
	}

	isET := false

	if useCache {
		for _, job := range allEtJobs {
			if ejt.isETJob(name, ns, job) {
				isET = true
				log.Debugf("the job %s for %s in namespace %s is found.", job.Name, name, ns)
				break
			}
		}
	} else {
		etjobList, err := ejt.jobClient.EtV1alpha1().TrainingJobs(ns).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("release=%s", name),
		})

		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(etjobList.Items) > 0 {
			isET = true
		}
	}

	return isET
}

// Get the training job from cache or directly
func (ejt *ETJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	if len(allEtJobs) > 0 {
		tj, err = ejt.getTrainingJobFromCache(name, namespace)
	} else {
		tj, err = ejt.getTrainingJob(name, namespace)
	}

	return tj, err
}

func (ejt *ETJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		trainingjob v1alpha1.TrainingJob
	)

	// 1. Get the batchJob of training Job
	trainingjobList, err := ejt.jobClient.EtV1alpha1().TrainingJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})
	if err != nil {
		return nil, err
	}
	if len(trainingjobList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		trainingjob = trainingjobList.Items[0]
	}

	// 2. Find the pod list, and determine the pod of the job
	podList, err := ejt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, err
	}

	pods, chiefPod := getPodsOfETJob(name, ejt, podList.Items)

	// 3. Find the other resources, like statefulset,job
	resources, err := ejt.resources(name, namespace, pods)
	if err != nil {
		return nil, err
	}
	return &ETJob{
		BasicJobInfo: &BasicJobInfo{
			resources: resources,
			name:      name,
		},
		trainingjob: trainingjob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: ejt.Type(),
	}, nil

}

// Get the training job from Cache
func (ejt *ETJobTrainer) getTrainingJobFromCache(name, ns string) (TrainingJob, error) {

	var (
		trainingjob v1alpha1.TrainingJob
	)

	// 1. Find the job from select etjobs in current system
	// every time call NewETTrainer() will get allEtJobs
	for _, item := range allEtJobs {
		if ejt.isETJob(name, ns, item) {
			trainingjob = item
			break
		}
	}

	// 2. Find the pods, and determine the pod of the job
	pods, chiefPod := getPodsOfETJob(name, ejt, allPods)

	return &ETJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(pods),
			name:      name,
		},
		trainingjob: trainingjob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: ejt.Type(),
	}, nil
}

func (ejt *ETJobTrainer) isChiefPod(item v1.Pod) bool {

	if val, ok := item.Labels[etLabelTrainingJobRole]; ok && (val == "launcher") {
		log.Debugf("the pod %s with labels %s", item.Name, val)
	} else {
		return false
	}

	return true
}

func (ejt *ETJobTrainer) isETJob(name, ns string, item v1alpha1.TrainingJob) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the etjob: %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "etjob") {
		log.Debugf("the etjob: %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func (ejt *ETJobTrainer) isETPod(name, ns string, item v1.Pod) bool {
	log.Debugf("pod.name: %s: %v", item.Name, item.Labels)
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the etjob pod: %s with labels['release'] %s is found.", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "etjob") {
		log.Debugf("the etjob pod: %s with labels['app'] %s is found.", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels[etLabelGroupName]; ok && (val == "kai.alibabacloud.com") {
		log.Debugf("the etjob pod: %s with labels[%s] %s is found.", item.Name, etLabelGroupName, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func (ejt *ETJobTrainer) resources(name string, namespace string, pods []v1.Pod) ([]Resource, error) {
	var resources []Resource

	// 2. Find the pod list, and determine the pod of the job
	stsList, err := ejt.client.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("%s=%s", etLabelTrainingJobName, name),
	})
	if err != nil {
		return resources, err
	}
	for _, sts := range stsList.Items {
		resources = append(resources, Resource{
			Name:         sts.Name,
			Uid:          string(sts.UID),
			ResourceType: ResourceTypeStatefulSet,
		})
	}

	// 2. Find the pod list, and determine the pod of the job
	jobs, err := ejt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("%s=%s", etLabelTrainingJobName, name),
	})
	if err != nil {
		return resources, err
	}
	for _, job := range jobs.Items {
		resources = append(resources, Resource{
			Name:         job.Name,
			Uid:          string(job.UID),
			ResourceType: ResourceTypeJob,
		})
	}
	resources = append(resources, podResources(pods)...)
	return resources, nil
}

/**
* List Training jobs
 */
func (ejt *ETJobTrainer) ListTrainingJobs() (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	var jobInfos []types.TrainingJobInfo
	for _, etjob := range allEtJobs {
		jobInfo := types.TrainingJobInfo{}
		log.Debugf("find etjob %s in %s", etjob.Name, etjob.Namespace)
		if val, ok := etjob.Labels["release"]; ok && (etjob.Name == fmt.Sprintf("%s-%s", val, ejt.Type())) {
			log.Debugf("the etjob %s with labels %s found in List", etjob.Name, val)
			jobInfo.Name = val
		} else {
			jobInfo.Name = etjob.Name
		}

		jobInfo.Namespace = etjob.Namespace
		jobInfos = append(jobInfos, jobInfo)
		// jobInfos = append(jobInfos, types.TrainingJobInfo{Name: mpijob.})
	}
	log.Debugf("jobInfos %v", jobInfos)

	for _, jobInfo := range jobInfos {
		job, err := ejt.getTrainingJobFromCache(jobInfo.Name, jobInfo.Namespace)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (ej *ETJob) isSucceeded() bool {
	// status.ETJobLauncherStatusType
	return ej.trainingjob.Status.Phase == "Succeeded"
}

func (ej *ETJob) isFailed() bool {
	return ej.trainingjob.Status.Phase == "Failed"
}

func (ej *ETJob) isScaling() bool {
	return ej.trainingjob.Status.Phase == "Scaling"
}

func (ej *ETJob) isPending() bool {

	if len(ej.chiefPod.Name) == 0 || ej.chiefPod.Status.Phase == v1.PodPending {
		log.Debugf("The ETJob is pending due to chiefPod is not ready")
		return true
	}

	return false
}

func parseAnnotations(trainingjob v1alpha1.TrainingJob) (launcherSpec map[string]interface{}, workerSpec map[string]interface{}) {
	jobName := trainingjob.Name
	jobNamespace := trainingjob.Namespace
	raw := trainingjob.Annotations
	if raw == nil {
		log.Warnf("get trainingjob: %v/%v annotations failed.", jobNamespace, jobName)
		return nil, nil
	}

	var annotations map[string]interface{}
	if temp, ok := raw[etJobMetaDataAnnotationsKey]; ok {
		err := json.Unmarshal([]byte(temp), &annotations)
		if err != nil {
			log.Warnf("json Unmarshal error: %v", err.Error())
			return
		}
		if _, ok := annotations["spec"]; ok {
			spec := annotations["spec"].(map[string]interface{})
			if _, ok := spec["etReplicaSpecs"]; ok {
				etReplicaSpecs := spec["etReplicaSpecs"].(map[string]interface{})
				if _, ok := etReplicaSpecs["launcher"]; ok {
					launcherSpec = etReplicaSpecs["launcher"].(map[string]interface{})
				} else {
					log.Warnf("parse trainingjob(%v/%v) launcherSpec failed.", jobNamespace, jobName)
				}
				if _, ok := etReplicaSpecs["worker"]; ok {
					workerSpec = etReplicaSpecs["worker"].(map[string]interface{})
				} else {
					log.Warnf("parse trainingjob(%v/%v) workerSpec failed.", jobNamespace, jobName)
				}
			} else {
				log.Warnf("parse trainingjob(%v/%v) etReplicaSpecs failed.", jobNamespace, jobName)
			}
		} else {
			log.Warnf("parse trainingjob(%v/%v) specs failed.", jobNamespace, jobName)
		}
	} else {
		log.Warnf("parse trainingjob(%v/%v) metadata.annotations[%v] failed.", jobNamespace, jobName, etJobMetaDataAnnotationsKey)
	}
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

func getPodsOfETJob(name string, ejt *ETJobTrainer, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
	pods = []v1.Pod{}
	for _, item := range podList {
		if !ejt.isETPod(name, namespace, item) {
			continue
		}
		if ejt.isChiefPod(item) && item.CreationTimestamp.After(chiefPod.CreationTimestamp.Time) {
			// If there are some failed chiefPod, and the new chiefPod haven't started, set the latest failed pod as chief pod
			if chiefPod.Name != "" && item.Status.Phase == v1.PodPending {
				continue
			}
			chiefPod = item
		}

		// for non-job pod, add it into the pod list
		pods = append(pods, item)
		log.Debugf("add pod %v to pods", item)
	}
	return pods, chiefPod
}
