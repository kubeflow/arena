package training

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/arenacache"
	"github.com/kubeflow/arena/pkg/operators/spark-operator/apis/sparkoperator.k8s.io/v1beta1"
	"github.com/kubeflow/arena/pkg/operators/spark-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	errSparkJobNotFound = errors.New("sparkjob not found")
	SparkCRD            = "sparkapplications.sparkoperator.k8s.io"
)

// spark application wrapper
type SparkJob struct {
	*BasicJobInfo
	sparkjob    *v1beta1.SparkApplication
	trainerType types.TrainingJobType
	pods        []*v1.Pod
	chiefPod    *v1.Pod
}

func (sj *SparkJob) Name() string {
	return sj.name
}

func (sj *SparkJob) Uid() string {
	return string(sj.sparkjob.UID)
}

// return driver pod
func (sj *SparkJob) ChiefPod() *v1.Pod {
	return sj.chiefPod
}

// return trainerType: sparkjob
func (sj *SparkJob) Trainer() types.TrainingJobType {
	return sj.trainerType
}

// return pods from cache
func (sj *SparkJob) AllPods() []*v1.Pod {
	return sj.pods
}

func (sj *SparkJob) GetTrainJob() interface{} {
	return sj.sparkjob
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

func (sj *SparkJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
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

// spark job trainer
type SparkJobTrainer struct {
	client         *kubernetes.Clientset
	sparkjobClient *versioned.Clientset
	trainerType    types.TrainingJobType
	enabled        bool
}

func NewSparkJobTrainer() Trainer {
	log.Debugf("Init Spark job trainer")
	// TODO: disable the spark trainer,because there is some bugs to fix
	enable := false
	sparkjobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	return &SparkJobTrainer{
		sparkjobClient: sparkjobClient,
		client:         config.GetArenaConfiger().GetClientSet(),
		trainerType:    types.SparkTrainingJob,
		enabled:        enable,
	}
}

func (st *SparkJobTrainer) IsEnabled() bool {
	return st.enabled
}

func (st *SparkJobTrainer) Type() types.TrainingJobType {
	return st.trainerType
}

func (st *SparkJobTrainer) IsSupported(name, ns string) bool {
	if !st.enabled {
		return false
	}
	if config.GetArenaConfiger().IsDaemonMode() {
		_, err := st.getTrainingJobFromCache(name, ns)
		// if found the job,return true
		return err == nil
	}
	_, err := st.getTrainingJob(name, ns)
	return err == nil
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
	// if arena is daemon mode,get job from cache
	if config.GetArenaConfiger().IsDaemonMode() {
		return st.getTrainingJobFromCache(name, namespace)
	}
	// get job from api server
	return st.getTrainingJob(name, namespace)
}

func (st *SparkJobTrainer) getTrainingJobFromCache(name, namespace string) (TrainingJob, error) {
	// 1.find the mpijob from the cache
	sparkjob, pods := arenacache.GetArenaCache().GetSparkJob(namespace, name)
	if sparkjob == nil {
		return nil, errSparkJobNotFound
	}
	// 2. Find the pods, and determine the pod of the job
	filterPods, chiefPod := getPodsOfSparkJob(sparkjob, st, pods)
	return &SparkJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(filterPods),
			name:      name,
		},
		chiefPod:    chiefPod,
		sparkjob:    sparkjob,
		pods:        filterPods,
		trainerType: st.Type(),
	}, nil
}

func (st *SparkJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	sparkjob, err := st.sparkjobClient.SparkoperatorV1beta1().SparkApplications(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Debugf("failed to get job,reason: %v", err)
		if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, SparkCRD, name)) {
			return nil, errSparkJobNotFound
		}
		return nil, err
	}
	podList, err := st.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s,app=%v", name, st.trainerType),
	})
	if err != nil {
		return nil, err
	}
	pods := []*v1.Pod{}
	for _, pod := range podList.Items {
		pods = append(pods, pod.DeepCopy())
	}
	filterPods, chiefPod := getPodsOfSparkJob(sparkjob, st, pods)

	return &SparkJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(filterPods),
			name:      name,
		},
		sparkjob:    sparkjob,
		chiefPod:    chiefPod,
		pods:        filterPods,
		trainerType: st.Type(),
	}, nil
}

func (st *SparkJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) (jobs []TrainingJob, err error) {
	// if arena is configured as daemon,getting all mpijobs from cache is corrent
	if config.GetArenaConfiger().IsDaemonMode() {
		return st.listFromCache(namespace, allNamespace)
	}
	return st.listFromAPIServer(namespace, allNamespace)
}

// listFromAPIServer lists the sparkjobs from api server
func (st *SparkJobTrainer) listFromAPIServer(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	sparkJobList, err := st.sparkjobClient.SparkoperatorV1beta1().SparkApplications(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release"),
	})
	if err != nil {
		return nil, err
	}
	for _, item := range sparkJobList.Items {
		sparkjob := item.DeepCopy()
		podList, err := st.client.CoreV1().Pods(sparkjob.Namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: fmt.Sprintf("release=%s,app=%v", sparkjob.Name, st.trainerType),
		})
		if err != nil {
			return nil, err
		}
		pods := []*v1.Pod{}
		for _, pod := range podList.Items {
			pods = append(pods, pod.DeepCopy())
		}
		filterPods, chiefPod := getPodsOfSparkJob(sparkjob, st, pods)
		trainingJobs = append(trainingJobs, &SparkJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      sparkjob.Name,
			},
			sparkjob:    sparkjob,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: st.Type(),
		})
	}
	return trainingJobs, nil
}

// listFromCache lists sparkjobs from arena cache
func (st *SparkJobTrainer) listFromCache(namespace string, allNamespace bool) ([]TrainingJob, error) {
	// 1.define filter function
	filter := func(job *v1beta1.SparkApplication) bool { return job.Namespace == namespace }
	trainingJobs := []TrainingJob{}
	// 2.if all namespaces is true,get all mpijobs
	if allNamespace {
		filter = func(job *v1beta1.SparkApplication) bool { return true }
	}
	jobs, pods := arenacache.GetArenaCache().FilterSparkJobs(filter)
	for key, job := range jobs {
		filterPods, chiefPod := getPodsOfSparkJob(job, st, pods[key])
		trainingJobs = append(trainingJobs, &SparkJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      job.Name,
			},
			sparkjob:    job,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: st.Type(),
		})
	}
	return trainingJobs, nil
}

func (st *SparkJobTrainer) isSparkPod(name, ns string, item *v1.Pod) bool {
	return utils.IsSparkPod(name, ns, item)
}

func (st *SparkJobTrainer) isChiefPod(item *v1.Pod) bool {
	if val, ok := item.Labels["spark-role"]; ok && (val == "driver") {
		log.Debugf("the sparkjob %s with labels %s", item.Name, val)
	} else {
		return false
	}
	return true
}

func getPodsOfSparkJob(job *v1beta1.SparkApplication, st *SparkJobTrainer, podList []*v1.Pod) (pods []*v1.Pod, chiefPod *v1.Pod) {
	return getPodsOfTrainingJob(job.Name, job.Namespace, podList, st.isSparkPod, func(pod *v1.Pod) bool {
		return st.isChiefPod(pod)
	})
}
