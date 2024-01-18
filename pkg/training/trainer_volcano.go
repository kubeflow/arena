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
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

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
	volcanoClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	enable := false
	_, err := config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.VolcanoCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("VolcanoJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("VolcanoJobTrainer is disabled,reason: %v", err)
	}
	log.Debugf("Succeed to init Volcano job trainer")
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
	accesser := k8saccesser.GetK8sResourceAccesser()
	volcanoJob, err := accesser.GetVolcanoJob(st.volcanoJobClient, namespace, name)
	if err != nil {
		return nil, err
	}
	if err := CheckJobIsOwnedByTrainer(volcanoJob.Labels); err != nil {
		return nil, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("release=%v,app=%v", name, st.Type()), "", nil)
	if err != nil {
		return nil, err
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
	jobLabels := GetTrainingJobLabels(st.Type())
	jobs, err := k8saccesser.GetK8sResourceAccesser().ListVolcanoJobs(st.volcanoJobClient, namespace, jobLabels)
	if err != nil {
		return nil, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", st.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	trainingJobs := []TrainingJob{}
	for _, job := range jobs {
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
