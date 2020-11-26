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
	"errors"
	"fmt"
	"strings"

	commonv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/common/v1"

	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	apitypes "github.com/kubeflow/arena/pkg/apis/types"
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
)

var (
	errPytorchJobNotFound = errors.New("pytorchjob not found")
)

// PyTorch Job Information
type PyTorchJob struct {
	*BasicJobInfo
	pytorchjob   *pytorchv1.PyTorchJob
	pods         []*v1.Pod // all the pods including statefulset and job
	chiefPod     *v1.Pod   // the master pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  string // return trainer type: pytorchjob
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

func (pj *PyTorchJob) Trainer() string {
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
	trainerType      string
	// check if it's enabled
	enabled bool
}

// NewPyTorchJobTrainer
func NewPyTorchJobTrainer() Trainer {
	log.Debugf("Init PyTorch job trainer")
	// get pytorch operator client call pytorch operator api
	pytorchjobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	return &PyTorchJobTrainer{
		pytorchjobClient: pytorchjobClient,
		client:           config.GetArenaConfiger().GetClientSet(),
		trainerType:      string(apitypes.PytorchTrainingJob),
		enabled:          true,
	}
}

// Get the type
func (tt *PyTorchJobTrainer) Type() string {
	return tt.trainerType
}

// check if it's TensorFlow job
func (tt *PyTorchJobTrainer) IsSupported(name, ns string) bool {
	if !tt.enabled {
		return false
	}
	isPyTorchJob := false
	if config.GetArenaConfiger().IsDaemonMode() {
		_, err := tt.getTrainingJobFromCache(name, ns)
		if err != nil {
			return isPyTorchJob
		}
		return !isPyTorchJob
	}
	_, err := tt.getTrainingJob(name, ns)
	if err != nil {
		return isPyTorchJob
	}
	return !isPyTorchJob
}

// Get the training job from cache or directly
func (tt *PyTorchJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	// if arena is daemon mode,get job from cache
	if config.GetArenaConfiger().IsDaemonMode() {
		return tt.getTrainingJobFromCache(name, namespace)
	}
	// get job from api server
	return tt.getTrainingJob(name, namespace)
}

// getTrainingJob gets job from api server
func (tt *PyTorchJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	pytorchjob, err := tt.pytorchjobClient.KubeflowV1().PyTorchJobs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(`pytorchjobs.kubeflow.org "%v" not found`, name)) {
			return nil, errPytorchJobNotFound
		}
		return nil, fmt.Errorf("failed to find job %v,reason: %v", name, err)
	}
	// 2. Find the pod list, and determine the pod of the job
	podList, err := tt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s,app=%v", name, apitypes.PytorchTrainingJob),
	})
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

// Get the training job from Cache
func (tt *PyTorchJobTrainer) getTrainingJobFromCache(name, namespace string) (TrainingJob, error) {
	pyjob, pods := arenacache.GetArenaCache().GetPytorchJob(namespace, name)
	if pyjob == nil {
		return nil, errPytorchJobNotFound
	}
	pyjob.Status.Conditions = makeJobStatusSortedByTime(pyjob.Status.Conditions)
	filterPods, chiefPod := getPodsOfPyTorchJob(tt, pyjob, pods)
	return &PyTorchJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(filterPods),
			name:      name,
		},
		pytorchjob:  pyjob,
		chiefPod:    chiefPod,
		pods:        filterPods,
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
	if item.Labels["app"] != string(apitypes.PytorchTrainingJob) {
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

	// 2. Find the pod list, and determine the pod of the job
	stsList, err := tt.client.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("pytorch_job_name=%s", name),
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
	jobs, err := tt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("pytorch_job_name=%s", name),
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
func (tt *PyTorchJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) (jobs []TrainingJob, err error) {
	// if arena is configured as daemon,getting all tfjobs from cache is corrent
	if config.GetArenaConfiger().IsDaemonMode() {
		return tt.listFromCache(namespace, allNamespace)
	}
	return tt.listFromAPIServer(namespace, allNamespace)
}

func (tt *PyTorchJobTrainer) listFromAPIServer(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	// list all jobs from k8s apiserver
	pyjobList, err := tt.pytorchjobClient.KubeflowV1().PyTorchJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release"),
	})
	if err != nil {
		return trainingJobs, err
	}
	for _, job := range pyjobList.Items {
		pyjob := job.DeepCopy()
		podList, err := tt.client.CoreV1().Pods(pyjob.Namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: fmt.Sprintf("release=%s,app=%v", pyjob.Name, apitypes.TFTrainingJob),
		})
		if err != nil {
			log.Errorf("failed to get pods of job %v", pyjob.Name)
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

func (tt *PyTorchJobTrainer) listFromCache(namespace string, allNamespace bool) ([]TrainingJob, error) {
	filter := func(job *pytorchv1.PyTorchJob) bool { return job.Namespace == namespace }
	trainingJobs := []TrainingJob{}
	if allNamespace {
		filter = func(job *pytorchv1.PyTorchJob) bool { return true }
	}
	jobs, pods := arenacache.GetArenaCache().FilterPytorchJobs(filter)
	for jobKey, pyjob := range jobs {
		pyjob.Status.Conditions = makeJobStatusSortedByTime(pyjob.Status.Conditions)
		filterPods, chiefPod := getPodsOfPyTorchJob(tt, pyjob, pods[jobKey])
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
