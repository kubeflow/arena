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

package training

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/arenacache"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeflow/arena/pkg/operators/volcano-operator/apis/batch/v1alpha1"
	"github.com/kubeflow/arena/pkg/operators/volcano-operator/client/clientset/versioned"
)

// volcano Job wrapper
type VolcanoJob struct {
	*BasicJobInfo
	volcanoJob   *v1alpha1.Job
	trainerType  types.TrainingJobType
	pods         []*v1.Pod
	chiefPod     *v1.Pod
	requestedGPU int64
	allocatedGPU int64
}

func (vj *VolcanoJob) Name() string {
	return vj.name
}

func (vj *VolcanoJob) Uid() string {
	return string(vj.volcanoJob.UID)
}

// return driver pod
func (vj *VolcanoJob) ChiefPod() *v1.Pod {
	return vj.chiefPod
}

// return trainerType: volcano job
func (vj *VolcanoJob) Trainer() types.TrainingJobType {
	return vj.trainerType
}

// return pods from cache
func (vj *VolcanoJob) AllPods() []*v1.Pod {
	return vj.pods
}

func (vj *VolcanoJob) GetTrainJob() interface{} {
	return vj.volcanoJob
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

func (vj *VolcanoJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {

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

	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(vj.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}

	for _, pod := range vj.pods {
		vj.requestedGPU += gpuInPod(*pod)
	}
	return vj.requestedGPU
}

// volcano job without gpu supported
func (vj *VolcanoJob) AllocatedGPU() int64 {

	if vj.allocatedGPU > 0 {
		return vj.allocatedGPU
	}
	for _, pod := range vj.pods {
		vj.allocatedGPU += gpuInActivePod(*pod)
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

// volcano job trainer
type VolcanoJobTrainer struct {
	client           *kubernetes.Clientset
	volcanoJobClient *versioned.Clientset
	trainerType      types.TrainingJobType
	enabled          bool
}

func NewVolcanoJobTrainer() Trainer {
	log.Debugf("Init Volcano job trainer")
	volcanoClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	enable := true
	// this step is used to check operator is installed or not
	_, err := volcanoClient.BatchV1alpha1().Jobs("default").Get("test-operator", metav1.GetOptions{})
	if err != nil && strings.Contains(err.Error(), errNotFoundOperator.Error()) {
		log.Debugf("not found volcano operator,volcano trainer is disabled")
		enable = false
	}
	return &VolcanoJobTrainer{
		volcanoJobClient: volcanoClient,
		client:           config.GetArenaConfiger().GetClientSet(),
		trainerType:      types.VolcanoTrainingJob,
		enabled:          enable,
	}
}

// IsEnabled is used to get the trainer is enable or not
func (st *VolcanoJobTrainer) IsEnabled() bool {
	return st.enabled
}

func (st *VolcanoJobTrainer) Type() types.TrainingJobType {
	return st.trainerType
}

func (st *VolcanoJobTrainer) IsSupported(name, ns string) bool {
	if !st.enabled {
		return false
	}
	isVolcanoJob := false
	_, err := st.GetTrainingJob(name, ns)
	if err != nil {
		return isVolcanoJob
	}
	return !isVolcanoJob
}

func (st *VolcanoJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	volcanoJob := &v1alpha1.Job{}
	var err error
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, volcanoJob)
		if err != nil {
			log.Errorf("%v", err)
			if strings.Contains(err.Error(), fmt.Sprintf(`Job.batch.volcano.sh "%v" not found`, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find volcanojob %v from cache,reason: %v", name, err)
		}

	} else {
		volcanoJob, err = st.volcanoJobClient.BatchV1alpha1().Jobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`jobs.batch.volcano.sh "%v" not found`, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find volcanojob from api server,reason: %v", err)
		}
	}
	// get the pods from the api server
	labels := map[string]string{
		"release": name,
		"app":     string(st.Type()),
	}
	podList, err := listJobPods(st.client, namespace, labels)
	if err != nil {
		return nil, err
	}
	pods := []*v1.Pod{}
	for _, pod := range podList.Items {
		pods = append(pods, pod.DeepCopy())
	}
	// filter pods and find chief pod
	filterPods, chiefPod := getPodsOfVolcanoJob(volcanoJob, st, pods)
	return &VolcanoJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(filterPods),
			name:      name,
		},
		volcanoJob:  volcanoJob,
		chiefPod:    chiefPod,
		pods:        filterPods,
		trainerType: st.Type(),
	}, nil
}

func (st *VolcanoJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	jobList, err := st.listJobs(namespace)
	if err != nil {
		return nil, err
	}
	trainingJobs := []TrainingJob{}
	for _, item := range jobList.Items {
		job := item.DeepCopy()
		labels := map[string]string{
			"release": job.Name,
			"app":     string(st.Type()),
		}
		podList, err := listJobPods(st.client, job.Namespace, labels)
		if err != nil {
			return nil, err
		}
		pods := []*v1.Pod{}
		for _, pod := range podList.Items {
			pods = append(pods, pod.DeepCopy())
		}
		// filter pods and find chief pod
		filterPods, chiefPod := getPodsOfVolcanoJob(job, st, pods)
		trainingJobs = append(trainingJobs, &VolcanoJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      job.Name,
			},
			volcanoJob:  job,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: st.Type(),
		})
	}
	return trainingJobs, nil
}

func (st *VolcanoJobTrainer) listJobs(namespace string) (*v1alpha1.JobList, error) {
	if config.GetArenaConfiger().IsDaemonMode() {
		list := &v1alpha1.JobList{}
		return list, arenacache.GetCacheClient().ListTrainingJobs(list, namespace)
	}
	return st.volcanoJobClient.BatchV1alpha1().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release"),
	})
}

func (st *VolcanoJobTrainer) isVolcanoJob(name, ns string, job *v1alpha1.Job) bool {
	if job.Labels["release"] != name {
		return false
	}
	log.Debugf("the volcano job %s with labels release=%s", job.Name, name)

	if job.Labels["app"] != string(st.trainerType) {
		return false
	}
	log.Debugf("the volcano job %s with labels app=%v is found.", job.Name, st.trainerType)
	if job.Namespace != ns {
		return false
	}
	return true
}

func (st *VolcanoJobTrainer) isVolcanoPod(name, ns string, pod *v1.Pod) bool {
	return utils.IsVolcanoPod(name, ns, pod)
}

func (st *VolcanoJobTrainer) isChiefPod(pod *v1.Pod) bool {
	if pod.Labels["volcano-role"] != "driver" {
		return false
	}
	log.Debugf("the volcano job %s with labels volcano-role=driver", pod.Name)
	return true
}

func getPodsOfVolcanoJob(job *v1alpha1.Job, st *VolcanoJobTrainer, podList []*v1.Pod) ([]*v1.Pod, *v1.Pod) {
	return getPodsOfTrainingJob(job.Name, job.Namespace, podList, st.isVolcanoPod, func(pod *v1.Pod) bool {
		return st.isChiefPod(pod)
	})
}
