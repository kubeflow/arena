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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/arenacache"
	"github.com/kubeflow/arena/pkg/operators/pytorch-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
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
	PytorchCRD               = "pytorchjobs.kubeflow.org"
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
	log.Debugf("Init PyTorch job trainer")
	// get pytorch operator client call pytorch operator api
	enable := false
	pytorchjobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	// this step is used to check operator is installed or not
	for _, crdName := range config.GetArenaConfiger().GetClusterInstalledCRDs() {
		if crdName == PytorchCRD {
			enable = true
			break
		}
	}
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
	pytorchjob := &pytorchv1.PyTorchJob{}
	var err error
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, pytorchjob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`PyTorchJob.kubeflow.org "%v" not found`, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find pytorchjob %v from cache,reason: %v", name, err)
		}

	} else {
		pytorchjob, err = tt.pytorchjobClient.KubeflowV1().PyTorchJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, PytorchCRD, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find pytorchjob %v from api server,reason: %v", name, err)
		}
	}
	// 2. Find the pod list, and determine the pod of the job
	labels := map[string]string{
		"release": name,
		"app":     string(tt.Type()),
	}
	podList, err := listJobPods(tt.client, namespace, labels)
	if err != nil {
		return nil, err
	}
	allPods := []*v1.Pod{}
	for _, pod := range podList.Items {
		allPods = append(allPods, pod.DeepCopy())
	}
	// important for getting pytorchjob status
	pytorchjob.Status.Conditions = makeJobStatusSortedByTime(pytorchjob.Status.Conditions)
	pods, chiefPod := getPodsOfPyTorchJob(tt, pytorchjob, allPods)
	// 3. Find the other resources, like statefulset,job
	resources, err := tt.resources(name, namespace, pods)
	if err != nil {
		return nil, err
	}
	return &PyTorchJob{
		BasicJobInfo: &BasicJobInfo{
			resources: resources,
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

func (tt *PyTorchJobTrainer) resources(name string, namespace string, pods []*v1.Pod) ([]Resource, error) {
	resources := []Resource{}
	labels := map[string]string{
		"pytorch_job_name": name,
	}
	// 2. Find the pod list, and determine the pod of the job
	stsList, err := listStatefulSets(tt.client, namespace, labels)
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
	jobList, err := listJobBatchJobs(tt.client, namespace, labels)
	if err != nil {
		return resources, err
	}
	for _, job := range jobList.Items {
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

func (tt *PyTorchJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	// list all jobs from k8s apiserver
	pyjobList, err := tt.listJobs(namespace)
	if err != nil {
		return trainingJobs, err
	}
	for _, job := range pyjobList.Items {
		pyjob := job.DeepCopy()
		labels := map[string]string{
			"release": pyjob.Name,
			"app":     string(tt.Type()),
		}
		podList, err := listJobPods(tt.client, pyjob.Namespace, labels)
		if err != nil {
			log.Errorf("failed to get pods of job %v,reason: %v", pyjob.Name, err)
			continue
		}
		pods := []*v1.Pod{}
		for _, pod := range podList.Items {
			pods = append(pods, pod.DeepCopy())
		}
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

func (tt *PyTorchJobTrainer) listJobs(namespace string) (*pytorchv1.PyTorchJobList, error) {
	if config.GetArenaConfiger().IsDaemonMode() {
		list := &pytorchv1.PyTorchJobList{}
		return list, arenacache.GetCacheClient().ListTrainingJobs(list, namespace)
	}
	return tt.pytorchjobClient.KubeflowV1().PyTorchJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release"),
	})
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
