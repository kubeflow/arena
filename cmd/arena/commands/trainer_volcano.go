// Copyright 2019 The Kubeflow Authors
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
	"time"

	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/volcano-operator/apis/batch/v1alpha1"
	"github.com/kubeflow/arena/pkg/volcano-operator/client/clientset/versioned"
)

// all volcano jobs cache
var (
	allVolcanoJobs []v1alpha1.Job
)

// volcano Job wrapper
type VolcanoJob struct {
	name         string
	volcanoJob   v1alpha1.Job
	trainerType  string
	pods         []v1.Pod
	chiefPod     v1.Pod
	requestedGPU int64
	allocatedGPU int64
}

func (vj *VolcanoJob) Name() string {
	return vj.name
}

// return driver pod
func (vj *VolcanoJob) ChiefPod() v1.Pod {
	return vj.chiefPod
}

// return trainerType: volcano job
func (vj *VolcanoJob) Trainer() string {
	return vj.trainerType
}

// return pods from cache
func (vj *VolcanoJob) AllPods() []v1.Pod {
	return vj.pods
}

func (vj *VolcanoJob) GetStatus() (status string) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("volcano job may not complete,because of ", r)
		}
		return
	}()

	status = "UNKNOWN"

	// name is empty when the pod has not been scheduled
	if vj.volcanoJob.Name == "" {
		return status
	}

	if vj.isSucceeded() {
		status = "SUCCEEDED"
	} else if vj.isFailed() {
		status = "FAILED"
	} else if vj.isPending() {
		status = "PENDING"
	} else if vj.isSubmitted() {
		status = "SUBMITTED"
	} else if vj.isRunning() {
		status = "RUNNING"
	} else {
		status = string(vj.volcanoJob.Status.State.Phase)
	}

	return status
}

func (vj *VolcanoJob) isSucceeded() bool {

	return vj.volcanoJob.Status.State.Phase == v1alpha1.Completed
}

func (vj *VolcanoJob) isFailed() bool {

	return vj.volcanoJob.Status.State.Phase == v1alpha1.Failed
}

func (vj *VolcanoJob) isPending() bool {

	return vj.volcanoJob.Status.State.Phase == v1alpha1.Pending
}

func (vj *VolcanoJob) isSubmitted() bool {

	return vj.volcanoJob.Status.State.Phase == v1alpha1.Inqueue
}

func (vj *VolcanoJob) isRunning() bool {

	return vj.volcanoJob.Status.State.Phase == v1alpha1.Running
}

func (vj *VolcanoJob) StartTime() *metav1.Time {

	return &vj.volcanoJob.CreationTimestamp
}

func (vj *VolcanoJob) Age() time.Duration {
	job := vj.volcanoJob

	if job.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.CreationTimestamp.Time)
}

// Get the Job Training Duration
func (vj *VolcanoJob) Duration() time.Duration {
	job := vj.volcanoJob

	if job.CreationTimestamp.IsZero() {
		return 0
	}
	// need to update once the back end changes are done
	// TODO
	return metav1.Now().Sub(job.Status.State.LastTransitionTime.Time)
}

func (vj *VolcanoJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {

	urls := []string{}
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

	if len(vj.chiefPod.Spec.Containers) == 0 {
		return urls, fmt.Errorf("volcano driver pod is not ready!")
	}

	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		vj.chiefPod.Namespace,
		vj.chiefPod.Name,
		vj.chiefPod.Spec.Containers[0].Name,
		vj.chiefPod.Namespace)

	urls = append(urls, url)

	return urls, nil
}

// volcano job without gpu supported
func (vj *VolcanoJob) RequestedGPU() int64 {

	if vj.requestedGPU > 0 {
		return vj.requestedGPU
	}
	for _, pod := range vj.pods {
		vj.requestedGPU += gpuInPod(pod)
	}
	return vj.requestedGPU
}

// volcano job without gpu supported
func (vj *VolcanoJob) AllocatedGPU() int64 {

	if vj.allocatedGPU > 0 {
		return vj.allocatedGPU
	}
	for _, pod := range vj.pods {
		vj.allocatedGPU += gpuInActivePod(pod)
	}
	return vj.allocatedGPU
}

// Get the hostIP of the driver Pod
func (vj *VolcanoJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if vj.GetStatus() == "RUNNING" {
		hostIP = vj.chiefPod.Status.HostIP
	}
	return hostIP
}

func (vj *VolcanoJob) Namespace() string {
	return vj.volcanoJob.Namespace
}

// Get PriorityClass
func (vj *VolcanoJob) GetPriorityClass() string {
	return ""
}

func NewVolcanoJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init Volcano job trainer")
	jobClient, err := initVolcanoJobClient()

	if err != nil {
		log.Debugf("unsupported volcano job due to %v", err)
		return &VolcanoJobTrainer{
			trainerType: defaultVolcanoJobTrainingType,
			enabled:     false,
		}
	}
	// allPods have been cached, we do the same to allVolcanoJobs
	if useCache {
		ns := namespace
		if allNamespaces {
			ns = metav1.NamespaceAll
		}

		jobList, err := jobClient.BatchV1alpha1().Jobs(ns).List(metav1.ListOptions{})
		if err != nil {
			log.Debugf("unsupported volcanoJob due to %v", err)
			return &VolcanoJobTrainer{
				trainerType: defaultVolcanoJobTrainingType,
				enabled:     false,
			}
		}

		for _, vJob := range jobList.Items {
			allVolcanoJobs = append(allVolcanoJobs, vJob)
		}
	}

	return &VolcanoJobTrainer{
		volcanoJobClient: jobClient,
		client:           client,
		trainerType:      defaultVolcanoJobTrainingType,
		enabled:          true,
	}
}

func NewVolcanoJobTrainerSubmit(client *kubernetes.Clientset) *VolcanoJobTrainer {
	log.Debugf("Init Volcano job trainer")
	jobClient, err := initVolcanoJobClient()

	if err != nil {
		log.Debugf("unsupported volcano job due to %v", err)
		return &VolcanoJobTrainer{
			trainerType: defaultVolcanoJobTrainingType,
			enabled:     false,
		}
	}
	// allPods have been cached, we do the same to allVolcanoJobs
	if useCache {
		ns := namespace
		if allNamespaces {
			ns = metav1.NamespaceAll
		}

		jobList, err := jobClient.BatchV1alpha1().Jobs(ns).List(metav1.ListOptions{})
		if err != nil {
			log.Debugf("unsupported volcanoJob due to %v", err)
			return &VolcanoJobTrainer{
				trainerType: defaultVolcanoJobTrainingType,
				enabled:     false,
			}
		}

		for _, vJob := range jobList.Items {
			allVolcanoJobs = append(allVolcanoJobs, vJob)
		}
	}

	return &VolcanoJobTrainer{
		volcanoJobClient: jobClient,
		client:           client,
		trainerType:      defaultVolcanoJobTrainingType,
		enabled:          true,
	}
}

// init volcano job client
func initVolcanoJobClient() (jobClientset *versioned.Clientset, err error) {
	if restConfig == nil {
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	jobClientset = versioned.NewForConfigOrDie(restConfig)
	return jobClientset, nil
}

// volcano job trainer
type VolcanoJobTrainer struct {
	client           *kubernetes.Clientset
	volcanoJobClient *versioned.Clientset
	trainerType      string
	enabled          bool
}

func (st *VolcanoJobTrainer) Type() string {
	return st.trainerType
}

func (st *VolcanoJobTrainer) IsSupported(name, ns string) bool {
	if !st.enabled {
		return false
	}

	isVolcano := false

	if useCache {
		for _, job := range allVolcanoJobs {
			if st.isVolcanoJob(name, ns, job) {
				isVolcano = true
				break
			}
		}
	} else {
		volcanoJobList, err := st.volcanoJobClient.BatchV1alpha1().Jobs(ns).List(metav1.ListOptions{})

		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(volcanoJobList.Items) > 0 {
			isVolcano = true
		}
	}

	return isVolcano
}

func (st *VolcanoJobTrainer) isVolcanoJob(name, ns string, job v1alpha1.Job) bool {
	if val, ok := job.Labels["release"]; ok && (val == name) {
		log.Debugf("the volcano job %s with labels %s", job.Name, val)
	} else {
		return false
	}

	if val, ok := job.Labels["app"]; ok && (val == "volcanojob") {
		log.Debugf("the volcano job %s with labels %s is found.", job.Name, val)
	} else {
		return false
	}

	if job.Namespace != ns {
		return false
	}
	return true
}

func (st *VolcanoJobTrainer) GetTrainingJob(name, namespace string) (job TrainingJob, err error) {
	if len(allVolcanoJobs) > 0 {
		job, err = st.getTrainingJobFromCache(name, namespace)
	} else {
		job, err = st.getTrainingJob(name, namespace, false)
	}

	return job, err
}

func (st *VolcanoJobTrainer) GetTrainingJobAtSubmit(name, namespace string) (job TrainingJob, err error) {
	if len(allVolcanoJobs) > 0 {
		job, err = st.getTrainingJobFromCache(name, namespace)
	} else {
		job, err = st.getTrainingJob(name, namespace, true)
	}

	return job, err
}

func (st *VolcanoJobTrainer) getTrainingJobFromCache(name, namespace string) (job TrainingJob, err error) {
	var (
		volcanoJob v1alpha1.Job
	)

	for _, item := range allVolcanoJobs {
		if st.isVolcanoJob(name, namespace, item) {
			volcanoJob = item
			break
		}
	}

	pods, chiefPod := getPodsOfVolcanoJob(name, st, allPods)

	return &VolcanoJob{
		chiefPod:    chiefPod,
		volcanoJob:  volcanoJob,
		pods:        pods,
		name:        name,
		trainerType: st.Type(),
	}, nil
}

func (st *VolcanoJobTrainer) getTrainingJob(name, namespace string, flag bool) (job TrainingJob, err error) {
	var (
		volcanoJob v1alpha1.Job
	)

	volcanoJobList, err := st.volcanoJobClient.BatchV1alpha1().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, err
	}
	if len(volcanoJobList.Items) == 0 {
		// for submit flow we should return nil other flows should return failure
		if flag {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		volcanoJob = volcanoJobList.Items[0]
	}

	podList, err := st.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, err
	}

	pods, chiefPod := getPodsOfVolcanoJob(name, st, podList.Items)

	return &VolcanoJob{
		volcanoJob:  volcanoJob,
		chiefPod:    chiefPod,
		pods:        pods,
		name:        name,
		trainerType: st.Type(),
	}, nil
}

func (st *VolcanoJobTrainer) ListTrainingJobs() (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	jobInfos := []types.TrainingJobInfo{}
	for _, volcanoJob := range allVolcanoJobs {
		jobInfo := types.TrainingJobInfo{}
		log.Debugf("find volcano job %s in %s", volcanoJob.Name, volcanoJob.Namespace)
		if val, ok := volcanoJob.Labels["release"]; ok && (volcanoJob.Name == fmt.Sprintf("%s-%s", val, st.Type())) {
			log.Debugf("the volcano job %s with labels %s found in List", volcanoJob.Name, val)
			jobInfo.Name = val
		} else {
			jobInfo.Name = volcanoJob.Name
		}

		jobInfo.Namespace = volcanoJob.Namespace
		jobInfos = append(jobInfos, jobInfo)
	}
	log.Debugf("jobInfos %v", jobInfos)

	for _, jobInfo := range jobInfos {
		job, err := st.getTrainingJobFromCache(jobInfo.Name, jobInfo.Namespace)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (st *VolcanoJobTrainer) isVolcanoPod(name, ns string, item v1.Pod) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the volcano job %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "volcanojob") {
		log.Debugf("the volcano job %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func (st *VolcanoJobTrainer) isChiefPod(item v1.Pod) bool {
	if val, ok := item.Labels["volcano-role"]; ok && (val == "driver") {
		log.Debugf("the volcano job %s with labels %s", item.Name, val)
	} else {
		return false
	}
	return true
}

func getPodsOfVolcanoJob(name string, st *VolcanoJobTrainer, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
	pods = []v1.Pod{}
	for _, item := range podList {
		if !st.isVolcanoPod(name, namespace, item) {
			continue
		}
		if st.isChiefPod(item) && item.CreationTimestamp.After(chiefPod.CreationTimestamp.Time) {
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
