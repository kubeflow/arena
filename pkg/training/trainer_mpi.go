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

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/operators/mpi-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"time"

	v1alpha1 "github.com/kubeflow/arena/pkg/operators/mpi-operator/apis/kubeflow/v1alpha1"
)

// MPI Job Information
type MPIJob struct {
	*BasicJobInfo
	mpijob       *v1alpha1.MPIJob
	chiefjob     *batchv1.Job
	pods         []*v1.Pod // all the pods including statefulset and job
	chiefPod     *v1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType // return trainer type: TENSORFLOW
}

func (mj *MPIJob) Name() string {
	return mj.name
}

func (mj *MPIJob) Uid() string {
	return string(mj.mpijob.UID)
}

// Get the chief Pod of the Job.
func (mj *MPIJob) ChiefPod() *v1.Pod {
	return mj.chiefPod
}

// Get the name of the Training Job
// func (mj *MPIJob) Name() string {
// 	return
// }

func (mj *MPIJob) Trainer() types.TrainingJobType {
	return mj.trainerType
}

// Get all the pods of the Training Job
func (mj *MPIJob) AllPods() []*v1.Pod {
	return mj.pods
}

func (mj *MPIJob) GetTrainJob() interface{} {
	return mj.mpijob
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (mj *MPIJob) GetStatus() (status string) {
	status = "UNKNOWN"
	if mj.mpijob.Name == "" {
		return status
	}

	if mj.isSucceeded() {
		status = "SUCCEEDED"
	} else if mj.isFailed() {
		status = "FAILED"
	} else if mj.isPending() {
		status = "PENDING"
	} else {
		status = "RUNNING"
	}

	return status
}

// Get the start time
func (mj *MPIJob) StartTime() *metav1.Time {
	return &mj.mpijob.CreationTimestamp
}

// Get the Job Age
func (mj *MPIJob) Age() time.Duration {
	job := mj.mpijob

	// use creation timestamp
	if job.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.CreationTimestamp.Time)
}

// Get the Job Training Duration
func (mj *MPIJob) Duration() time.Duration {
	mpijob := mj.mpijob

	if mpijob.CreationTimestamp.IsZero() {
		return 0
	}

	if len(mj.chiefjob.Name) != 0 && mj.chiefjob.Status.CompletionTime != nil {
		return mj.chiefjob.Status.CompletionTime.Time.Sub(mpijob.CreationTimestamp.Time)
	}

	if mj.isFailed() {
		cond := getPodLatestCondition(mj.chiefPod)
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Time.Sub(mpijob.CreationTimestamp.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", mj.chiefPod.Name)
		}
	}

	return metav1.Now().Sub(mpijob.CreationTimestamp.Time)
}

// Get Dashboard url of the job
func (mj *MPIJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
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

	if len(mj.chiefPod.Spec.Containers) == 0 {
		return urls, fmt.Errorf("mpi launcher is not ready!")
	}

	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		mj.chiefPod.Namespace,
		mj.chiefPod.Name,
		mj.chiefPod.Spec.Containers[0].Name,
		mj.chiefPod.Namespace)

	urls = append(urls, url)

	return urls, nil
}

// Requested GPU count of the Job
func (mj *MPIJob) RequestedGPU() int64 {
	if mj.requestedGPU > 0 {
		return mj.requestedGPU
	}
	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(mj.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}
	for _, pod := range mj.pods {
		mj.requestedGPU += gpuInPod(*pod)
	}
	return mj.requestedGPU
}

// Requested GPU count of the Job
func (mj *MPIJob) AllocatedGPU() int64 {
	if mj.allocatedGPU > 0 {
		return mj.allocatedGPU
	}
	for _, pod := range mj.pods {
		mj.allocatedGPU += gpuInActivePod(*pod)
	}
	return mj.allocatedGPU
}

// Get the hostIP of the chief Pod
func (mj *MPIJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if mj.GetStatus() == "RUNNING" {
		hostIP = mj.chiefPod.Status.HostIP
	}

	return hostIP
}

func (mj *MPIJob) Namespace() string {
	return mj.mpijob.Namespace
}

// MPI Job trainer
type MPIJobTrainer struct {
	client       *kubernetes.Clientset
	mpijobClient *versioned.Clientset
	trainerType  types.TrainingJobType
	// check if it's enabled
	enabled bool
}

// NewMPIJobTrainer
func NewMPIJobTrainer() Trainer {
	mpijobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	enable := false
	// this step is used to check operator is installed or not
	_, err := config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.MPICRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("MPIJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("MPIJobTrainer is disabled,reason: %v", err)
	}
	log.Debugf("Succeed to init MPIJobTrainer")
	return &MPIJobTrainer{
		mpijobClient: mpijobClient,
		client:       config.GetArenaConfiger().GetClientSet(),
		trainerType:  types.MPITrainingJob,
		enabled:      enable,
	}
}

// IsEnabled is used to get the trainer is enable or not
func (tt *MPIJobTrainer) IsEnabled() bool {
	return tt.enabled
}

// Get the type
func (tt *MPIJobTrainer) Type() types.TrainingJobType {
	return tt.trainerType
}

// check if it's TensorFlow job
func (tt *MPIJobTrainer) IsSupported(name, ns string) bool {
	if !tt.enabled {
		return false
	}
	isMPIJob := false
	_, err := tt.GetTrainingJob(name, ns)
	if err != nil {
		return isMPIJob
	}
	return !isMPIJob
}

func (tt *MPIJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	// 0. Get the batchJob of training Job
	mpijob, err := k8saccesser.GetK8sResourceAccesser().GetMPIJob(tt.mpijobClient, namespace, name)
	if err != nil {
		return nil, err
	}
	// 1. get the batch job of the mpijob
	batchJobs, err := k8saccesser.GetK8sResourceAccesser().ListBatchJobs(namespace, "mpi_job_name")
	if err != nil {
		return nil, err
	}
	chiefJob := tt.getChiefJob(batchJobs, name, namespace)

	// 2.get the statefulset
	statefulsets, err := k8saccesser.GetK8sResourceAccesser().ListStatefulSets(namespace, "mpi_job_name")
	if err != nil {
		return nil, err
	}
	// 2. Find the pod list, and determine the pod of the job
	allPods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("release=%v,app=%v", name, tt.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	// get chief pod
	pods, chiefPod := getPodsOfMPIJob(tt, mpijob, allPods)

	// 3. Find the other resources, like statefulset,job
	resources := tt.resources(statefulsets, batchJobs, name, namespace, pods)
	return &MPIJob{
		BasicJobInfo: &BasicJobInfo{
			resources: resources,
			name:      name,
		},
		mpijob:      mpijob,
		chiefPod:    chiefPod,
		chiefjob:    chiefJob,
		pods:        pods,
		trainerType: tt.Type(),
	}, nil

}

func (tt *MPIJobTrainer) getChiefJob(jobs []*batchv1.Job, name string, namespace string) *batchv1.Job {
	chiefJob := &batchv1.Job{}
	for _, job := range jobs {
		if tt.isChiefJob(job, name, namespace) {
			chiefJob = job
			break
		}
	}
	return chiefJob
}

func (tt *MPIJobTrainer) isChiefJob(job *batchv1.Job, name string, namespace string) bool {
	if job.Namespace != namespace {
		log.Debugf("The job %s in namespace %s not the same namespace as the mpijob %s in the namespace %s",
			job.Name,
			job.Namespace,
			name,
			namespace)
		return false
	}
	if job.Name == fmt.Sprintf("%s-launcher", name) || job.Name == fmt.Sprintf("%s-mpijob-launcher", name) {
		log.Debugf("The job %s is the chief job of %s", job.Name, name)
		return true
	}
	return false
}

func (tt *MPIJobTrainer) isChiefPod(item *v1.Pod) bool {
	if item.Labels["mpi_role_type"] != "launcher" {
		return false
	}
	log.Debugf("the mpijob %s with labels mpi_role_type=launcher", item.Name)
	return true
}

func (tt *MPIJobTrainer) isMPIJob(name, ns string, item v1alpha1.MPIJob) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the mpijob %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "mpijob") {
		log.Debugf("the mpijob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func (tt *MPIJobTrainer) isMPIPod(name, ns string, pod *v1.Pod) bool {
	return utils.IsMPIPod(name, ns, pod)
}

func (tt *MPIJobTrainer) resources(statefulsets []*appv1.StatefulSet, batchJobs []*batchv1.Job, name string, namespace string, pods []*v1.Pod) []Resource {
	resources := []Resource{}
	// 2. Find the pod list, and determine the pod of the job
	for _, sts := range statefulsets {
		if sts.Labels["mpi_job_name"] != name {
			continue
		}
		if sts.Namespace != namespace {
			continue
		}
		resources = append(resources, Resource{
			Name:         sts.Name,
			Uid:          string(sts.UID),
			ResourceType: ResourceTypeStatefulSet,
		})
	}

	for _, job := range batchJobs {
		if job.Namespace != namespace {
			continue
		}
		if job.Labels["mpi_job_name"] != name && job.Labels["mpi_job_name"] != fmt.Sprintf("%v-mpijob", name) {
			continue
		}
		resources = append(resources, Resource{
			Name:         job.Name,
			Uid:          string(job.UID),
			ResourceType: ResourceTypeJob,
		})
	}
	resources = append(resources, podResources(pods)...)
	return resources
}

/**
* List Training jobs
 */

func (tt *MPIJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	mpijobs, err := k8saccesser.GetK8sResourceAccesser().ListMPIJobs(tt.mpijobClient, namespace)
	if err != nil {
		return trainingJobs, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", tt.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	batchJobs, err := k8saccesser.GetK8sResourceAccesser().ListBatchJobs(namespace, "mpi_job_name")
	if err != nil {
		return nil, err
	}
	for _, mpijob := range mpijobs {
		chiefJob := tt.getChiefJob(batchJobs, mpijob.Name, mpijob.Namespace)
		// 3.find the pods, and determine the pod of the job
		filterPods, chiefPod := getPodsOfMPIJob(tt, mpijob, pods)
		trainingJobs = append(trainingJobs, &MPIJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      mpijob.Name,
			},
			mpijob:      mpijob,
			chiefPod:    chiefPod,
			pods:        filterPods,
			chiefjob:    chiefJob,
			trainerType: tt.Type(),
		})
	}
	return trainingJobs, nil
}

func (mj *MPIJob) isSucceeded() bool {
	// status.MPIJobLauncherStatusType
	return mj.mpijob.Status.LauncherStatus == v1alpha1.LauncherSucceeded
}

func (mj *MPIJob) isFailed() bool {
	return mj.mpijob.Status.LauncherStatus == v1alpha1.LauncherFailed
}

func (mj *MPIJob) isPending() bool {
	// return false
	if len(mj.chiefjob.Name) == 0 {
		log.Debugf("The MPIJob is pending due to chiefJob is not ready")
		return true
	}

	if len(mj.chiefPod.Name) == 0 || mj.chiefPod.Status.Phase == v1.PodPending {
		log.Debugf("The MPIJob is pending due to chiefPod is not ready")
		return true
	}

	return false
}

// Get PriorityClass
func (m *MPIJob) GetPriorityClass() string {
	// return ""
	return m.mpijob.Spec.Template.Spec.PriorityClassName
}

// filter out all pods and chief pod (master pod) of mpijob from pods in current system
func getPodsOfMPIJob(tt *MPIJobTrainer, mpijob *v1alpha1.MPIJob, podList []*v1.Pod) ([]*v1.Pod, *v1.Pod) {
	return getPodsOfTrainingJob(mpijob.Name, mpijob.Namespace, podList, tt.isMPIPod, func(pod *v1.Pod) bool {
		return tt.isChiefPod(pod)
	})
}
