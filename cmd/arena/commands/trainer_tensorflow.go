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

	"github.com/kubeflow/arena/util"
	"github.com/kubeflow/tf-operator/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	tfv1alpha2 "github.com/kubeflow/tf-operator/pkg/apis/tensorflow/v1alpha2"
)

var (
	allTfjobs []tfv1alpha2.TFJob
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
	tfjob        tfv1alpha2.TFJob
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
	status = "UNKNOWN"
	if tj.tfjob.Name == "" {
		return status
	}

	t := checkStatus(tj.tfjob.Status)
	if t == tfv1alpha2.TFJobCreated || t == tfv1alpha2.TFJobRestarting {
		status = "PENDING"
	} else {
		status = strings.ToUpper(string(t))
	}

	return status
}

func (tj *TensorFlowJob) StartTime() *metav1.Time {
	return tj.tfjob.Status.StartTime
}

// Get the Job Age
func (tj *TensorFlowJob) Age() string {
	job := tj.tfjob

	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return "0s"
	}
	d := metav1.Now().Sub(job.Status.StartTime.Time)

	return util.ShortHumanDuration(d)
}

// Get Dashboard url of the job
func (tj *TensorFlowJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	urls := []string{}
	// dashboardURL, err := dashboard(client, "kubeflow", "tf-job-dashboard")
	dashboardURL, err := dashboard(client, arenaNamespace, "tf-job-dashboard")

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
	if len(allPods) > 0 {
		tfjobList, err := tfjobClient.KubeflowV1alpha2().TFJobs(metav1.NamespaceAll).List(metav1.ListOptions{})
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

	if len(allTfjobs) > 0 {
		for _, job := range allTfjobs {
			if tt.isTensorFlowJob(name, ns, job) {
				isTensorFlow = true
				log.Debugf("the job %s for %s in namespace %s is found.", job.Name, name, ns)
				break
			}
		}
	} else {
		tfjobList, err := tt.tfjobClient.KubeflowV1alpha2().TFJobs(namespace).List(metav1.ListOptions{
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
	if len(allTfjobs) > 0 {
		tj, err = tt.getTrainingJobFromCache(name, namespace)
	} else {
		tj, err = tt.getTrainingJob(name, namespace)
	}

	return tj, err
}

func (tt *TensorFlowJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		tfjob    tfv1alpha2.TFJob
	)

	// 1. Get the batchJob of training Job
	pods := []v1.Pod{}

	tfjobList, err := tt.tfjobClient.KubeflowV1alpha2().TFJobs(namespace).List(metav1.ListOptions{
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
	pods, chiefPod := getPodsOfTFJob(tt, tfjob, podList.Items)

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
		tfjob    tfv1alpha2.TFJob
	)

	// 1. Find the batch job
	for _, item := range allTfjobs {
		if tt.isTensorFlowJob(name, ns, item) {
			tfjob = item
			break
		}
	}


	// 2. Find the pods, and determine the pod of the job
	pods, chiefPod := getPodsOfTFJob(tt, tfjob, allPods)

	return &TensorFlowJob{
		tfjob:       tfjob,
		chiefPod:    chiefPod,
		pods:        pods,
		name:        name,
		trainerType: tt.Type(),
	}, nil
}

func (tt *TensorFlowJobTrainer) isChiefPod(tfjob tfv1alpha2.TFJob, item v1.Pod) bool {

	// find chief pod in chief mode
	if _, ok := tfjob.Spec.TFReplicaSpecs[tfv1alpha2.TFReplicaTypeChief]; ok {
		log.Debugf("The distributed tensorflow is in chief mode")
		if val, ok := item.Labels["tf-replica-type"]; ok && (val == "chief") {
			log.Debugf("the tfjob %s with labels %s is the chief pod", item.Name, val)
			return true
		} else {
			return false
		}
	}

	if val, ok := item.Labels["tf-replica-type"]; ok && (val == "worker") {
		log.Debugf("the tfjob %s with labels %s is the chief pod", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["tf-replica-index"]; ok && (val == "0") {
		log.Debugf("the chief pod of tfjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	return true
}

func (tt *TensorFlowJobTrainer) isTensorFlowJob(name, ns string, item tfv1alpha2.TFJob) bool {

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

	if val, ok := item.Labels["group_name"]; ok && (val == "kubeflow.org") {
		log.Debugf("the tfjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

type orderedTrainingJobConditionsByTime []tfv1alpha2.TFJobCondition

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

func makeJobStatusSortedByTime(conditions []tfv1alpha2.TFJobCondition) []tfv1alpha2.TFJobCondition {
	newConditions := make(orderedTrainingJobConditionsByTime, 0, len(conditions))
	for _, c := range conditions {
		newConditions = append(newConditions, c)
	}
	sort.Sort(newConditions)
	return []tfv1alpha2.TFJobCondition(newConditions)
}

func hasCondition(status tfv1alpha2.TFJobStatus, condType tfv1alpha2.TFJobConditionType) bool {
	for _, condition := range status.Conditions {
		if condition.Type == condType && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func checkStatus(status tfv1alpha2.TFJobStatus) tfv1alpha2.TFJobConditionType {
	t := tfv1alpha2.TFJobConditionType("Unknown")
	for _, condition := range status.Conditions {
		if condition.Status == v1.ConditionTrue {
			t = condition.Type
			break
		}
	}
	return t
}

func getPodsOfTFJob(tt *TensorFlowJobTrainer, tfjob tfv1alpha2.TFJob, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
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
