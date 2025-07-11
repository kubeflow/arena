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
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	pytorchv1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
	"github.com/kubeflow/arena/pkg/operators/pytorch-operator/client/clientset/versioned"
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
	pods         []*corev1.Pod // all the pods including statefulset and job
	chiefPod     *corev1.Pod   // the master pod
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
func (pj *PyTorchJob) ChiefPod() *corev1.Pod {
	return pj.chiefPod
}

func (pj *PyTorchJob) Trainer() types.TrainingJobType {
	return pj.trainerType
}

// Get all the pods of the Training Job
func (pj *PyTorchJob) AllPods() []*corev1.Pod {
	return pj.pods
}

func (pj *PyTorchJob) GetTrainJob() interface{} {
	return pj.pytorchjob
}

func (pj *PyTorchJob) GetLabels() map[string]string {
	return pj.pytorchjob.Labels
}

// GetStatus returns the status of the Job i.e. QUEUING, PENDING, RUNNING, SUCCEEDED and FAILED.
func (pj *PyTorchJob) GetStatus() string {
	status := string(types.TrainingJobPending)
	defer log.Debugf("Get status of PyTorchJob %s: %s", pj.pytorchjob.Name, status)

	if pj.pytorchjob.Name == "" {
		return status
	}

	status = getStatus(pj.pytorchjob.Status)
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
		return pytorchjob.Status.CompletionTime.Sub(pytorchjob.Status.StartTime.Time)
	}

	if pj.GetStatus() == "FAILED" {
		cond := getPodLatestCondition(pj.chiefPod)
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Sub(pytorchjob.CreationTimestamp.Time)
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
		return urls, fmt.Errorf("no LOGVIEWER Installed")
	}

	if len(pj.chiefPod.Spec.Containers) == 0 {
		return urls, fmt.Errorf("pytorch launcher is not ready")
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
	// important for getting pytorchjob status
	pytorchjob.Status.Conditions = makeJobStatusSortedByTime(pytorchjob.Status.Conditions)
	pods, chiefPod := getPodsOfPyTorchJob(tt, pytorchjob, allPods)
	// 3. Find the other resources, like statefulset,job
	return &PyTorchJob{
		BasicJobInfo: &BasicJobInfo{
			resources: tt.resources(name, namespace, pods),
			name:      name,
		},
		pytorchjob:  pytorchjob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: tt.Type(),
	}, nil

}

func (tt *PyTorchJobTrainer) isChiefPod(pytorchjob *pytorchv1.PyTorchJob, item *corev1.Pod) bool {
	isChiefPod := false

	if val, ok := item.Labels[pytorchReplicaTypeLabel]; ok && val == "master" {
		isChiefPod = true
	}

	if val, ok := item.Labels[TrainingReplicaTypeLabel]; ok && val == "master" {
		isChiefPod = true
	}

	return isChiefPod
}

// Determine whether it is a pod of pytorchjobs submitted by Arena
// check pod label: release==pytorchjob.name/app=="pytorchjob"/group-name=='kubeflow.org', namespace
func (tt *PyTorchJobTrainer) isPyTorchPod(name, ns string, pod *corev1.Pod) bool {
	return utils.IsPyTorchPod(name, ns, pod)
}

func (tt *PyTorchJobTrainer) resources(name string, namespace string, pods []*corev1.Pod) []Resource {
	resources := podResources(pods)
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
func getPodsOfPyTorchJob(tt *PyTorchJobTrainer, pytorchjob *pytorchv1.PyTorchJob, podList []*corev1.Pod) ([]*corev1.Pod, *corev1.Pod) {
	return getPodsOfTrainingJob(pytorchjob.Name, pytorchjob.Namespace, podList, tt.isPyTorchPod, func(pod *corev1.Pod) bool {
		return tt.isChiefPod(pytorchjob, pod)
	})
}
