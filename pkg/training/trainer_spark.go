package training

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/operators/spark-operator/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/kubeflow/arena/pkg/operators/spark-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// spark application wrapper
type SparkJob struct {
	*BasicJobInfo
	sparkjob    *v1beta2.SparkApplication
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
	return sj.sparkjob.Status.AppState.State == v1beta2.CompletedState
}

func (sj *SparkJob) isFailed() bool {
	return sj.sparkjob.Status.AppState.State == v1beta2.FailedState
}

func (sj *SparkJob) isPending() bool {
	return sj.sparkjob.Status.AppState.State == v1beta2.PendingRerunState
}

func (sj *SparkJob) isSubmitted() bool {
	return sj.sparkjob.Status.AppState.State == v1beta2.SubmittedState
}

func (sj *SparkJob) isRunning() bool {
	return sj.sparkjob.Status.AppState.State == v1beta2.RunningState
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
	// TODO: disable the spark trainer,because there is some bugs to fix
	enable := false
	_, err := config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.SparkCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("SparkJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("SparkJobTrainer is disabled,reason: %v", err)
	}
	sparkjobClient := versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig())
	log.Debugf("Succeed to init SparkJobTrainer")
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
	_, err := st.GetTrainingJob(name, ns)
	return err == nil
}

func (st *SparkJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	sparkJob, err := k8saccesser.GetK8sResourceAccesser().GetSparkJob(st.sparkjobClient, namespace, name)
	if err != nil {
		return nil, err
	}
	if err := CheckJobIsOwnedByTrainer(sparkJob.Labels); err != nil {
		return nil, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("release=%v,app=%v", name, st.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	filterPods, chiefPod := getPodsOfSparkJob(sparkJob, st, pods)
	return &SparkJob{
		BasicJobInfo: &BasicJobInfo{
			resources: podResources(filterPods),
			name:      name,
		},
		sparkjob:    sparkJob,
		chiefPod:    chiefPod,
		pods:        filterPods,
		trainerType: st.Type(),
	}, nil
}

func (st *SparkJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}
	trainingJobs := []TrainingJob{}
	jobLabels := GetTrainingJobLabels(st.Type())
	sparkjobs, err := k8saccesser.GetK8sResourceAccesser().ListSparkJobs(st.sparkjobClient, namespace, jobLabels)
	if err != nil {
		return nil, err
	}
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", st.Type()), "", nil)
	if err != nil {
		return nil, err
	}
	for _, sparkjob := range sparkjobs {
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

func getPodsOfSparkJob(job *v1beta2.SparkApplication, st *SparkJobTrainer, podList []*v1.Pod) (pods []*v1.Pod, chiefPod *v1.Pod) {
	return getPodsOfTrainingJob(job.Name, job.Namespace, podList, st.isSparkPod, func(pod *v1.Pod) bool {
		return st.isChiefPod(pod)
	})
}
