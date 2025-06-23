// Copyright 2024 The Kubeflow Authors
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
	"github.com/ray-project/kuberay/ray-operator/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
)

const (
	// rayjob-operator added labels for pods and servers.
	labelRayJobName = "job-name"
)

// RayJob Information
type RayJob struct {
	*BasicJobInfo
	rayJob       *rayv1.RayJob
	pods         []*corev1.Pod // all the pods including statefulset and job
	chiefPod     *corev1.Pod   // the head pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType // return trainer type: rayjob
}

// RayJob implements the TrainingJob interface.
var _ TrainingJob = &RayJob{}

// Name returns the RayJob name
func (rj *RayJob) Name() string {
	return rj.name
}

func (rj *RayJob) Uid() string {
	return string(rj.rayJob.UID)
}

// Get the head Pod of the Job.
func (rj *RayJob) ChiefPod() *corev1.Pod {
	return rj.chiefPod
}

func (rj *RayJob) Trainer() types.TrainingJobType {
	return rj.trainerType
}

// Get all the pods of the Training Job
func (rj *RayJob) AllPods() []*corev1.Pod {
	return rj.pods
}

func (rj *RayJob) GetTrainJob() interface{} {
	return rj.rayJob
}

func (rj *RayJob) GetLabels() map[string]string {
	return rj.rayJob.Labels
}

// Get the Status of the rayJob: PENDING, RUNNING, STOPPED, SUCCEEDED, FAILED
func (rj *RayJob) GetStatus() (status string) {
	status = string(rj.rayJob.Status.JobStatus)
	if status == string(rayv1.JobStatusNew) {
		status = "PENDING"
	}
	return
}

// Get the start time
func (rj *RayJob) StartTime() *metav1.Time {
	return &rj.rayJob.CreationTimestamp
}

// Get the Job Age
func (rj *RayJob) Age() time.Duration {
	job := rj.rayJob

	// use creation timestamp
	if job.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.CreationTimestamp.Time)
}

// Get the Job Training Duration.
func (rj *RayJob) Duration() time.Duration {
	rayjob := rj.rayJob

	if rayjob.Status.StartTime == nil ||
		rayjob.Status.StartTime.IsZero() {
		return 0
	}

	if !rayjob.Status.EndTime.IsZero() {
		return rayjob.Status.EndTime.Sub(rayjob.Status.StartTime.Time)
	}

	if rj.GetStatus() == "FAILED" {
		cond := getPodLatestCondition(rj.chiefPod)
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Sub(rayjob.CreationTimestamp.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", rj.chiefPod.Name)
		}
	}

	return metav1.Now().Sub(rayjob.Status.StartTime.Time)
}

// Get Dashboard url of the job
func (rj *RayJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
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
		return urls, fmt.Errorf("no LOGVIEWER Installed")
	}

	if len(rj.chiefPod.Spec.Containers) == 0 {
		return urls, fmt.Errorf("head pod is not ready")
	}

	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		rj.chiefPod.Namespace,
		rj.chiefPod.Name,
		rj.chiefPod.Spec.Containers[0].Name,
		rj.chiefPod.Namespace)

	urls = append(urls, url)

	return urls, nil
}

// Requested GPU count of the Job
func (rj *RayJob) RequestedGPU() int64 {
	if rj.requestedGPU > 0 {
		return rj.requestedGPU
	}
	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(rj.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}
	for _, pod := range rj.pods {
		rj.requestedGPU += gpuInPod(*pod)
	}
	return rj.requestedGPU
}

// Requested GPU count of the Job
func (rj *RayJob) AllocatedGPU() int64 {
	if rj.allocatedGPU > 0 {
		return rj.allocatedGPU
	}
	for _, pod := range rj.pods {
		rj.allocatedGPU += gpuInActivePod(*pod)
	}
	return rj.allocatedGPU
}

// Get the hostIP of the master Pod
func (rj *RayJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if rj.GetStatus() == "RUNNING" {
		hostIP = rj.chiefPod.Status.HostIP
	}

	return hostIP
}

func (rj *RayJob) Namespace() string {
	return rj.rayJob.Namespace
}

// Get PriorityClass. return the PriorityClassName of HeadPod
func (rj *RayJob) GetPriorityClass() string {
	pc := ""
	log.Debugf("rayjob: %v", rj.rayJob)
	pc = rj.rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec.PriorityClassName
	log.Debugf("PriorityClassName: %v", pc)

	return pc
}

// RayJob Job trainer
type RayJobTrainer struct {
	client       *kubernetes.Clientset
	RayJobClient *versioned.Clientset
	trainerType  types.TrainingJobType
	// check if it's enabled
	enabled bool
}

// RayJobTrainer implements the Trainer interface.
var _ Trainer = &RayJobTrainer{}

// NewRayJobTrainer
func NewRayJobTrainer() Trainer {
	// get rayjob operator client call rayjob operator api
	enable := false
	RayJobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	_, err := config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.RayJobCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("RayJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("RayJobTrainer is disabled,reason: %v", err)
	}
	log.Debugf("Succeed to init RayJobTrainer")
	return &RayJobTrainer{
		RayJobClient: RayJobClient,
		client:       config.GetArenaConfiger().GetClientSet(),
		trainerType:  types.RayJob,
		enabled:      enable,
	}
}

// IsEnabled is used to get the trainer is enable or not
func (rjt *RayJobTrainer) IsEnabled() bool {
	return rjt.enabled
}

// Get the type
func (rjt *RayJobTrainer) Type() types.TrainingJobType {
	return rjt.trainerType
}

// check if it's ray job
func (rjt *RayJobTrainer) IsSupported(name, ns string) bool {
	if !rjt.enabled {
		return false
	}
	isRayJob := false
	_, err := rjt.GetTrainingJob(name, ns)
	if err != nil {
		return isRayJob
	}
	return !isRayJob
}

// Get the training job from cache or directly
func (rjt *RayJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	rayJob, err := k8saccesser.GetK8sResourceAccesser().GetRayJob(rjt.RayJobClient, namespace, name)
	if err != nil {
		return nil, err
	}
	if err := CheckJobIsOwnedByTrainer(rayJob.Labels); err != nil {
		return nil, err
	}
	// 2. Find the pod list, and determine the pod of the ray job
	rayClusterPods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("release=%v,app=%v", name, rjt.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	rayJobPods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("%v=%v", labelRayJobName, name), "", nil)
	if err != nil {
		return nil, err
	}
	allPods := append(rayClusterPods, rayJobPods...)
	pods, chiefPod := getPodsOfRayJob(rjt, rayJob, allPods)
	// 3. Find the other resources, like statefulset,job
	return &RayJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(pods),
			name:      name,
		},
		rayJob:      rayJob,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: rjt.Type(),
	}, nil

}

func (rjt *RayJobTrainer) isChiefPod(rayJob *rayv1.RayJob, item *corev1.Pod) bool {
	if val, ok := item.Labels[labelRayJobName]; ok && rayJob.Name == val {
		return true
	} else {
		return false
	}
}

// Determine whether it is a pod of RayJobs submitted by Arena
func (rjt *RayJobTrainer) isRayJobPod(name, ns string, pod *corev1.Pod) bool {
	return utils.IsRayJobPod(name, ns, pod)
}

/**
* List Training jobs
 */

func (rjt *RayJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	jobLabels := GetTrainingJobLabels(rjt.Type())
	// list all jobs from k8s apiserver
	rayJobs, err := k8saccesser.GetK8sResourceAccesser().ListRayJobs(rjt.RayJobClient, namespace, jobLabels)
	if err != nil {
		return trainingJobs, err
	}
	rayClusterPods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", rjt.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	rayJobPods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("%v", labelRayJobName), "", nil)
	if err != nil {
		return nil, err
	}
	allPods := append(rayClusterPods, rayJobPods...)
	for _, rayJob := range rayJobs {
		filterPods, chiefPod := getPodsOfRayJob(rjt, rayJob, allPods)
		trainingJobs = append(trainingJobs, &RayJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      rayJob.Name,
			},
			rayJob:      rayJob,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: rjt.Type(),
		})
	}
	return trainingJobs, nil
}

// filter out all pods and chief pod (head pod) of RayJob from pods in current system
func getPodsOfRayJob(rjt *RayJobTrainer, rayJob *rayv1.RayJob, podList []*corev1.Pod) ([]*corev1.Pod, *corev1.Pod) {
	return getPodsOfTrainingJob(rayJob.Name, rayJob.Namespace, podList, rjt.isRayJobPod, func(pod *corev1.Pod) bool {
		return rjt.isChiefPod(rayJob, pod)
	})
}
