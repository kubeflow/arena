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
	"time"

	cmdTypes "github.com/kubeflow/arena/cmd/arena/types"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	common "github.com/kubeflow/arena/cmd/arena/commands/mpi/api/common/v1"
	mpi "github.com/kubeflow/arena/cmd/arena/commands/mpi/api/v1alpha2"
	mpiClient "github.com/kubeflow/arena/cmd/arena/commands/mpi/client/clientset/versioned"
)

var (
	allMPIjobs []MPIJob
)

// MPI Job Information
type MPIJob struct {
	*cmdTypes.BasicJobInfo
	mpijob       mpi.MPIJob
	chiefjob     batchv1.Job
	pods         []v1.Pod // all the pods including statefulset and job
	chiefPod     v1.Pod   // the chief pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  string // return trainer type: TENSORFLOW
	podMetadata  metav1.ObjectMeta
	imageName    string
}

func (mj *MPIJob) Name() string {
	return mj.mpijob.Name
}

func (mj *MPIJob) Uid() string {
	return string(mj.mpijob.UID)
}

// Get the chief Pod of the Job.
func (mj *MPIJob) ChiefPod() *v1.Pod {
	return &mj.chiefPod
}

// Get the name of the Training Job
// func (mj *MPIJob) Name() string {
// 	return
// }

func (mj *MPIJob) Trainer() string {
	return mj.trainerType
}

func (mj *MPIJob) CreatedByCLI() bool {
	return true
}

func (mj *MPIJob) Image() (status string) {
	return mj.mpijob.Annotations["image"]
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (mj *MPIJob) GetStatus() (status string) {
	status = "UNKNOWN"
	if mj.mpijob.Name == "" {
		return status
	}

	if mj.isSucceeded() {
		status = "Succeeded"
	} else if mj.isFailed() {
		status = "Failed"
	} else if mj.isPending() {
		status = "Pending"
	} else {
		status = string(mj.chiefPod.Status.Phase)
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
	// urls := []string{}
	// // dashboardURL, err := dashboard(client, "kubeflow", "tf-job-dashboard")
	// dashboardURL, err := dashboard(client, namespace, "kubernetes-dashboard")

	// if err != nil {
	// 	log.Debugf("Get dashboard failed due to %v", err)
	// 	// retry for the existing customers, will be deprecated in the future
	// 	dashboardURL, err = dashboard(client, arenaNamespace, "kubernetes-dashboard")
	// 	if err != nil {
	// 		log.Debugf("Get dashboard failed due to %v", err)
	// 	}
	// }

	// if err != nil {
	// 	log.Debugf("Get dashboard failed due to %v", err)
	// 	// retry for the existing customers, will be deprecated in the future
	// 	dashboardURL, err = dashboard(client, "kube-system", "kubernetes-dashboard")
	// 	if err != nil {
	// 		log.Debugf("Get dashboard failed due to %v", err)
	// 	}
	// }

	// if dashboardURL == "" {
	// 	return urls, fmt.Errorf("No LOGVIEWER Installed.")
	// }

	// if len(mj.chiefPod.Spec.Containers) == 0 {
	// 	return urls, fmt.Errorf("mpi launcher is not ready!")
	// }

	// url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
	// 	dashboardURL,
	// 	mj.chiefPod.Namespace,
	// 	mj.chiefPod.Name,
	// 	mj.chiefPod.Spec.Containers[0].Name,
	// 	mj.chiefPod.Namespace)

	// urls = append(urls, url)

	// return urls, nil

	return []string{}, nil
}

// Requested GPU count of the Job
func (mj *MPIJob) RequestedGPU() float64 {
	if mj.requestedGPU > 0 {
		return float64(mj.requestedGPU)
	}
	for _, pod := range mj.pods {
		mj.requestedGPU += gpuInPod(pod)
	}
	return float64(mj.requestedGPU)
}

// Requested GPU count of the Job
func (mj *MPIJob) AllocatedGPU() float64 {
	if mj.allocatedGPU > 0 {
		return float64(mj.allocatedGPU)
	}
	for _, pod := range mj.pods {
		mj.allocatedGPU += int64(gpuInActivePod(pod))
	}
	return float64(mj.allocatedGPU)
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
	client       kubernetes.Interface
	mpiclientset mpiClient.Interface
	trainerType  string
	// check if it's enabled
	enabled bool
}

// NewMPIJobTrainer
func NewMPIJobTrainer(kubeClient client.Client) Trainer {
	resourcesList, err := kubeClient.GetClientset().Discovery().ServerResourcesForGroupVersion("kubeflow.org/v1alpha2")

	if err != nil {
		return &MPIJobTrainer{
			trainerType: "mpijob",
			enabled:     false,
		}
	}

	for _, resource := range resourcesList.APIResources {
		if resource.Kind == "MPIJob" {
			return &MPIJobTrainer{
				client:       kubeClient.GetClientset(),
				mpiclientset: mpiClient.NewForConfigOrDie(kubeClient.GetRestConfig()),
				trainerType:  "mpijob",
				enabled:      true,
			}
		}
	}

	return &MPIJobTrainer{
		enabled: false,
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

	mpiJobs, err := tt.getMpiJobs(ns, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err == nil && len(mpiJobs) > 0 {
		return true
	}

	return false
}

// Get the training job from cache or directly
func (tt *MPIJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	return tt.getTrainingJob(name, namespace)
}

func (tt *MPIJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		mpijob mpi.MPIJob
		job    batchv1.Job
	)

	// 0. get the batch job of the mpijob
	job = tt.getChiefJob(name, namespace)

	mpiJobs, err := tt.getMpiJobs(namespace, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch mpijob due to %s", err.Error())
	}

	if len(mpiJobs) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		mpijob = mpiJobs[0]
	}

	// 2. Find the pod list, and determine the pod of the job
	podList, err := tt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("mpi_job_name=%s", name),
	})

	if err != nil {
		return nil, err
	}

	pods, chiefPod := getPodsOfMPIJob(name, namespace, tt, podList.Items)

	// 3. Find the other resources, like statefulset,job
	resources, err := tt.resources(name, namespace, pods)
	if err != nil {
		return nil, err
	}
	return &MPIJob{
		BasicJobInfo: cmdTypes.NewBasicJobInfo(name, resources),
		mpijob:       mpijob,
		chiefPod:     chiefPod,
		chiefjob:     job,
		pods:         pods,
		trainerType:  tt.Type(),
	}, nil

}

// Get the training job from Cache
func (tt *MPIJobTrainer) getTrainingJobInfo(name string, ns string, mpiJob mpi.MPIJob, allPods []v1.Pod, allJobs []batchv1.Job) (TrainingJob, error) {

	var (
		job batchv1.Job
	)

	// 0. Find the batch job
	//isChiefJob
	for _, item := range allJobs {
		if tt.isChiefJob(item, name, ns) {
			job = item
			break
		}
	}

	// 2. Find the pods, and determine the pod of the job
	pods, chiefPod := getPodsOfMPIJob(name, ns, tt, allPods)

	return &MPIJob{
		BasicJobInfo: cmdTypes.NewBasicJobInfo(name, cmdTypes.PodResources(pods)),
		mpijob:       mpiJob,
		chiefPod:     chiefPod,
		pods:         pods,
		chiefjob:     job,
		trainerType:  tt.Type(),
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
		return true
	} else {
		log.Debugf("The job %s is not the chief job of %s", job.Name, name)
	}

	return false
}

func (tt *MPIJobTrainer) isChiefPod(item v1.Pod) bool {
	if val, ok := item.Labels["mpi_role_type"]; ok && (val == "launcher") {
		return true
	}

	return false
}

func (tt *MPIJobTrainer) isMPIJob(name, ns string, item mpi.MPIJob) bool {
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

func (tt *MPIJobTrainer) isPodOfMPiJob(name, ns string, item v1.Pod) bool {
	if value, ok := item.ObjectMeta.Labels["mpi_job_name"]; ok && (value == name) {
		return true
	}

	return false
}

func IsMPIPod(item v1.Pod) bool {
	_, ok := item.ObjectMeta.Labels["mpi_job_name"]
	return ok
}

func (tt *MPIJobTrainer) resources(name string, namespace string, pods []v1.Pod) ([]cmdTypes.Resource, error) {
	resources := []cmdTypes.Resource{}

	// 2. Find the pod list, and determine the pod of the job
	stsList, err := tt.client.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("mpi_job_name=%s", name),
	})
	if err != nil {
		return resources, err
	}
	for _, sts := range stsList.Items {
		resources = append(resources, cmdTypes.Resource{
			Name:         sts.Name,
			Uid:          string(sts.UID),
			ResourceType: cmdTypes.ResourceTypeStatefulSet,
		})
	}

	// 2. Find the pod list, and determine the pod of the job
	jobs, err := tt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("mpi_job_name=%s", name),
	})
	if err != nil {
		return resources, err
	}
	for _, job := range jobs.Items {
		resources = append(resources, cmdTypes.Resource{
			Name:         job.Name,
			Uid:          string(job.UID),
			ResourceType: cmdTypes.ResourceTypeJob,
		})
	}
	resources = append(resources, cmdTypes.PodResources(pods)...)
	return resources, nil
}

func (tt *MPIJobTrainer) IsEnabled() bool {
	return tt.enabled
}

func (tt *MPIJobTrainer) getMpiJobs(namespace string, listOptions metav1.ListOptions) ([]mpi.MPIJob, error) {

	mpiJobList, err := tt.mpiclientset.KubeflowV1alpha2().MPIJobs(namespace).List(listOptions)
	// mpiResource := schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1alpha2", Resource: "mpijobs"}
	// mpijobListUnstructured, err := tt.dynamicClient.Resource(mpiResource).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return []mpi.MPIJob{}, err
	}

	// mpiJobsList := []mpi.MPIJob{}

	// for _, mpiJobUnstructered := range mpijobListUnstructured.Items {
	// 	var mpiJob mpi.MPIJob
	// 	err = mapstructure.Decode(mpiJobUnstructered.Object, &mpiJob)
	// 	if err != nil {
	// 		return mpiJobsList, err
	// 	}
	// }

	return mpiJobList.Items, nil
}

/**
* List Training jobs
 */
func (tt *MPIJobTrainer) ListTrainingJobs(namespace string) (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}

	mpiJobs, err := tt.getMpiJobs(namespace, metav1.ListOptions{})

	if err != nil {
		return []TrainingJob{}, err
	}

	podsList, err := tt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return []TrainingJob{}, nil
	}

	jobsList, err := tt.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	if err != nil {
		return []TrainingJob{}, nil
	}

	for _, mpijob := range mpiJobs {

		jobInfo := types.TrainingJobInfo{}
		if val, ok := mpijob.Labels["release"]; ok && (mpijob.Name == fmt.Sprintf("%s-%s", val, tt.Type())) {
			jobInfo.Name = val
		} else {
			jobInfo.Name = mpijob.Name
		}

		job, err := tt.getTrainingJobInfo(jobInfo.Name, jobInfo.Namespace, mpijob, podsList.Items, jobsList.Items)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (mj *MPIJob) isSucceeded() bool {
	return hasCondition(mj.mpijob.Status, common.JobSucceeded)
	// return mj.mpijob.Status.LauncherStatus == v1alpha2.LauncherSucceeded
}

func (mj *MPIJob) isFailed() bool {
	return hasCondition(mj.mpijob.Status, common.JobFailed)
	// return mj.mpijob.Status.LauncherStatus == v1alpha2.LauncherFailed
}

func (mj *MPIJob) isPending() bool {
	// return false
	if len(mj.chiefjob.Name) == 0 {
		log.Debugf("The MPIJob is pending due to chiefJob is not ready")
		return true
	}

	return false
}

func hasCondition(status common.JobStatus, condType common.JobConditionType) bool {
	for _, condition := range status.Conditions {
		if condition.Type == condType && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func (mj *MPIJob) Project() string {
	return mj.chiefPod.ObjectMeta.Labels["project"]
}

func (mj *MPIJob) User() string {
	return mj.chiefPod.ObjectMeta.Labels["user"]
}

// Get all the pods of the Training Job
func (mj *MPIJob) AllPods() []v1.Pod {
	return mj.pods
}

// Get all the kubernetes resource of the Training Job
func (mj *MPIJob) Resources() []cmdTypes.Resource {
	return mj.BasicJobInfo.Resources()
}

// Get PriorityClass
func (m *MPIJob) GetPriorityClass() string {
	return ""
}

func getPodsOfMPIJob(name string, namespace string, tt *MPIJobTrainer, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
	pods = []v1.Pod{}
	for _, item := range podList {
		if !tt.isPodOfMPiJob(name, namespace, item) {
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
	}
	return pods, chiefPod
}
