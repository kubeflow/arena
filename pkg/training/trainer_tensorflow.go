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
	"sort"
	"strings"

	"github.com/kubeflow/arena/pkg/operators/tf-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"time"

	commonv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/common/v1"
	tfv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/tensorflow/v1"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/arenacache"
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
	errTFJobNotFound = errors.New("tfjob not found")
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
	log.Debugf("Init TensorFlow job trainer")
	arenaConfiger := config.GetArenaConfiger()
	tfjobClient := versioned.NewForConfigOrDie(arenaConfiger.GetRestConfig())
	return &TensorFlowJobTrainer{
		tfjobClient: tfjobClient,
		client:      arenaConfiger.GetClientSet(),
		trainerType: types.TFTrainingJob,
		enabled:     true,
	}
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
	if config.GetArenaConfiger().IsDaemonMode() {
		_, err := tt.getTrainingJobFromCache(name, namespace)
		if err != nil {
			return isTensorFlow
		}
		return !isTensorFlow
	}
	_, err := tt.getTrainingJob(name, namespace)
	if err != nil {
		return isTensorFlow
	}
	return !isTensorFlow
}

func (tt *TensorFlowJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	if config.GetArenaConfiger().IsDaemonMode() {
		return tt.getTrainingJobFromCache(name, namespace)
	}
	// get job from api server
	return tt.getTrainingJob(name, namespace)
}

func (tt *TensorFlowJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	// 1.get the tfjob
	tfjob, err := tt.tfjobClient.KubeflowV1().TFJobs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(`tfjobs.kubeflow.org "%v" not found`, name)) {
			return nil, errTFJobNotFound
		}
		return nil, fmt.Errorf("failed to find job %v,reason: %v", name, err)
	}
	// 2.get the pods of the job
	allPods := []*v1.Pod{}
	podList, err := tt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s,app=%v", name, tt.trainerType),
	})
	for _, pod := range podList.Items {
		allPods = append(allPods, pod.DeepCopy())
	}
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

// Get the training job from Cache
func (tt *TensorFlowJobTrainer) getTrainingJobFromCache(name, namespace string) (TrainingJob, error) {
	tfjob, pods := arenacache.GetArenaCache().GetTFJob(namespace, name)
	if tfjob == nil {
		return nil, errTFJobNotFound
	}
	tfjob.Status.Conditions = makeJobStatusSortedByTime(tfjob.Status.Conditions)
	filterPods, chiefPod := getPodsOfTFJob(tt, tfjob, pods)
	return &TensorFlowJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(filterPods),
			name:      name,
		},
		tfjob:       tfjob,
		chiefPod:    chiefPod,
		pods:        filterPods,
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
	// if arena is configured as daemon,getting all tfjobs from cache is corrent
	if config.GetArenaConfiger().IsDaemonMode() {
		return tt.listFromCache(namespace, allNamespace)
	}
	return tt.listFromAPIServer(namespace, allNamespace)
}

func (tt *TensorFlowJobTrainer) listFromAPIServer(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	// list all jobs from k8s apiserver
	tfjobList, err := tt.tfjobClient.KubeflowV1().TFJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release"),
	})
	if err != nil {
		return trainingJobs, err
	}
	for _, job := range tfjobList.Items {
		tfjob := job.DeepCopy()
		podList, err := tt.client.CoreV1().Pods(tfjob.Namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: fmt.Sprintf("release=%s,app=%v", tfjob.Name, tt.trainerType),
		})
		if err != nil {
			log.Errorf("failed to get pods of job %v,reason: %v", tfjob.Name, err)
			continue
		}
		pods := []*v1.Pod{}
		for _, pod := range podList.Items {
			pods = append(pods, pod.DeepCopy())
		}
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

func (tt *TensorFlowJobTrainer) listFromCache(namespace string, allNamespace bool) ([]TrainingJob, error) {
	filter := func(job *tfv1.TFJob) bool { return job.Namespace == namespace }
	trainingJobs := []TrainingJob{}
	if allNamespace {
		filter = func(job *tfv1.TFJob) bool { return true }
	}
	jobs, pods := arenacache.GetArenaCache().FilterTFJobs(filter)
	for jobKey, tfjob := range jobs {
		tfjob.Status.Conditions = makeJobStatusSortedByTime(tfjob.Status.Conditions)
		filterPods, chiefPod := getPodsOfTFJob(tt, tfjob, pods[jobKey])
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
