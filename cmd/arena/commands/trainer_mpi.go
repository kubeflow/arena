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

	"github.com/kubeflow/arena/pkg/mpi-operator/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"time"

	v1alpha1 "github.com/kubeflow/arena/pkg/mpi-operator/apis/kubeflow/v1alpha1"
)

var (
	allMPIjobs []v1alpha1.MPIJob
)

func initMPIJobClient() (mpijobClientset *versioned.Clientset, err error) {
	if restConfig == nil {
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	// create the mpijobClientset
	mpijobClientset = versioned.NewForConfigOrDie(restConfig)
	return mpijobClientset, nil
}

// MPI Job Information
type MPIJob struct {
	name         string
	mpijob       v1alpha1.MPIJob
	chiefjob     batchv1.Job
	pods         []v1.Pod // all the pods including statefulset and job
	chiefPod     v1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  string // return trainer type: TENSORFLOW
}

func (mj *MPIJob) Name() string {
	return mj.name
}

// Get the chief Pod of the Job.
func (mj *MPIJob) ChiefPod() v1.Pod {
	return mj.chiefPod
}

// Get the name of the Training Job
// func (mj *MPIJob) Name() string {
// 	return
// }

func (mj *MPIJob) Trainer() string {
	return mj.trainerType
}

// Get all the pods of the Training Job
func (mj *MPIJob) AllPods() []v1.Pod {
	return mj.pods
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
func (mj *MPIJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
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
	for _, pod := range mj.pods {
		mj.requestedGPU += gpuInPod(pod)
	}
	return mj.requestedGPU
}

// Requested GPU count of the Job
func (mj *MPIJob) AllocatedGPU() int64 {
	if mj.allocatedGPU > 0 {
		return mj.allocatedGPU
	}
	for _, pod := range mj.pods {
		mj.allocatedGPU += gpuInActivePod(pod)
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
	trainerType  string
	// check if it's enabled
	enabled bool
}

// NewMPIJobTrainer
func NewMPIJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init MPI job trainer")
	mpijobClient, err := initMPIJobClient()
	if err != nil {
		log.Debugf("unsupported mpijob due to %v", err)
		return &MPIJobTrainer{
			trainerType: "mpijob",
			enabled:     false,
		}
	}
	// allPods have been cached, we do the same to allMPIjobs
	if useCache {
		ns := namespace
		if allNamespaces {
			ns = metav1.NamespaceAll
		}

		mpijobList, err := mpijobClient.KubeflowV1alpha1().MPIJobs(ns).List(metav1.ListOptions{})
		// mpijobList, err := mpijobClient.KubeflowV1alpha2().mpijob(metav1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			log.Debugf("unsupported mpijob due to %v", err)
			return &MPIJobTrainer{
				trainerType: "mpijob",
				enabled:     false,
			}
		}

		for _, mpijob := range mpijobList.Items {
			allMPIjobs = append(allMPIjobs, mpijob)
		}
	}

	return &MPIJobTrainer{
		mpijobClient: mpijobClient,
		client:       client,
		trainerType:  "mpijob",
		enabled:      true,
	}
}

// Get the type
func (tt *MPIJobTrainer) Type() string {
	return tt.trainerType
}

// check if it's TensorFlow job
func (tt *MPIJobTrainer) IsSupported(name, ns string) bool {
	if !tt.enabled {
		return false
	}

	isMPI := false

	if useCache {
		for _, job := range allMPIjobs {
			if tt.isMPIJob(name, ns, job) {
				isMPI = true
				log.Debugf("the job %s for %s in namespace %s is found.", job.Name, name, ns)
				break
			}
		}
	} else {
		mpijobList, err := tt.mpijobClient.KubeflowV1alpha1().MPIJobs(ns).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("release=%s", name),
		})

		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(mpijobList.Items) > 0 {
			isMPI = true
		}
	}

	return isMPI
}

// Get the training job from cache or directly
func (tt *MPIJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	if len(allMPIjobs) > 0 {
		tj, err = tt.getTrainingJobFromCache(name, namespace)
	} else {
		tj, err = tt.getTrainingJob(name, namespace)
	}

	return tj, err
}

func (tt *MPIJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		mpijob v1alpha1.MPIJob
		job    batchv1.Job
	)

	// 0. get the batch job of the mpijob
	job = tt.getChiefJob(name, namespace)

	// 1. Get the batchJob of training Job
	mpijobList, err := tt.mpijobClient.KubeflowV1alpha1().MPIJobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})
	if err != nil {
		return nil, err
	}
	if len(mpijobList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		mpijob = mpijobList.Items[0]
	}

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

	pods, chiefPod := getPodsOfMPIJob(name, tt, podList.Items)

	return &MPIJob{
		mpijob:      mpijob,
		chiefPod:    chiefPod,
		chiefjob:    job,
		pods:        pods,
		name:        name,
		trainerType: tt.Type(),
	}, nil

}

// Get the training job from Cache
func (tt *MPIJobTrainer) getTrainingJobFromCache(name, ns string) (TrainingJob, error) {

	var (
		mpijob v1alpha1.MPIJob
		job    batchv1.Job
	)

	// 0. Find the batch job
	//isChiefJob
	for _, item := range allJobs {
		if tt.isChiefJob(item, name, ns) {
			job = item
			break
		}
	}

	// 1. Find the mpi job
	for _, item := range allMPIjobs {
		if tt.isMPIJob(name, ns, item) {
			mpijob = item
			break
		}
	}

	// 2. Find the pods, and determine the pod of the job
	pods, chiefPod := getPodsOfMPIJob(name, tt, allPods)

	return &MPIJob{
		mpijob:      mpijob,
		chiefPod:    chiefPod,
		pods:        pods,
		name:        name,
		chiefjob:    job,
		trainerType: tt.Type(),
	}, nil
}

func (tt *MPIJobTrainer) getChiefJob(name string, namespace string) (job batchv1.Job) {
	// try to search batch job of the mpijob, it may be name or name-mpijob
	jobList, err := tt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("mpi_job_name=%s", name),
	})

	if len(jobList.Items) > 0 {
		job = jobList.Items[0]
		return job
	}

	if err != nil {
		log.Debugf("mpijob list failed due to %v with mpi_job_name=%s", err, name)
	}

	jobList, err = tt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("mpi_job_name=%s-mpijob", name),
	})

	if len(jobList.Items) > 0 {
		job = jobList.Items[0]
		return job
	}

	if err != nil {
		log.Debugf("mpijob list failed due to %v with mpi_job_name=%s", err, name)
	}

	if len(jobList.Items) > 0 {
		job = jobList.Items[0]
	}

	return job
}

func (tt *MPIJobTrainer) isChiefJob(job batchv1.Job, name string, namespace string) bool {
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
	} else {
		log.Debugf("The job %s is not the chief job of %s", job.Name, name)
	}

	return false
}

func (tt *MPIJobTrainer) isChiefPod(item v1.Pod) bool {

	if val, ok := item.Labels["mpi_role_type"]; ok && (val == "launcher") {
		log.Debugf("the mpijob %s with labels %s", item.Name, val)
	} else {
		return false
	}

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

func (tt *MPIJobTrainer) isMPIPod(name, ns string, item v1.Pod) bool {
	log.Debugf("pod.name: %s: %v", item.Name, item.Labels)
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

	if val, ok := item.Labels["group_name"]; ok && (val == "kubeflow.org") {
		log.Debugf("the mpijob %s with labels %s is found.", item.Name, val)
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
func (tt *MPIJobTrainer) ListTrainingJobs() (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	jobInfos := []types.TrainingJobInfo{}
	for _, mpijob := range allMPIjobs {
		jobInfo := types.TrainingJobInfo{}
		log.Debugf("find mpijob %s in %s", mpijob.Name, mpijob.Namespace)
		if val, ok := mpijob.Labels["release"]; ok && (mpijob.Name == fmt.Sprintf("%s-%s", val, tt.Type())) {
			log.Debugf("the mpijob %s with labels %s found in List", mpijob.Name, val)
			jobInfo.Name = val
		} else {
			jobInfo.Name = mpijob.Name
		}

		jobInfo.Namespace = mpijob.Namespace
		jobInfos = append(jobInfos, jobInfo)
		// jobInfos = append(jobInfos, types.TrainingJobInfo{Name: mpijob.})
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

func getPodsOfMPIJob(name string, tt *MPIJobTrainer, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
	pods = []v1.Pod{}
	for _, item := range podList {
		if !tt.isMPIPod(name, namespace, item) {
			continue
		}
		if tt.isChiefPod(item) && item.CreationTimestamp.After(chiefPod.CreationTimestamp.Time) {
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
