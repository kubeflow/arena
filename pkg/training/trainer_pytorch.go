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
	"fmt"
	"strings"

	commonv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/common/v1"

	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/operators/pytorch-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	pytorchv1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
)

const (
	// pytorch-operator added labels for pods and servers.
	pytorchReplicaTypeLabel  = "pytorch-replica-type"
	pytorchReplicaIndexLabel = "pytorch-replica-index"
	labelPyTorchGroupName    = "group-name"
	labelPyTorchJobName      = "pytorch-job-name"
	labelPyTorchJobRole      = "job-role"
)

// PyTorch Job Information
type PyTorchJob struct {
	*BasicJobInfo
	pytorchjob   *pytorchv1.PyTorchJob
	pods         []*v1.Pod // all the pods including statefulset and job
	chiefPod     *v1.Pod   // the master pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType // return trainer type: pytorchjob
}

func (pj *PyTorchJob) Name() string {
	return pj.name
}

func (pj *PyTorchJob) Uid() string {
	return string(pj.pytorchjob.UID)
}

// Get the master Pod of the Job.
func (pj *PyTorchJob) ChiefPod() *v1.Pod {
	return pj.chiefPod
}

func (pj *PyTorchJob) Trainer() types.TrainingJobType {
	return pj.trainerType
}

// Get all the pods of the Training Job
func (pj *PyTorchJob) AllPods() []*v1.Pod {
	return pj.pods
}

func (pj *PyTorchJob) GetTrainJob() interface{} {
	return pj.pytorchjob
}

func checkPyTorchStatus(status commonv1.JobStatus) commonv1.JobConditionType {
	t := commonv1.JobConditionType("Pending")
	for _, condition := range status.Conditions {
		if condition.Status == v1.ConditionTrue {
			t = condition.Type
			break
		}
	}
	return t
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (pj *PyTorchJob) GetStatus() (status string) {
	status = "PENDING"
	pytorchjob := pj.pytorchjob
	if pytorchjob.Name == "" {
		return status
	}

	p := checkStatus(pytorchjob.Status)
	if p == commonv1.JobCreated || p == commonv1.JobRestarting {
		status = "PENDING"
	} else {
		status = strings.ToUpper(string(p))
	}

	return status
}

// Get the start time
func (pj *PyTorchJob) StartTime() *metav1.Time {
	return &pj.pytorchjob.CreationTimestamp
}

// Get the Job Age
func (pj *PyTorchJob) Age() time.Duration {
	job := pj.pytorchjob

	// use creation timestamp
	if job.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.CreationTimestamp.Time)
}

// Get the Job Training Duration
func (pj *PyTorchJob) Duration() time.Duration {
	pytorchjob := pj.pytorchjob

	if pytorchjob.Status.StartTime == nil ||
		pytorchjob.Status.StartTime.IsZero() {
		return 0
	}

	if !pytorchjob.Status.CompletionTime.IsZero() {
		return pytorchjob.Status.CompletionTime.Time.Sub(pytorchjob.Status.StartTime.Time)
	}

	if pj.GetStatus() == "FAILED" {
		cond := getPodLatestCondition(pj.chiefPod)
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Time.Sub(pytorchjob.CreationTimestamp.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", pj.chiefPod.Name)
		}
	}

	return metav1.Now().Sub(pytorchjob.Status.StartTime.Time)
}

// Get Dashboard url of the job
func (pj *PyTorchJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
	urls := []string{}
	// dashboardURL, err := dashboard(client, "kubeflow", "tf-job-dashboard")
	dashboardURL, err := dashboard(client, namespace, "kubernetes-dashboard")

	if err != nil {
		log.Debugf("Get dashboard failed due to %v", err)
		// retry for the existing customers, will be deprecated in the future
		dashboardURL, err = dashboard(client, arenaNamespace, "kubernetes-dashboard")
		if err != nil {
			log.Debugf("Get dashboard failed due to %v", err)
		}
	}

	if err != nil {
		log.Debugf("Get dashboard failed due to %v", err)
		// retry for the existing customers, will be deprecated in the future
		dashboardURL, err = dashboard(client, "kube-system", "kubernetes-dashboard")
		if err != nil {
			log.Debugf("Get dashboard failed due to %v", err)
		}
	}

	if dashboardURL == "" {
		return urls, fmt.Errorf("No LOGVIEWER Installed.")
	}

	if len(pj.chiefPod.Spec.Containers) == 0 {
		return urls, fmt.Errorf("pytorch launcher is not ready!")
	}

	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		pj.chiefPod.Namespace,
		pj.chiefPod.Name,
		pj.chiefPod.Spec.Containers[0].Name,
		pj.chiefPod.Namespace)

	urls = append(urls, url)

	return urls, nil
}

// Requested GPU count of the Job
func (pj *PyTorchJob) RequestedGPU() int64 {
	if pj.requestedGPU > 0 {
		return pj.requestedGPU
	}
	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(pj.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}
	for _, pod := range pj.pods {
		pj.requestedGPU += gpuInPod(*pod)
	}
	return pj.requestedGPU
}

// Requested GPU count of the Job
func (pj *PyTorchJob) AllocatedGPU() int64 {
	if pj.allocatedGPU > 0 {
		return pj.allocatedGPU
	}
	for _, pod := range pj.pods {
		pj.allocatedGPU += gpuInActivePod(*pod)
	}
	return pj.allocatedGPU
}

// Get the hostIP of the master Pod
func (pj *PyTorchJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if pj.GetStatus() == "RUNNING" {
		hostIP = pj.chiefPod.Status.HostIP
	}

	return hostIP
}

func (pj *PyTorchJob) Namespace() string {
	return pj.pytorchjob.Namespace
}

// PyTorch Job trainer
type PyTorchJobTrainer struct {
	client           *kubernetes.Clientset
	pytorchjobClient *versioned.Clientset
	trainerType      types.TrainingJobType
	// check if it's enabled
	enabled bool
}

// NewPyTorchJobTrainer
func NewPyTorchJobTrainer() Trainer {
	// get pytorch operator client call pytorch operator api
	enable := false
	pytorchjobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	_, err := config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.PytorchCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("PytorchJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("PytorchJobTrainer is disabled,reason: %v", err)
	}
	log.Debugf("Succeed to init PytorchJobTrainer")
	return &PyTorchJobTrainer{
		pytorchjobClient: pytorchjobClient,
		client:           config.GetArenaConfiger().GetClientSet(),
		trainerType:      types.PytorchTrainingJob,
		enabled:          enable,
	}
}

// IsEnabled is used to get the trainer is enable or not
func (tt *PyTorchJobTrainer) IsEnabled() bool {
	return tt.enabled
}

// Get the type
func (tt *PyTorchJobTrainer) Type() types.TrainingJobType {
	return tt.trainerType
}

// check if it's TensorFlow job
func (tt *PyTorchJobTrainer) IsSupported(name, ns string) bool {
	if !tt.enabled {
		return false
	}
	isPyTorchJob := false
	_, err := tt.GetTrainingJob(name, ns)
	if err != nil {
		return isPyTorchJob
	}
	return !isPyTorchJob
}

// Get the training job from cache or directly

func (tt *PyTorchJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	pytorchjob, err := k8saccesser.GetK8sResourceAccesser().GetPytorchJob(tt.pytorchjobClient, namespace, name)
	if err != nil {
		return nil, err
	}
	if err := CheckJobIsOwnedByTrainer(pytorchjob.Labels); err != nil {
		return nil, err
	}
	// 2. Find the pod list, and determine the pod of the job
	allPods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("release=%v,app=%v", name, tt.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	batchJobs, err := k8saccesser.GetK8sResourceAccesser().ListBatchJobs(namespace, "pytorch_job_name")
	if err != nil {
		return nil, err
	}
	statefulsets, err := k8saccesser.GetK8sResourceAccesser().ListStatefulSets(namespace, "pytorch_job_name")
	if err != nil {
		return nil, err
	}
	// important for getting pytorchjob status
	pytorchjob.Status.Conditions = makeJobStatusSortedByTime(pytorchjob.Status.Conditions)
	pods, chiefPod := getPodsOfPyTorchJob(tt, pytorchjob, allPods)
	// 3. Find the other resources, like statefulset,job
	return &PyTorchJob{
		BasicJobInfo: &BasicJobInfo{
			resources: tt.resources(statefulsets, batchJobs, name, namespace, pods),
			name:      name,
		},
		pytorchjob:  pytorchjob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: tt.Type(),
	}, nil

}

func (tt *PyTorchJobTrainer) isChiefPod(pytorchjob *pytorchv1.PyTorchJob, item *v1.Pod) bool {
	if item.Labels[pytorchReplicaTypeLabel] != "master" {
		return false
	}
	log.Debugf("the pytorchjob %s with labels master", item.Name)
	return true
}

// check Labels: release==pytorchjob.name/app=="pytorchjob", namespace
func (tt *PyTorchJobTrainer) isPyTorchJob(name, ns string, item *pytorchv1.PyTorchJob) bool {
	if item.Namespace != ns {
		return false
	}
	if item.Labels["release"] != name {
		return false
	}
	if item.Labels["app"] != string(tt.trainerType) {
		return false
	}
	return true
}

// Determine whether it is a pod of pytorchjobs submitted by Arena
// check pod label: release==pytorchjob.name/app=="pytorchjob"/group-name=='kubeflow.org', namespace
func (tt *PyTorchJobTrainer) isPyTorchPod(name, ns string, pod *v1.Pod) bool {
	return utils.IsPyTorchPod(name, ns, pod)
}

func (tt *PyTorchJobTrainer) resources(statefulsets []*appv1.StatefulSet, batchJobs []*batchv1.Job, name string, namespace string, pods []*v1.Pod) []Resource {
	resources := []Resource{}
	// 2. Find the pod list, and determine the pod of the job
	for _, sts := range statefulsets {
		if sts.Labels["pytorch_job_name"] != name {
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
		if job.Labels["pytorch_job_name"] != name {
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

func (tt *PyTorchJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	jobLabels := GetTrainingJobLabels(tt.Type())
	// list all jobs from k8s apiserver
	pytorchjobs, err := k8saccesser.GetK8sResourceAccesser().ListPytorchJobs(tt.pytorchjobClient, namespace, jobLabels)
	if err != nil {
		return trainingJobs, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", tt.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	for _, pyjob := range pytorchjobs {
		pyjob.Status.Conditions = makeJobStatusSortedByTime(pyjob.Status.Conditions)
		filterPods, chiefPod := getPodsOfPyTorchJob(tt, pyjob, pods)
		trainingJobs = append(trainingJobs, &PyTorchJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      pyjob.Name,
			},
			pytorchjob:  pyjob,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: tt.Type(),
		})
	}
	return trainingJobs, nil
}

// Get PriorityClass
func (p *PyTorchJob) GetPriorityClass() string {
	pc := ""
	log.Debugf("pytorchjob: %v", p.pytorchjob)
	specs := p.pytorchjob.Spec.PyTorchReplicaSpecs
	log.Debugf("specs %v", specs)
	for _, spec := range specs {
		if spec.Template.Spec.PriorityClassName != "" {
			pc = spec.Template.Spec.PriorityClassName
			break
		}
	}

	return pc
}

// filter out all pods and chief pod (master pod) of pytorchjob from pods in current system
func getPodsOfPyTorchJob(tt *PyTorchJobTrainer, pytorchjob *pytorchv1.PyTorchJob, podList []*v1.Pod) ([]*v1.Pod, *v1.Pod) {
	return getPodsOfTrainingJob(pytorchjob.Name, pytorchjob.Namespace, podList, tt.isPyTorchPod, func(pod *v1.Pod) bool {
		return tt.isChiefPod(pytorchjob, pod)
	})
}
