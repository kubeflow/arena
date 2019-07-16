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
	"sort"
	"strings"

	"github.com/kubeflow/arena/pkg/tf-operator/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"time"

	commonv1 "github.com/kubeflow/arena/pkg/tf-operator/apis/common/v1"
	tfv1 "github.com/kubeflow/arena/pkg/tf-operator/apis/tensorflow/v1"
)

const (
	// tf-operator added labels for pods and servers.
	tfReplicaTypeLabel     = "tf-replica-type"
	tfReplicaIndexLabel    = "tf-replica-index"
	labelGroupName         = "group-name"
	labelGroupNameV1alpha2 = "group_name"
	labelTFJobName         = "tf-job-name"
	labelTFJobRole         = "tf-job-role"
)

var (
	allTfjobs []tfv1.TFJob
)

func initTFJobClient() (tfjobClientset *versioned.Clientset, err error) {
	if restConfig == nil {
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	// create the tfjobClientset
	tfjobClientset = versioned.NewForConfigOrDie(restConfig)
	return tfjobClientset, nil
}

// TensorFlow Job Information
type TensorFlowJob struct {
	name         string
	tfjob        tfv1.TFJob
	pods         []v1.Pod // all the pods including statefulset and job
	chiefPod     v1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  string // return trainer type: TENSORFLOW
}

func (tj *TensorFlowJob) Name() string {
	return tj.name
}

// Get the chief Pod of the Job.
func (tj *TensorFlowJob) ChiefPod() v1.Pod {
	return tj.chiefPod
}

// Get the name of the Training Job
// func (tj *TensorFlowJob) Name() string {
// 	return
// }

func (tj *TensorFlowJob) Trainer() string {
	return tj.trainerType
}

// Get all the pods of the Training Job
func (tj *TensorFlowJob) AllPods() []v1.Pod {
	return tj.pods
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (tj *TensorFlowJob) GetStatus() (status string) {
	status = "PENDING"
	if tj.tfjob.Name == "" {
		return status
	}

	t := checkStatus(tj.tfjob.Status)
	if t == commonv1.JobCreated || t == commonv1.JobRestarting {
		status = "PENDING"
	} else {
		status = strings.ToUpper(string(t))
	}

	return status
}

func (tj *TensorFlowJob) StartTime() *metav1.Time {
	return tj.tfjob.Status.StartTime
}

func (tj *TensorFlowJob) Namespace() string {
	return tj.tfjob.Namespace
}

// Get the Job Age
func (tj *TensorFlowJob) Age() time.Duration {
	job := tj.tfjob

	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.Status.StartTime.Time)
}

// Get the Job Training Duration
func (tj *TensorFlowJob) Duration() time.Duration {
	job := tj.tfjob

	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}

	if !job.Status.CompletionTime.IsZero() {
		return job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time)
	}

	if tj.GetStatus() == "FAILED" {
		cond := getPodLatestCondition(tj.chiefPod)
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Time.Sub(job.Status.StartTime.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", tj.chiefPod.Name)
		}
	}

	return metav1.Now().Sub(job.Status.StartTime.Time)
}

// Get Dashboard url of the job
func (tj *TensorFlowJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	urls := []string{}
	// dashboardURL, err := dashboard(client, "kubeflow", "tf-job-dashboard")
	dashboardURL, err := dashboard(client, namespace, "tf-job-dashboard")

	if err != nil {
		log.Debugf("Get dashboard failed due to %v", err)
		// retry for the existing customers, will be deprecated in the future
		dashboardURL, err = dashboard(client, arenaNamespace, "tf-job-dashboard")
		if err != nil {
			log.Debugf("Get dashboard failed due to %v", err)
		}
	}

	if err != nil {
		log.Debugf("Get dashboard failed due to %v", err)
		// retry for the existing customers, will be deprecated in the future
		dashboardURL, err = dashboard(client, "kubeflow", "tf-job-dashboard")
		if err != nil {
			log.Debugf("Get dashboard failed due to %v", err)
		}
	}

	if dashboardURL == "" {
		return urls, fmt.Errorf("No LOGVIEWER Installed.")
	}

	tfjob := tj.tfjob
	url := fmt.Sprintf("%s/tfjobs/ui/#/%s/%s\n",
		dashboardURL,
		tfjob.Namespace,
		tfjob.Name)

	urls = append(urls, url)

	return urls, nil
}

// Requested GPU count of the Job
func (tj *TensorFlowJob) RequestedGPU() int64 {
	if tj.requestedGPU > 0 {
		return tj.requestedGPU
	}
	for _, pod := range tj.pods {
		tj.requestedGPU += gpuInPod(pod)
	}
	return tj.requestedGPU
}

// Requested GPU count of the Job
func (tj *TensorFlowJob) AllocatedGPU() int64 {
	if tj.allocatedGPU > 0 {
		return tj.allocatedGPU
	}
	for _, pod := range tj.pods {
		tj.allocatedGPU += gpuInActivePod(pod)
	}
	return tj.allocatedGPU
}

// Get the hostIP of the chief Pod
func (tj *TensorFlowJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if tj.GetStatus() == "RUNNING" {
		hostIP = tj.chiefPod.Status.HostIP
	}

	return hostIP
}

// Get PriorityClass
func (t *TensorFlowJob) GetPriorityClass() string {
	pc := ""
	specs := t.tfjob.Spec.TFReplicaSpecs
	log.Debugf("specs %v", specs)
	for _, spec := range specs {
		if spec.Template.Spec.PriorityClassName != "" {
			pc = spec.Template.Spec.PriorityClassName
			break
		}
	}

	return pc
}

// TensorFlow Job trainer
type TensorFlowJobTrainer struct {
	client      *kubernetes.Clientset
	tfjobClient *versioned.Clientset
	trainerType string
	// check if it's enabled
	enabled bool
}

func NewTensorFlowJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init TensorFlow job trainer")
	tfjobClient, err := initTFJobClient()
	if err != nil {
		log.Debugf("unsupported tfjobs due to %v", err)
		return &TensorFlowJobTrainer{
			trainerType: "tfjob",
			enabled:     false,
		}
	}
	// allPods have been cached, we do the same to allTfjobs
	if useCache {
		ns := namespace
		if allNamespaces {
			ns = metav1.NamespaceAll
		}

		tfjobList, err := tfjobClient.KubeflowV1().TFJobs(ns).List(metav1.ListOptions{})
		if err != nil {
			log.Debugf("unsupported tfjobs due to %v", err)
			return &TensorFlowJobTrainer{
				trainerType: "tfjob",
				enabled:     false,
			}
		}

		for _, tfjob := range tfjobList.Items {
			allTfjobs = append(allTfjobs, tfjob)
		}
	}

	return &TensorFlowJobTrainer{
		tfjobClient: tfjobClient,
		client:      client,
		trainerType: "tfjob",
		enabled:     true,
	}
}

func (tt *TensorFlowJobTrainer) Type() string {
	return tt.trainerType
}

// check if it's TensorFlow job
func (tt *TensorFlowJobTrainer) IsSupported(name, ns string) bool {
	if !tt.enabled {
		return false
	}

	isTensorFlow := false

	if useCache {
		for _, job := range allTfjobs {
			if tt.isTensorFlowJob(name, ns, job) {
				isTensorFlow = true
				log.Debugf("the job %s for %s in namespace %s is found.", job.Name, name, ns)
				break
			}
		}
	} else {
		tfjobList, err := tt.tfjobClient.KubeflowV1().TFJobs(ns).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("release=%s", name),
		})

		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(tfjobList.Items) > 0 {
			isTensorFlow = true
		}
	}

	return isTensorFlow
}

func (tt *TensorFlowJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	if useCache {
		tj, err = tt.getTrainingJobFromCache(name, namespace)
	} else {
		tj, err = tt.getTrainingJob(name, namespace)
	}

	return tj, err
}

func (tt *TensorFlowJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		tfjob tfv1.TFJob
	)

	// 1. Get the batchJob of training Job
	pods := []v1.Pod{}

	tfjobList, err := tt.tfjobClient.KubeflowV1().TFJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})
	if err != nil {
		return nil, err
	}

	if len(tfjobList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		tfjob = tfjobList.Items[0]
	}

	// Sort tfjob status conditions and make the newest condition at first
	tfjob.Status.Conditions = makeJobStatusSortedByTime(tfjob.Status.Conditions)

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
	pods, chiefPod := getPodsOfTFJob(name, tt, tfjob, podList.Items)

	return &TensorFlowJob{
		tfjob:       tfjob,
		chiefPod:    chiefPod,
		pods:        pods,
		name:        name,
		trainerType: tt.Type(),
	}, nil

}

// Get the training job from Cache
func (tt *TensorFlowJobTrainer) getTrainingJobFromCache(name, ns string) (TrainingJob, error) {

	var (
		tfjob tfv1.TFJob
	)

	// 1. Find the batch job
	for _, item := range allTfjobs {
		if tt.isTensorFlowJob(name, ns, item) {
			tfjob = item
			break
		}
	}
	tfjob.Status.Conditions = makeJobStatusSortedByTime(tfjob.Status.Conditions)
	// 2. Find the pods, and determine the pod of the job
	pods, chiefPod := getPodsOfTFJob(name, tt, tfjob, allPods)

	return &TensorFlowJob{
		tfjob:       tfjob,
		chiefPod:    chiefPod,
		pods:        pods,
		name:        name,
		trainerType: tt.Type(),
	}, nil
}

func (tt *TensorFlowJobTrainer) isChiefPod(tfjob tfv1.TFJob, item v1.Pod) bool {

	// find chief pod in chief mode
	if _, ok := tfjob.Spec.TFReplicaSpecs[tfv1.TFReplicaTypeChief]; ok {
		log.Debugf("The distributed tensorflow is in chief mode")
		if val, ok := item.Labels[tfReplicaTypeLabel]; ok && (val == "chief") {
			log.Debugf("the tfjob %s with labels %s is the chief pod", item.Name, val)
			return true
		} else {
			return false
		}
	}

	if val, ok := item.Labels[tfReplicaTypeLabel]; ok && (val == "worker") {
		log.Debugf("the tfjob %s with labels %s is the chief pod", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels[tfReplicaIndexLabel]; ok && (val == "0") {
		log.Debugf("the chief pod of tfjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	return true
}

func (tt *TensorFlowJobTrainer) isTensorFlowJob(name, ns string, item tfv1.TFJob) bool {

	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the tfjob %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "tfjob") {
		log.Debugf("the tfjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func (tt *TensorFlowJobTrainer) isTensorFlowPod(name, ns string, item v1.Pod) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the tfjob %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "tfjob") {
		log.Debugf("the tfjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels[labelGroupName]; ok && (val == "kubeflow.org") {
		log.Debugf("the tfjob %s with labels %s is found.", item.Name, val)
	} else if val, ok := item.Labels[labelGroupNameV1alpha2]; ok && (val == "kubeflow.org") {
		log.Debugf("the tfjob v1alpha2 %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

/**
* List Training jobs
 */
func (tt *TensorFlowJobTrainer) ListTrainingJobs() (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	jobInfos := []types.TrainingJobInfo{}
	for _, tfjob := range allTfjobs {
		jobInfo := types.TrainingJobInfo{}
		log.Debugf("find tfjob %s in %s", tfjob.Name, tfjob.Namespace)
		if val, ok := tfjob.Labels["release"]; ok && (tfjob.Name == fmt.Sprintf("%s-%s", val, tt.Type())) {
			log.Debugf("the tfjob %s with labels %s found in List", tfjob.Name, val)
			jobInfo.Name = val
		} else {
			jobInfo.Name = tfjob.Name
		}

		jobInfo.Namespace = tfjob.Namespace
		jobInfos = append(jobInfos, jobInfo)
		// jobInfos = append(jobInfos, types.TrainingJobInfo{Name: tfjob.})
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

type orderedTrainingJobConditionsByTime []commonv1.JobCondition

func (t orderedTrainingJobConditionsByTime) Len() int {
	return len(t)
}

func (t orderedTrainingJobConditionsByTime) Less(i, j int) bool {
	if t[i].LastUpdateTime.IsZero() {
		return true
	} else if t[j].LastUpdateTime.IsZero() {
		return false
	}

	return t[i].LastUpdateTime.After(t[j].LastUpdateTime.Time)
}

func (t orderedTrainingJobConditionsByTime) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func makeJobStatusSortedByTime(conditions []commonv1.JobCondition) []commonv1.JobCondition {
	newConditions := make(orderedTrainingJobConditionsByTime, 0, len(conditions))
	for _, c := range conditions {
		newConditions = append(newConditions, c)
	}
	sort.Sort(newConditions)
	return []commonv1.JobCondition(newConditions)
}

func hasCondition(status commonv1.JobStatus, condType commonv1.JobConditionType) bool {
	for _, condition := range status.Conditions {
		if condition.Type == condType && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func checkStatus(status commonv1.JobStatus) commonv1.JobConditionType {
	t := commonv1.JobConditionType("Pending")
	for _, condition := range status.Conditions {
		if condition.Status == v1.ConditionTrue {
			t = condition.Type
			break
		}
	}
	return t
}

func getPodsOfTFJob(name string, tt *TensorFlowJobTrainer, tfjob tfv1.TFJob, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
	pods = []v1.Pod{}
	for _, item := range podList {
		if !tt.isTensorFlowPod(name, namespace, item) {
			continue
		}
		if tt.isChiefPod(tfjob, item) && item.CreationTimestamp.After(chiefPod.CreationTimestamp.Time) {
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
