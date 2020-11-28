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

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/arenacache"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"time"
)

var (
	errHorovodJobNotFound = errors.New("horovod job not found")
)

// Horovod Job Information
type HorovodJob struct {
	*BasicJobInfo
	job          *batchv1.Job
	pods         []*v1.Pod // all the pods including statefulset and job
	chiefPod     *v1.Pod   // the pod of job
	gpuCount     int64
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType // return trainer type: MPI, STANDALONE, TENSORFLOW
}

func (h *HorovodJob) Name() string {
	return h.name
}

func (h *HorovodJob) Uid() string {
	return string(h.job.UID)
}

func (h *HorovodJob) Trainer() types.TrainingJobType {
	return h.trainerType
}

// Get the chief Pod of the Job.
func (h *HorovodJob) ChiefPod() *v1.Pod {
	return h.chiefPod
}

// Get all the pods of the Training Job
func (h *HorovodJob) AllPods() []*v1.Pod {
	return h.pods
}

// Requested GPU count of the Job
func (h *HorovodJob) RequestedGPU() int64 {
	if h.requestedGPU > 0 {
		return h.requestedGPU
	}
	for _, pod := range h.pods {
		h.requestedGPU += gpuInPod(*pod)
	}
	return h.requestedGPU
}

// Requested GPU count of the Job
func (h *HorovodJob) AllocatedGPU() int64 {
	if h.allocatedGPU > 0 {
		return h.allocatedGPU
	}
	for _, pod := range h.pods {
		h.allocatedGPU += gpuInActivePod(*pod)
	}
	return h.allocatedGPU
}

func (h *HorovodJob) Age() time.Duration {
	job := h.job
	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.Status.StartTime.Time)
}

// Get the Job Training Duration
func (h *HorovodJob) Duration() time.Duration {
	job := h.job

	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}
	if job.Status.CompletionTime != nil {
		return job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time)
	}

	if h.GetStatus() == "FAILED" {
		cond := getPodLatestCondition(h.ChiefPod())
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Time.Sub(job.Status.StartTime.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", h.ChiefPod().Name)
		}
	}

	return metav1.Now().Sub(job.Status.StartTime.Time)
}

func (h *HorovodJob) StartTime() *metav1.Time {
	return h.job.Status.StartTime
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (h *HorovodJob) GetStatus() (status string) {
	job := h.job
	pod := h.chiefPod
	if job.Status.Active > 0 {
		status = "RUNNING"
	} else if job.Status.Succeeded > 0 {
		status = "SUCCEEDED"
	} else if job.Status.Failed > 0 {
		status = "FAILED"
	}

	if status == "RUNNING" {
		hostIP := pod.Status.HostIP
		if hostIP == "" {
			status = "PENDING"
		} else if pod.Status.Phase == v1.PodPending {
			status = "PENDING"
		}
	}
	return status
}

func (h *HorovodJob) Namespace() string {
	return h.job.Namespace
}

func (h *HorovodJob) GetTrainJob() interface{} {
	return h.job
}

// Get Dashboard url of the job
func (h *HorovodJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
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

	spec := h.chiefPod.Spec
	job := h.job
	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		job.Namespace,
		h.chiefPod.Name,
		spec.Containers[0].Name,
		job.Namespace)
	urls = append(urls, url)
	return urls, nil
}

// Get the hostIP of the chief Pod
func (h *HorovodJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if h.GetStatus() == "RUNNING" {
		hostIP = h.chiefPod.Status.HostIP
	}

	return hostIP
}

// Get PriorityClass
func (hj *HorovodJob) GetPriorityClass() string {
	return ""
}

// Horovod Job trainer
type HorovodJobTrainer struct {
	client      *kubernetes.Clientset
	trainerType types.TrainingJobType
	enabled     bool
}

// Create HorovodJob Trainer
func NewHorovodJobTrainer() Trainer {
	log.Debugf("Init Horovod job trainer")
	return &HorovodJobTrainer{
		client:      config.GetArenaConfiger().GetClientSet(),
		trainerType: types.HorovodTrainingJob,
		enabled:     true,
	}
}

// IsEnabled is used to get the trainer is enable or not
func (h *HorovodJobTrainer) IsEnabled() bool {
	return h.enabled
}

// check if it's Horovod job
func (h *HorovodJobTrainer) IsSupported(name, ns string) bool {
	if !h.enabled {
		return false
	}
	isHorovodJob := false
	if config.GetArenaConfiger().IsDaemonMode() {
		_, err := h.getTrainingJobFromCache(name, ns)
		if err != nil {
			return isHorovodJob
		}
		return !isHorovodJob
	}
	_, err := h.getTrainingJob(name, ns)
	if err != nil {
		return isHorovodJob
	}
	return !isHorovodJob
}

func (h *HorovodJobTrainer) Type() types.TrainingJobType {
	return h.trainerType
}

func (h *HorovodJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	// if arena is daemon mode,get job from cache
	if config.GetArenaConfiger().IsDaemonMode() {
		return h.getTrainingJobFromCache(name, namespace)
	}
	// get job from api server
	return h.getTrainingJob(name, namespace)
}

func (h *HorovodJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	// 1. Get the batchJob of training Job
	jobList, err := h.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s,app=tf-horovod", name),
	})
	if err != nil {
		return nil, err
	}
	if len(jobList.Items) == 0 {
		return nil, errHorovodJobNotFound
	}
	job := jobList.Items[0]
	// 2. Find the pod list, and determine the pod of the job
	podList, err := h.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s,app=tf-horovod", name),
	})
	if err != nil {
		return nil, err
	}
	pods := []*v1.Pod{}
	for _, pod := range podList.Items {
		pods = append(pods, pod.DeepCopy())
	}
	// get chief pod
	filterPods, chiefPod := getPodsOfHorovodJob(h, &job, pods)

	// 3. Find the other resources, like statefulset,job
	return &HorovodJob{
		BasicJobInfo: &BasicJobInfo{
			name:      name,
			resources: podResources(filterPods),
		},
		job:         &job,
		chiefPod:    chiefPod,
		pods:        filterPods,
		trainerType: h.Type(),
	}, nil
}

// Get the training job from Cache
func (h *HorovodJobTrainer) getTrainingJobFromCache(name, namespace string) (TrainingJob, error) {
	// 1.find the k8s job from the cache
	jobs, allPods := arenacache.GetArenaCache().FilterK8sJobs(func(job *batchv1.Job) bool {
		return h.isHorovodJob(name, namespace, job)
	}, func(job *batchv1.Job, pod *v1.Pod) bool {
		return h.isHorovodPod(job.Name, job.Namespace, pod)
	})
	if len(jobs) == 0 {
		return nil, errHorovodJobNotFound
	}
	job := &batchv1.Job{}
	pods := []*v1.Pod{}
	for key, j := range jobs {
		job = j
		pods = allPods[key]
		break
	}
	// 2.find the pods, and determine the pod of the job
	filterPods, chiefPod := getPodsOfHorovodJob(h, job, pods)

	return &HorovodJob{
		BasicJobInfo: &BasicJobInfo{
			name:      name,
			resources: podResources(filterPods),
		},
		job:         job,
		chiefPod:    chiefPod,
		pods:        filterPods,
		trainerType: h.Type(),
	}, nil
}

/**
* List Training jobs
 */
func (h *HorovodJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) (jobs []TrainingJob, err error) {
	// if arena is configured as daemon,getting all mpijobs from cache is corrent
	if config.GetArenaConfiger().IsDaemonMode() {
		return h.listFromCache(namespace, allNamespace)
	}
	return h.listFromAPIServer(namespace, allNamespace)
}

// listFromAPIServer lists the horovod from api server
func (h *HorovodJobTrainer) listFromAPIServer(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	// 1. Get the batchJob of training Job
	jobList, err := h.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release,app=tf-horovod"),
	})
	if err != nil {
		return nil, err
	}
	for _, item := range jobList.Items {
		job := item.DeepCopy()
		// 2. Find the pod list, and determine the pod of the job
		podList, err := h.client.CoreV1().Pods(job.Namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: fmt.Sprintf("release=%s,app=tf-horovod", job.Name),
		})
		if err != nil {
			log.Errorf("failed to get pods of job %v,reason: %v", job.Name, err)
			continue
		}
		pods := []*v1.Pod{}
		for _, pod := range podList.Items {
			pods = append(pods, pod.DeepCopy())
		}
		// get chief pod
		filterPods, chiefPod := getPodsOfHorovodJob(h, job, pods)
		trainingJobs = append(trainingJobs, &HorovodJob{
			BasicJobInfo: &BasicJobInfo{
				name:      job.Name,
				resources: podResources(filterPods),
			},
			job:         job,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: h.Type(),
		})
	}
	return trainingJobs, nil
}

// listFromCache lists mpijobs from arena cache
func (h *HorovodJobTrainer) listFromCache(namespace string, allNamespace bool) ([]TrainingJob, error) {
	// 1.define filter function
	filter := func(job *batchv1.Job) bool { return job.Namespace == namespace }
	// 2.if all namespaces is true,get all mpijobs
	if allNamespace {
		filter = func(job *batchv1.Job) bool { return true }
	}
	jobs, allPods := arenacache.GetArenaCache().FilterK8sJobs(filter, func(job *batchv1.Job, pod *v1.Pod) bool {
		return h.isHorovodPod(job.Name, job.Namespace, pod)
	})
	trainingJobs := []TrainingJob{}
	for key, job := range jobs {
		// find the pods, and determine the pod of the job
		filterPods, chiefPod := getPodsOfHorovodJob(h, job, allPods[key])
		trainingJobs = append(trainingJobs, &HorovodJob{
			BasicJobInfo: &BasicJobInfo{
				name:      job.Name,
				resources: podResources(filterPods),
			},
			job:         job,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: h.Type(),
		})
	}
	return trainingJobs, nil
}

func (h *HorovodJobTrainer) isHorovodJob(name, ns string, job *batchv1.Job) bool {
	if job.Labels["release"] != name {
		return false
	}
	if job.Labels["app"] != "tf-horovod" {
		return false
	}
	if job.Namespace != ns {
		return false
	}
	return true
}

func (h *HorovodJobTrainer) isHorovodPod(name, ns string, pod *v1.Pod) bool {
	return utils.IsHorovodPod(name, ns, pod)
}

func (h *HorovodJobTrainer) isChiefPod(pod *v1.Pod) bool {
	return true
}

// filter out all pods and chief pod (master pod) of mpijob from pods in current system
func getPodsOfHorovodJob(tt *HorovodJobTrainer, job *batchv1.Job, podList []*v1.Pod) ([]*v1.Pod, *v1.Pod) {
	return getPodsOfTrainingJob(job.Name, job.Namespace, podList, tt.isHorovodPod, tt.isChiefPod)
}
