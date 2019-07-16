package commands

import (
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/spark-operator/apis/sparkoperator.k8s.io/v1beta1"
	"github.com/kubeflow/arena/pkg/spark-operator/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// all spark jobs cache
var (
	allSparkJobs []v1beta1.SparkApplication
)

// spark application wrapper
type SparkJob struct {
	name        string
	sparkjob    v1beta1.SparkApplication
	trainerType string
	pods        []v1.Pod
	chiefPod    v1.Pod
}

func (sj *SparkJob) Name() string {
	return sj.name
}

// return driver pod
func (sj *SparkJob) ChiefPod() v1.Pod {
	return sj.chiefPod
}

// return trainerType: sparkjob
func (sj *SparkJob) Trainer() string {
	return sj.trainerType
}

// return pods from cache
func (sj *SparkJob) AllPods() []v1.Pod {
	return sj.pods
}

/*
				spark job driver state
	-------------------------------------------------------
	NewState              ApplicationStateType = ""
	SubmittedState        ApplicationStateType = "SUBMITTED"
	RunningState          ApplicationStateType = "RUNNING"
	CompletedState        ApplicationStateType = "COMPLETED"
	FailedState           ApplicationStateType = "FAILED"
	FailedSubmissionState ApplicationStateType = "SUBMISSION_FAILED"
	PendingRerunState     ApplicationStateType = "PENDING_RERUN"
	InvalidatingState     ApplicationStateType = "INVALIDATING"
	SucceedingState       ApplicationStateType = "SUCCEEDING"
	FailingState          ApplicationStateType = "FAILING"
	UnknownState          ApplicationStateType = "UNKNOWN"


				spark job executor state
	-------------------------------------------------------
	ExecutorPendingState   ExecutorState = "PENDING"
	ExecutorRunningState   ExecutorState = "RUNNING"
	ExecutorCompletedState ExecutorState = "COMPLETED"
	ExecutorFailedState    ExecutorState = "FAILED"
	ExecutorUnknownState   ExecutorState = "UNKNOWN"
*/
func (sj *SparkJob) GetStatus() (status string) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("spark job may not complete,because of ", r)
		}
		return
	}()

	status = "UNKNOWN"

	// name is empty when the pod has not been scheduled
	if sj.sparkjob.Name == "" {
		return status
	}

	if sj.isSucceeded() {
		status = "SUCCEEDED"
	} else if sj.isFailed() {
		status = "FAILED"
	} else if sj.isPending() {
		status = "PENDING"
	} else if sj.isSubmitted() {
		status = "SUBMITTED"
	} else if sj.isRunning() {
		status = "RUNNING"
	} else {
		status = string(sj.sparkjob.Status.AppState.State)
	}

	return status
}

func (sj *SparkJob) isSucceeded() bool {
	return sj.sparkjob.Status.AppState.State == v1beta1.CompletedState
}

func (sj *SparkJob) isFailed() bool {
	return sj.sparkjob.Status.AppState.State == v1beta1.FailedState
}

func (sj *SparkJob) isPending() bool {
	return sj.sparkjob.Status.AppState.State == v1beta1.PendingRerunState
}

func (sj *SparkJob) isSubmitted() bool {
	return sj.sparkjob.Status.AppState.State == v1beta1.SubmittedState
}

func (sj *SparkJob) isRunning() bool {
	return sj.sparkjob.Status.AppState.State == v1beta1.RunningState
}

func (sj *SparkJob) StartTime() *metav1.Time {
	return &sj.sparkjob.CreationTimestamp
}

func (sj *SparkJob) Age() time.Duration {
	job := sj.sparkjob

	if job.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.CreationTimestamp.Time)
}

// Get the Job Training Duration
func (sj *SparkJob) Duration() time.Duration {
	sparkjob := sj.sparkjob

	if sparkjob.CreationTimestamp.IsZero() {
		return 0
	}

	if sparkjob.Status.TerminationTime.IsZero() {
		return 0
	}

	//todo
	return sparkjob.Status.TerminationTime.Sub(sparkjob.CreationTimestamp.Time)
}

func (sj *SparkJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
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

	if len(sj.chiefPod.Spec.Containers) == 0 {
		return urls, fmt.Errorf("spark driver pod is not ready!")
	}

	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		sj.chiefPod.Namespace,
		sj.chiefPod.Name,
		sj.chiefPod.Spec.Containers[0].Name,
		sj.chiefPod.Namespace)

	urls = append(urls, url)

	return urls, nil
}

// spark job without gpu supported
func (sj *SparkJob) RequestedGPU() int64 {
	return 0
}

// spark job without gpu supported
func (sj *SparkJob) AllocatedGPU() int64 {
	return 0
}

// Get the hostIP of the driver Pod
func (sj *SparkJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if sj.GetStatus() == "RUNNING" {
		hostIP = sj.chiefPod.Status.HostIP
	}
	return hostIP
}

func (sj *SparkJob) Namespace() string {
	return sj.sparkjob.Namespace
}

// Get PriorityClass TODO: @moyuan
func (sj *SparkJob) GetPriorityClass() string {
	return ""
}

func NewSparkJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init Spark job trainer")
	sparkjobClient, err := initSparkJobClient()

	if err != nil {
		log.Debugf("unsupported spark job due to %v", err)
		return &SparkJobTrainer{
			trainerType: defaultSparkJobTrainingType,
			enabled:     false,
		}
	}
	// allPods have been cached, we do the same to allSparkJobs
	if useCache {
		ns := namespace
		if allNamespaces {
			ns = metav1.NamespaceAll
		}

		sparkJobList, err := sparkjobClient.SparkoperatorV1beta1().SparkApplications(ns).List(metav1.ListOptions{})
		if err != nil {
			log.Debugf("unsupported sparkJob due to %v", err)
			return &SparkJobTrainer{
				trainerType: defaultSparkJobTrainingType,
				enabled:     false,
			}
		}

		for _, sparkJob := range sparkJobList.Items {
			allSparkJobs = append(allSparkJobs, sparkJob)
		}
	}

	return &SparkJobTrainer{
		sparkjobClient: sparkjobClient,
		client:         client,
		trainerType:    defaultSparkJobTrainingType,
		enabled:        true,
	}
}

// init spark job client
func initSparkJobClient() (sparkjobClientset *versioned.Clientset, err error) {
	if restConfig == nil {
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	sparkjobClientset = versioned.NewForConfigOrDie(restConfig)
	return sparkjobClientset, nil
}

// spark job trainer
type SparkJobTrainer struct {
	client         *kubernetes.Clientset
	sparkjobClient *versioned.Clientset
	trainerType    string
	enabled        bool
}

func (st *SparkJobTrainer) Type() string {
	return st.trainerType
}

func (st *SparkJobTrainer) IsSupported(name, ns string) bool {
	if !st.enabled {
		return false
	}

	isSpark := false

	if useCache {
		for _, job := range allSparkJobs {
			if st.isSparkJob(name, ns, job) {
				isSpark = true
				break
			}
		}
	} else {
		sparkJobList, err := st.sparkjobClient.SparkoperatorV1beta1().SparkApplications(ns).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("release=%s", name),
		})

		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(sparkJobList.Items) > 0 {
			isSpark = true
		}
	}

	return isSpark
}

func (st *SparkJobTrainer) isSparkJob(name, ns string, job v1beta1.SparkApplication) bool {
	if val, ok := job.Labels["release"]; ok && (val == name) {
		log.Debugf("the sparkjob %s with labels %s", job.Name, val)
	} else {
		return false
	}

	if val, ok := job.Labels["app"]; ok && (val == "sparkjob") {
		log.Debugf("the sparkjob %s with labels %s is found.", job.Name, val)
	} else {
		return false
	}

	if job.Namespace != ns {
		return false
	}
	return true
}

func (st *SparkJobTrainer) GetTrainingJob(name, namespace string) (job TrainingJob, err error) {
	if len(allSparkJobs) > 0 {
		job, err = st.getTrainingJobFromCache(name, namespace)
	} else {
		job, err = st.getTrainingJob(name, namespace)
	}

	return job, err
}

func (st *SparkJobTrainer) getTrainingJobFromCache(name, namespace string) (job TrainingJob, err error) {
	var (
		sparkJob v1beta1.SparkApplication
	)

	for _, item := range allSparkJobs {
		if st.isSparkJob(name, namespace, item) {
			sparkJob = item
			break
		}
	}

	pods, chiefPod := getPodsOfSparkJob(name, st, allPods)

	return &SparkJob{
		chiefPod:    chiefPod,
		sparkjob:    sparkJob,
		pods:        pods,
		name:        name,
		trainerType: st.Type(),
	}, nil
}

func (st *SparkJobTrainer) getTrainingJob(name, namespace string) (job TrainingJob, err error) {
	var (
		sparkjob v1beta1.SparkApplication
	)

	sparkjobList, err := st.sparkjobClient.SparkoperatorV1beta1().SparkApplications(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s", name),
	})
	if err != nil {
		return nil, err
	}
	if len(sparkjobList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		sparkjob = sparkjobList.Items[0]
	}

	podList, err := st.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, err
	}

	pods, chiefPod := getPodsOfSparkJob(name, st, podList.Items)

	return &SparkJob{
		sparkjob:    sparkjob,
		chiefPod:    chiefPod,
		pods:        pods,
		name:        name,
		trainerType: st.Type(),
	}, nil
}

func (st *SparkJobTrainer) ListTrainingJobs() (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	jobInfos := []types.TrainingJobInfo{}
	for _, sparkJob := range allSparkJobs {
		jobInfo := types.TrainingJobInfo{}
		log.Debugf("find sparkjob %s in %s", sparkJob.Name, sparkJob.Namespace)
		if val, ok := sparkJob.Labels["release"]; ok && (sparkJob.Name == fmt.Sprintf("%s-%s", val, st.Type())) {
			log.Debugf("the sparkjob %s with labels %s found in List", sparkJob.Name, val)
			jobInfo.Name = val
		} else {
			jobInfo.Name = sparkJob.Name
		}

		jobInfo.Namespace = sparkJob.Namespace
		jobInfos = append(jobInfos, jobInfo)
	}
	log.Debugf("jobInfos %v", jobInfos)

	for _, jobInfo := range jobInfos {
		job, err := st.getTrainingJobFromCache(jobInfo.Name, jobInfo.Namespace)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (st *SparkJobTrainer) isSparkPod(name, ns string, item v1.Pod) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the sparkjob %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "sparkjob") {
		log.Debugf("the sparkjob %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func (st *SparkJobTrainer) isChiefPod(item v1.Pod) bool {
	if val, ok := item.Labels["spark-role"]; ok && (val == "driver") {
		log.Debugf("the sparkjob %s with labels %s", item.Name, val)
	} else {
		return false
	}
	return true
}

func getPodsOfSparkJob(name string, st *SparkJobTrainer, podList []v1.Pod) (pods []v1.Pod, chiefPod v1.Pod) {
	pods = []v1.Pod{}
	for _, item := range podList {
		if !st.isSparkPod(name, namespace, item) {
			continue
		}
		if st.isChiefPod(item) && item.CreationTimestamp.After(chiefPod.CreationTimestamp.Time) {
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
