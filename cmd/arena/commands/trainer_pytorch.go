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
	"fmt"
	commonv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/common/v1"
	"strings"

	"github.com/kubeflow/arena/pkg/operators/pytorch-operator/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"time"

	pytorchv1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
)

const (
	// pytorch-operator added labels for pods and servers.
	pytorchReplicaTypeLabel     = "pytorch-replica-type"
	pytorchReplicaIndexLabel    = "pytorch-replica-index"
	labelPyTorchGroupName         = "group-name"
	labelPyTorchJobName         = "pytorch-job-name"
	labelPyTorchJobRole         = "job-role"
)

var (
	allPyTorchJobs []pytorchv1.PyTorchJob
)

func initPyTorchJobClient() (pytorchjobClientset *versioned.Clientset, err error) {
	if restConfig == nil {
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	// create the pytorchjobClientset
	pytorchjobClientset = versioned.NewForConfigOrDie(restConfig)
	return pytorchjobClientset, nil
}

// PyTorch Job Information
type PyTorchJob struct {
	*BasicJobInfo
	pytorchjob       pytorchv1.PyTorchJob
	pods         []v1.Pod // all the pods including statefulset and job
	chiefPod     v1.Pod   // the master pod
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
func (pj *PyTorchJob) ChiefPod() v1.Pod {
	return pj.chiefPod
}

func (pj *PyTorchJob) Trainer() string {
	return pj.trainerType
}

// Get all the pods of the Training Job
func (pj *PyTorchJob) AllPods() []v1.Pod {
	return pj.pods
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
func (pj *PyTorchJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
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
		pj.requestedGPU += gpuInPod(pod)
	}
	return pj.requestedGPU
}

// Requested GPU count of the Job
func (pj *PyTorchJob) AllocatedGPU() int64 {
	if pj.allocatedGPU > 0 {
		return pj.allocatedGPU
	}
	for _, pod := range pj.pods {
		pj.allocatedGPU += gpuInActivePod(pod)
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
	client       *kubernetes.Clientset
	pytorchjobClient *versioned.Clientset
	trainerType  string
	// check if it's enabled
	enabled bool
}

// NewPyTorchJobTrainer
func NewPyTorchJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init PyTorch job trainer")
	// get pytorch operator client call pytorch operator api
	pytorchjobClient, err := initPyTorchJobClient()
	if err != nil {
		log.Debugf("unsupported pytorchjob due to %v", err)
		return &PyTorchJobTrainer{
			trainerType: "pytorchjob",
			enabled:     false,
		}
	}
	// allPods have been cached, we do the same to allPyTorchJobs
	if useCache {
		ns := namespace
		if allNamespaces {
			ns = metav1.NamespaceAll
		}

		pytorchjobList, err := pytorchjobClient.KubeflowV1().PyTorchJobs(ns).List(metav1.ListOptions{})
		if err != nil {
			log.Debugf("unsupported pytorchjob due to %v", err)
			return &PyTorchJobTrainer{
				trainerType: "pytorchjob",
				enabled:     false,
			}
		}


		for _, pytorchjob := range pytorchjobList.Items {
			allPyTorchJobs = append(allPyTorchJobs, pytorchjob)
		}
		log.Debugf("allPyTorchJobs: %v", allPyTorchJobs)

	}

	return &PyTorchJobTrainer{
		pytorchjobClient: pytorchjobClient,
		client:       client,
		trainerType:  "pytorchjob",
		enabled:      true,
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

	isPyTorch := false

	if useCache {
		for _, job := range allPyTorchJobs {
			if tt.isPyTorchJob(name, ns, job) {
				isPyTorch = true
				log.Debugf("the job %s for %s in namespace %s is found.", job.Name, name, ns)
				break
			}
		}
	} else {
		pytorchjobList, err := tt.pytorchjobClient.KubeflowV1().PyTorchJobs(ns).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("release=%s", name),
		})

		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(pytorchjobList.Items) > 0 {
			isPyTorch = true
		}
	}

	return isPyTorch
}

// Get the training job from cache or directly
func (tt *PyTorchJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	if len(allPyTorchJobs) > 0 {
		tj, err = tt.getTrainingJobFromCache(name, namespace)
	} else {
		tj, err = tt.getTrainingJob(name, namespace)
	}

	return tj, err
}

func (tt *PyTorchJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		pytorchjob pytorchv1.PyTorchJob
	)

	// 1. Get the batchJob of training Job
	pytorchjobList, err := tt.pytorchjobClient.KubeflowV1().PyTorchJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})
	if err != nil {
		return nil, err
	}
	if len(pytorchjobList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		pytorchjob = pytorchjobList.Items[0]
	}

	// important for getting pytorchjob status
	pytorchjob.Status.Conditions = makeJobStatusSortedByTime(pytorchjob.Status.Conditions)

	// 2. Find the pod list, and determine the pod of the job
	podList, err := tt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, err
	}

	pods, chiefPod := getPodsOfPyTorchJob(name, tt, pytorchjob, podList.Items)

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
		pytorchjob:      pytorchjob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: tt.Type(),
	}, nil

}

// Get the training job from Cache
func (tt *PyTorchJobTrainer) getTrainingJobFromCache(name, ns string) (TrainingJob, error) {

	var (
		pytorchjob pytorchv1.PyTorchJob
	)

	// 1. Find the pytorch job from select pytorchjobs in current system
	// every time call NewPyTorchJobTrainer() will get allPyTorchJobs
	for _, item := range allPyTorchJobs {
		if tt.isPyTorchJob(name, ns, item) {
			pytorchjob = item
			break
		}
	}

	// important for getting pytorchjob status
	pytorchjob.Status.Conditions = makeJobStatusSortedByTime(pytorchjob.Status.Conditions)

	// 2. Find the pods, and determine the pod of the job
	// call arena list interface, will get allPods in current system
	pods, chiefPod := getPodsOfPyTorchJob(name, tt, pytorchjob, allPods)

	return &PyTorchJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(pods),
			name:      name,
		},
		pytorchjob:      pytorchjob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: tt.Type(),
	}, nil
}



func (tt *PyTorchJobTrainer) isChiefPod(pytorchjob pytorchv1.PyTorchJob, item v1.Pod) bool {

	if val, ok := item.Labels[pytorchReplicaTypeLabel]; ok && (val == "master") {
		log.Debugf("the pytorchjob %s with labels %s", item.Name, val)
	} else {
		return false
	}

	return true
}

// check Labels: release==pytorchjob.name/app=="pytorchjob", namespace
func (tt *PyTorchJobTrainer) isPyTorchJob(name, ns string, item pytorchv1.PyTorchJob) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the pytorchjob %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "pytorchjob") {
		log.Debugf("the pytorchjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

// Determine whether it is a pod of pytorchjobs submitted by Arena
// check pod label: release==pytorchjob.name/app=="pytorchjob"/group-name=='kubeflow.org', namespace
func (tt *PyTorchJobTrainer) isPyTorchPod(name, ns string, item v1.Pod) bool {
	log.Debugf("pod.name: %s: %v", item.Name, item.Labels)
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the pytorchjob %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "pytorchjob") {
		log.Debugf("the pytorchjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels[labelPyTorchGroupName]; ok && (val == "kubeflow.org") {
		log.Debugf("the pytorchjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func (tt *PyTorchJobTrainer) resources(name string, namespace string, pods []v1.Pod) ([]Resource, error) {
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
func (tt *PyTorchJobTrainer) ListTrainingJobs() (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	jobInfos := []types.TrainingJobInfo{}
	for _, pytorchjob := range allPyTorchJobs {
		jobInfo := types.TrainingJobInfo{}
		log.Debugf("find pytorchjob %s in %s", pytorchjob.Name, pytorchjob.Namespace)
		// TODO: why? seems like never step here
		if val, ok := pytorchjob.Labels["release"]; ok && (pytorchjob.Name == fmt.Sprintf("%s-%s", val, tt.Type())) {
			log.Debugf("the pytorchjob %s with labels %s found in List", pytorchjob.Name, val)
			jobInfo.Name = val
		} else {
			jobInfo.Name = pytorchjob.Name
		}

		jobInfo.Namespace = pytorchjob.Namespace
		jobInfos = append(jobInfos, jobInfo)

	}
	log.Debugf("jobInfos %v", jobInfos)

	for _, jobInfo := range jobInfos {
		job, err := tt.getTrainingJobFromCache(jobInfo.Name, jobInfo.Namespace)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Get PriorityClass
func (p *PyTorchJob) GetPriorityClass() string {
	pc := ""
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
func getPodsOfPyTorchJob(name string, tt *PyTorchJobTrainer, pytorchjob pytorchv1.PyTorchJob, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
	pods = []v1.Pod{}
	for _, item := range podList {
		if !tt.isPyTorchPod(name, namespace, item) {
			continue
		}
		if tt.isChiefPod(pytorchjob, item) && item.CreationTimestamp.After(chiefPod.CreationTimestamp.Time) {
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
