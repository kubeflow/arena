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
	"sort"
	"strings"

	"github.com/kubeflow/arena/pkg/operators/tf-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"time"

	commonv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/common/v1"
	tfv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/tensorflow/v1"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/arenacache"
	//k8stypes "k8s.io/apimachinery/pkg/types"
)

const (
	// tf-operator added labels for pods and servers.
	tfReplicaTypeLabel     = "tf-replica-type"
	tfReplicaIndexLabel    = "tf-replica-index"
	labelGroupName         = "group-name"
	labelGroupNameV1alpha2 = "group_name"
	labelTFJobName         = "tf-job-name"
	labelTFJobRole         = "tf-job-role"
	TensorFlowCRD          = "tfjobs.kubeflow.org"
)

// TensorflowJob implements the TrainingJob
// TensorFlow Job Information
type TensorFlowJob struct {
	*BasicJobInfo
	tfjob        *tfv1.TFJob
	pods         []*v1.Pod // all the pods including statefulset and job
	chiefPod     *v1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType // return trainer type
}

// Name returns the TensorflowJob name
func (tj *TensorFlowJob) Name() string {
	return tj.name
}

// Uid returns the TensorflowJob uid
func (tj *TensorFlowJob) Uid() string {
	return string(tj.tfjob.UID)
}

// ChiefPod gets the chief Pod of the Job.
func (tj *TensorFlowJob) ChiefPod() *v1.Pod {
	return tj.chiefPod
}

// Trainer returns the trainer
func (tj *TensorFlowJob) Trainer() types.TrainingJobType {
	return tj.trainerType
}

// AllPods Get all the pods of the Training Job
func (tj *TensorFlowJob) AllPods() []*v1.Pod {
	return tj.pods
}

// GetTrainJob returns the training job
func (tj *TensorFlowJob) GetTrainJob() interface{} {
	return tj.tfjob
}

// GetStatus returns the status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (tj *TensorFlowJob) GetStatus() (status string) {
	status = "PENDING"
	if tj.tfjob.Name == "" {
		return status
	}
	t := checkStatus(tj.tfjob.Status)
	switch t {
	case commonv1.JobCreated, commonv1.JobRestarting:
		status = "PENDING"
	default:
		status = strings.ToUpper(string(t))
	}
	return status
}

// StartTime returns the start time
func (tj *TensorFlowJob) StartTime() *metav1.Time {
	return tj.tfjob.Status.StartTime
}

// Namespace returns the namespace of tfjob
func (tj *TensorFlowJob) Namespace() string {
	return tj.tfjob.Namespace
}

// Age returns the age of tfjob
func (tj *TensorFlowJob) Age() time.Duration {
	job := tj.tfjob

	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.Status.StartTime.Time)
}

//  Duration returns the duration of tfjob
func (tj *TensorFlowJob) Duration() time.Duration {
	job := tj.tfjob
	if job.Status.StartTime == nil || job.Status.StartTime.IsZero() {
		return 0
	}
	if !job.Status.CompletionTime.IsZero() {
		return job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time)
	}

	if tj.GetStatus() != "FAILED" {
		return metav1.Now().Sub(job.Status.StartTime.Time)
	}
	cond := getPodLatestCondition(tj.chiefPod)
	if !cond.LastTransitionTime.IsZero() {
		return cond.LastTransitionTime.Time.Sub(job.Status.StartTime.Time)
	}
	log.Debugf("the latest condition's time is zero of pod %s", tj.chiefPod.Name)
	return metav1.Now().Sub(job.Status.StartTime.Time)
}

// Get Dashboard url of the job
func (tj *TensorFlowJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
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
	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(tj.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}
	for _, pod := range tj.pods {
		tj.requestedGPU += gpuInPod(*pod)
	}
	return tj.requestedGPU
}

// Requested GPU count of the Job
func (tj *TensorFlowJob) AllocatedGPU() int64 {
	if tj.allocatedGPU > 0 {
		return tj.allocatedGPU
	}
	for _, pod := range tj.pods {
		tj.allocatedGPU += gpuInActivePod(*pod)
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
	// the k8s client
	client *kubernetes.Clientset
	// tfjob client
	tfjobClient *versioned.Clientset
	// trainer type
	trainerType types.TrainingJobType
	// stores the jobs
	// check if it's enabled
	enabled bool
}

func NewTensorFlowJobTrainer() Trainer {
	arenaConfiger := config.GetArenaConfiger()
	tfjobClient := versioned.NewForConfigOrDie(arenaConfiger.GetRestConfig())
	enable := false
	_, err := arenaConfiger.GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(TensorFlowCRD, metav1.GetOptions{})
	if err == nil {
		log.Debugf("TensorflowJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("TensorflowJobTrainer is disabled")
	}
	log.Debugf("Succeed to init TensorflowJobTrainer")
	return &TensorFlowJobTrainer{
		tfjobClient: tfjobClient,
		client:      arenaConfiger.GetClientSet(),
		trainerType: types.TFTrainingJob,
		enabled:     enable,
	}
}

// IsEnabled is used to get the trainer is enable or not
func (tt *TensorFlowJobTrainer) IsEnabled() bool {
	return tt.enabled
}

func (tt *TensorFlowJobTrainer) Type() types.TrainingJobType {
	return tt.trainerType
}

// check if it's TensorFlow job
func (tt *TensorFlowJobTrainer) IsSupported(name, namespace string) bool {
	if !tt.enabled {
		return false
	}
	isTensorFlow := false
	_, err := tt.GetTrainingJob(name, namespace)
	if err != nil {
		return isTensorFlow
	}
	return !isTensorFlow
}

func (tt *TensorFlowJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	// 1.get the tfjob
	tfjob := &tfv1.TFJob{}
	var err error
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, tfjob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`TFJob.kubeflow.org "%v" not found`, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find tfjob %v from cache,reason: %v", name, err)
		}

	} else {
		tfjob, err = tt.tfjobClient.KubeflowV1().TFJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, TensorFlowCRD, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find job %v from api server,reason: %v", name, err)
		}
	}
	// 2.get the pods of the job
	labels := map[string]string{
		"release": name,
		"app":     string(tt.Type()),
	}
	allPods, err := listJobPods(tt.client, namespace, labels)
	// Sort tfjob status conditions and make the newest condition at first
	tfjob.Status.Conditions = makeJobStatusSortedByTime(tfjob.Status.Conditions)
	if err != nil {
		return nil, err
	}
	pods, chiefPod := getPodsOfTFJob(tt, tfjob, allPods)

	return &TensorFlowJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(pods),
			name:      name,
		},
		tfjob:       tfjob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: tt.Type(),
	}, nil

}

func (tt *TensorFlowJobTrainer) isChiefPod(tfjob *tfv1.TFJob, item *v1.Pod) bool {

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

func (tt *TensorFlowJobTrainer) isTensorFlowJob(name, ns string, item *tfv1.TFJob) bool {

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

func (tt *TensorFlowJobTrainer) isTensorFlowPod(name, ns string, item *v1.Pod) bool {
	return utils.IsTensorFlowPod(name, ns, item)
}

/**
* List Training jobs
 */

func (tt *TensorFlowJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	tfjobList, err := tt.listJobs(namespace)
	if err != nil {
		return trainingJobs, err
	}
	labels := map[string]string{
		"app": string(tt.Type()),
	}
	pods, err := listJobPods(tt.client, namespace, labels)
	if err != nil {
		return nil, err
	}
	for _, job := range tfjobList.Items {
		tfjob := job.DeepCopy()
		tfjob.Status.Conditions = makeJobStatusSortedByTime(tfjob.Status.Conditions)
		filterPods, chiefPod := getPodsOfTFJob(tt, tfjob, pods)
		trainingJobs = append(trainingJobs, &TensorFlowJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      tfjob.Name,
			},
			tfjob:       tfjob,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: tt.Type(),
		})
	}
	return trainingJobs, nil
}

func (tt *TensorFlowJobTrainer) listJobs(namespace string) (*tfv1.TFJobList, error) {
	if config.GetArenaConfiger().IsDaemonMode() {
		list := &tfv1.TFJobList{}
		return list, arenacache.GetCacheClient().ListTrainingJobs(list, namespace)
	}
	return tt.tfjobClient.KubeflowV1().TFJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release"),
	})
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

func getPodsOfTFJob(tt *TensorFlowJobTrainer, tfjob *tfv1.TFJob, podList []*v1.Pod) ([]*v1.Pod, *v1.Pod) {
	return getPodsOfTrainingJob(tfjob.Name, tfjob.Namespace, podList, tt.isTensorFlowPod, func(pod *v1.Pod) bool {
		return tt.isChiefPod(tfjob, pod)
	})
}
