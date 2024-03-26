package analyze

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type modelProcessor struct {
	client  *kubernetes.Clientset
	jobType types.ModelJobType
}

func NewModelProcessor(jobType types.ModelJobType) Processor {
	p := &modelProcessor{
		jobType: jobType,
		client:  config.GetArenaConfiger().GetClientSet(),
	}
	return p
}

func (p *modelProcessor) Type() types.ModelJobType {
	return p.jobType
}

func (p *modelProcessor) GetModelJob(namespace, name string) (ModelJob, error) {
	job, err := k8saccesser.GetK8sResourceAccesser().GetJob(name, namespace)
	if err != nil {
		return nil, err
	}

	if p.jobType == types.AllModelJob {
		p.jobType = utils.TransferModelJobType(job.ObjectMeta.Labels["type"])
	}

	selector := fmt.Sprintf("app=modeljob,release=%s,type=%s", name, p.jobType)
	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, selector, "", nil)
	if err != nil {
		return nil, err
	}

	return &modelJob{
		name:      job.Name,
		namespace: job.Namespace,
		jobType:   p.jobType,
		job:       job,
		pods:      pods,
	}, nil
}

func (p *modelProcessor) ListModelJobs(namespace string, allNamespaces bool) ([]ModelJob, error) {
	if allNamespaces {
		namespace = metav1.NamespaceAll
	}

	var jobSelector string
	if p.jobType == types.AllModelJob {
		jobSelector = "app=modeljob"
	} else {
		jobSelector = fmt.Sprintf("app=modeljob, type=%s", p.jobType)
	}
	jobs, err := k8saccesser.GetK8sResourceAccesser().ListJobs(namespace, jobSelector, "", nil)
	if err != nil {
		return nil, err
	}

	var modelJobs []ModelJob

	for _, job := range jobs {
		var singleJobType types.ModelJobType
		if p.jobType == types.AllModelJob {
			singleJobType = utils.TransferModelJobType(job.ObjectMeta.Labels["type"])
		}

		jobName := job.ObjectMeta.Name
		podSelector := fmt.Sprintf("app=modeljob,release=%s,type=%s", jobName, singleJobType)
		pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, podSelector, "", nil)
		if err != nil {
			log.Errorf("list pods of job %s in namespace %s failed", jobName, namespace)
			return nil, err
		}

		modelJob := &modelJob{
			name:      job.Name,
			namespace: job.Namespace,
			jobType:   singleJobType,
			job:       job,
			pods:      pods,
		}

		modelJobs = append(modelJobs, modelJob)
	}

	return modelJobs, nil
}

type modelJob struct {
	name      string
	namespace string
	jobType   types.ModelJobType
	pods      []*v1.Pod
	job       *batchv1.Job
}

func (m *modelJob) Uid() string {
	return string(m.job.UID)
}

func (m *modelJob) Name() string {
	return m.name
}

func (m *modelJob) Namespace() string {
	return m.namespace
}

func (m *modelJob) Type() types.ModelJobType {
	return m.jobType
}

func (m *modelJob) Pods() []*v1.Pod {
	return m.pods
}

func (m *modelJob) Job() *batchv1.Job {
	return m.job
}

func (m *modelJob) Age() time.Duration {
	return time.Since(m.job.ObjectMeta.CreationTimestamp.Time)
}

func (m *modelJob) Duration() time.Duration {
	job := m.job

	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}

	if !job.Status.CompletionTime.IsZero() {
		return job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time)
	}

	//if pj.GetStatus() == "FAILED" {
	//	cond := getPodLatestCondition(pj.chiefPod)
	//	if !cond.LastTransitionTime.IsZero() {
	//		return cond.LastTransitionTime.Time.Sub(pytorchjob.CreationTimestamp.Time)
	//	} else {
	//		log.Debugf("the latest condition's time is zero of pod %s", pj.chiefPod.Name)
	//	}
	//}

	return metav1.Now().Sub(job.Status.StartTime.Time)
}

func (m *modelJob) Status() string {
	if m.job.Status.Active > 0 {
		return string(types.ModelJobRunning)
	}

	if m.job.Status.Succeeded > 0 {
		return string(types.ModelJobComplete)
	}

	if m.job.Status.Failed > 0 {
		return string(types.ModelJobFailed)
	}

	return string(types.ModelJobUnknown)
}

func (m *modelJob) StartTime() *metav1.Time {
	return &m.job.ObjectMeta.CreationTimestamp
}

func (m *modelJob) RequestCPUs() int64 {
	var podCPUs int64
	for _, c := range m.job.Spec.Template.Spec.Containers {
		if val, ok := c.Resources.Limits[v1.ResourceName(types.CPUResourceName)]; ok {
			podCPUs += val.Value()
		}
	}
	return podCPUs
}

func (m *modelJob) RequestGPUs() int64 {
	var podGPUs int64
	for _, c := range m.job.Spec.Template.Spec.Containers {
		if val, ok := c.Resources.Limits[v1.ResourceName(types.NvidiaGPUResourceName)]; ok {
			podGPUs += val.Value()
		}
		if val, ok := c.Resources.Limits[v1.ResourceName(types.AliyunGPUResourceName)]; ok {
			podGPUs += val.Value()
		}
	}
	return podGPUs
}

func (m *modelJob) RequestGPUMemory() int64 {
	var podGPUMemory int64
	for _, c := range m.job.Spec.Template.Spec.Containers {
		if val, ok := c.Resources.Limits[v1.ResourceName(types.GPUShareResourceName)]; ok {
			podGPUMemory += val.Value()
		}
	}
	return podGPUMemory
}

func (m *modelJob) RequestGPUCore() int64 {
	var podGPUCore int64
	for _, c := range m.job.Spec.Template.Spec.Containers {
		if val, ok := c.Resources.Limits[v1.ResourceName(types.GPUCoreShareResourceName)]; ok {
			podGPUCore += val.Value()
		}
	}
	return podGPUCore
}

func (m *modelJob) Instances() []types.ModelJobInstance {
	var instances []types.ModelJobInstance
	for index, pod := range m.pods {
		status, totalContainers, restart, readyContainer := utils.DefinePodPhaseStatus(*pod)
		age := util.ShortHumanDuration(time.Since(pod.ObjectMeta.CreationTimestamp.Time))
		gpuMemory := utils.GPUMemoryCountInPod(pod)
		gpuCore := utils.GPUCoreCountInPod(pod)
		gpus := getPodGPUs(pod, gpuMemory, index)
		instances = append(instances, types.ModelJobInstance{
			Name:              pod.Name,
			Status:            status,
			Age:               age,
			NodeIP:            pod.Status.HostIP,
			NodeName:          pod.Spec.NodeName,
			IP:                pod.Status.PodIP,
			ReadyContainer:    readyContainer,
			TotalContainer:    totalContainers,
			RestartCount:      restart,
			RequestGPUs:       gpus,
			RequestGPUMemory:  gpuMemory,
			RequestGPUCore:    gpuCore,
			CreationTimestamp: pod.CreationTimestamp.Unix(),
		})
	}
	return instances
}

func (m *modelJob) Params() map[string]string {
	commands := m.job.Spec.Template.Spec.Containers[0].Command
	command := commands[len(commands)-1]
	arr := strings.Split(command, " ")
	params := make(map[string]string)
	for _, value := range arr {
		if strings.HasPrefix(value, "--") {
			kv := strings.Split(value, "=")
			params[fmt.Sprintf("--%s", kv[0])] = value[len(kv[0])+1:]
		}
	}
	return params
}

func (m *modelJob) Convert2JobInfo() types.ModelJobInfo {
	modelJobType := types.ModelTypeMap[m.jobType].Alias

	servingJobInfo := types.ModelJobInfo{
		UUID:              m.Uid(),
		Name:              m.name,
		Namespace:         m.namespace,
		Type:              modelJobType,
		Age:               util.ShortHumanDuration(m.Age()),
		Duration:          util.ShortHumanDuration(m.Duration()),
		Status:            m.Status(),
		RequestCPUs:       m.RequestCPUs(),
		RequestGPUs:       m.RequestGPUs(),
		RequestGPUMemory:  m.RequestGPUMemory(),
		RequestGPUCore:    m.RequestGPUCore(),
		Instances:         m.Instances(),
		CreationTimestamp: m.StartTime().Unix(),
		Params:            m.Params(),
	}
	return servingJobInfo
}

func getPodGPUs(pod *v1.Pod, gpuMemory int, index int) float64 {
	if utils.IsCompletedPod(pod) {
		return float64(0)
	}
	if pod.Status.Phase != v1.PodRunning {
		return float64(0)
	}
	if len(pod.Spec.NodeName) == 0 {
		return float64(0)
	}
	if gpuMemory != 0 {
		nodeGPUMemory := getNodeGPUMemory(pod.Spec.NodeName)
		if index == 0 {
			log.Debugf("node name: %v,single gpu memory: %vGiB\n", pod.Spec.NodeName, nodeGPUMemory)
		}
		if nodeGPUMemory == float64(0) {
			return float64(0)
		}
		return math.Round(float64(gpuMemory)/nodeGPUMemory*10) / 10
	}
	return float64(utils.GPUCountInPod(pod) + utils.AliyunGPUCountInPod(pod))
}

func getNodeGPUMemory(nodeName string) float64 {
	node, err := k8saccesser.GetK8sResourceAccesser().GetNode(nodeName)
	if err != nil {
		log.Debugf("failed to get node gpu memory,reason: %v", err)
		return float64(0)
	}
	totalGPUs := getResourceOfGPUShareNode(node, types.GPUShareCountName)
	totalGPUMemory := getResourceOfGPUShareNode(node, types.GPUShareResourceName)
	if totalGPUs == 0 {
		return float64(0)
	}
	return totalGPUMemory / totalGPUs
}

//
//func getNodeGPUCore(nodeName string) int64 {
//	node, err := k8saccesser.GetK8sResourceAccesser().GetNode(nodeName)
//	if err != nil {
//		log.Debugf("failed to get node gpu core,reason: %v", err)
//		return int64(0)
//	}
//	totalGPUs := getResourceOfGPUShareNode(node, types.GPUShareCountName)
//	totalGPUMemory := getResourceOfGPUShareNode(node, types.GPUCoreShareResourceName)
//	if totalGPUs == 0 {
//		return int64(0)
//	}
//	return totalGPUMemory / totalGPUs
//}

func getResourceOfGPUShareNode(node *v1.Node, resourceName string) float64 {
	val, ok := node.Status.Capacity[v1.ResourceName(resourceName)]
	if !ok {
		return 0
	}
	return float64(val.Value())
}
