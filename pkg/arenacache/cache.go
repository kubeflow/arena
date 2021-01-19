package arenacache

import (
	"fmt"
	"sync"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	etv1alpha1 "github.com/AliyunContainerService/et-operator/pkg/api/v1alpha1"
	v1alpha1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v1alpha1"
	pytorchv1 "github.com/kubeflow/pytorch-operator/pkg/apis/pytorch/v1"
	sparkv1beta1 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta1"
	tfv1 "github.com/kubeflow/tf-operator/pkg/apis/tensorflow/v1"
	volv1alpha1 "github.com/volcano.sh/volcano/pkg/apis/batch/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

var once sync.Once
var arenaCache *ArenaCache

func GetArenaCache() *ArenaCache {
	once.Do(func() {
		locker := new(sync.RWMutex)
		arenaCache = &ArenaCache{
			locker:      locker,
			arenaConfig: config.GetArenaConfiger(),
			pods:        map[string]*v1.Pod{},
			tfjobs:      map[string]*tfv1.TFJob{},
			pyjobs:      map[string]*pytorchv1.PyTorchJob{},
			mpijobs:     map[string]*v1alpha1.MPIJob{},
			k8sjobs:     map[string]*batchv1.Job{},
			volcanojobs: map[string]*volv1alpha1.Job{},
			etjobs:      map[string]*etv1alpha1.TrainingJob{},
			sparkjobs:   map[string]*sparkv1beta1.SparkApplication{},
			deployments: map[string]*appv1.Deployment{},
			services:    map[string]*v1.Service{},
			genKey:      func(namespace, name string) string { return fmt.Sprintf("%v/%v", namespace, name) },
		}
	})
	return arenaCache
}

type ArenaCache struct {
	arenaConfig *config.ArenaConfiger
	pods        map[string]*v1.Pod
	tfjobs      map[string]*tfv1.TFJob
	pyjobs      map[string]*pytorchv1.PyTorchJob
	mpijobs     map[string]*v1alpha1.MPIJob
	k8sjobs     map[string]*batchv1.Job
	volcanojobs map[string]*volv1alpha1.Job
	etjobs      map[string]*etv1alpha1.TrainingJob
	sparkjobs   map[string]*sparkv1beta1.SparkApplication
	deployments map[string]*appv1.Deployment
	services    map[string]*v1.Service
	genKey      func(namespace, name string) string
	locker      *sync.RWMutex
}

// AddOrUpdatePod adds or updates pod
func (a *ArenaCache) AddOrUpdatePod(pods ...*v1.Pod) {
	a.locker.Lock()
	defer a.locker.Unlock()
	for _, pod := range pods {
		key := a.genKey(pod.Namespace, pod.Name)
		a.pods[key] = pod.DeepCopy()
	}
}

// DeletePod deletes the pod
func (a *ArenaCache) DeletePod(namespace, name string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	delete(a.pods, a.genKey(namespace, name))
}

func (a *ArenaCache) deletePodsByFilter(filter func(pod *v1.Pod) bool) {
	for _, pod := range a.pods {
		if filter(pod) {
			delete(a.pods, a.genKey(pod.Namespace, pod.Name))
		}
	}
}

// GetPod returns the target pod,if pod not exists,return null
func (a *ArenaCache) GetPod(namespace string, name string) *v1.Pod {
	a.locker.RLock()
	defer a.locker.RUnlock()
	return a.pods[a.genKey(namespace, name)]
}

// FilterPods filter pods that we needs
func (a *ArenaCache) FilterPods(filter func(pod *v1.Pod) bool) []*v1.Pod {
	a.locker.RLock()
	defer a.locker.RUnlock()
	return a.filterPods(filter)
}

// FilterPods filter pods that we needs
func (a *ArenaCache) filterPods(filter func(pod *v1.Pod) bool) []*v1.Pod {
	pods := []*v1.Pod{}
	for _, pod := range a.pods {
		if filter(pod) {
			pods = append(pods, pod)
		}
	}
	return pods
}

// AddOrUpdateDeployment adds or updates the deployment
func (a *ArenaCache) AddOrUpdateDeployment(deployment *appv1.Deployment) {
	a.locker.Lock()
	defer a.locker.Unlock()
	a.deployments[a.genKey(deployment.Namespace, deployment.Name)] = deployment
}

// DeleteDeployment deletes the depoyment
func (a *ArenaCache) DeleteDeployment(namespace, name string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	delete(a.deployments, a.genKey(namespace, name))
}

// AddOrUpdateService adds or updates the services
func (a *ArenaCache) AddOrUpdateService(s *v1.Service) {
	a.locker.Lock()
	defer a.locker.Unlock()
	a.services[a.genKey(s.Namespace, s.Name)] = s
}

// DeleteService delets the service
func (a *ArenaCache) DeleteService(namespace, name string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	delete(a.services, a.genKey(namespace, name))
}

// AddOrUpdateJob adds or updates the job,includes: job,tfjob
func (a *ArenaCache) AddOrUpdateJob(job interface{}) {
	a.locker.Lock()
	defer a.locker.Unlock()
	switch v := job.(type) {
	case *tfv1.TFJob:
		a.tfjobs[a.genKey(v.Namespace, v.Name)] = v
	case *pytorchv1.PyTorchJob:
		a.pyjobs[a.genKey(v.Namespace, v.Name)] = v
	case *v1alpha1.MPIJob:
		a.mpijobs[a.genKey(v.Namespace, v.Name)] = v
	case *batchv1.Job:
		a.k8sjobs[a.genKey(v.Namespace, v.Name)] = v
	case *volv1alpha1.Job:
		a.volcanojobs[a.genKey(v.Namespace, v.Name)] = v
	case *etv1alpha1.TrainingJob:
		a.etjobs[a.genKey(v.Namespace, v.Name)] = v
	case *sparkv1beta1.SparkApplication:
		a.sparkjobs[a.genKey(v.Namespace, v.Name)] = v
	}
}

func (a *ArenaCache) DeleteTrainingJob(namespace, name string, jobType types.TrainingJobType) {
	a.locker.Lock()
	defer a.locker.Unlock()
	var isMatched func(namespace, name string, pod *v1.Pod) bool
	switch jobType {
	case types.PytorchTrainingJob:
		isMatched = utils.IsPyTorchPod
		delete(a.pyjobs, a.genKey(namespace, name))
	case types.TFTrainingJob:
		isMatched = utils.IsTensorFlowPod
		delete(a.tfjobs, a.genKey(namespace, name))
	case types.MPITrainingJob:
		isMatched = utils.IsMPIPod
		delete(a.mpijobs, a.genKey(namespace, name))
	case types.VolcanoTrainingJob:
		isMatched = utils.IsVolcanoPod
		delete(a.volcanojobs, a.genKey(namespace, name))
	case types.ETTrainingJob:
		isMatched = utils.IsETPod
		delete(a.etjobs, a.genKey(namespace, name))
	case types.SparkTrainingJob:
		isMatched = utils.IsSparkPod
		delete(a.sparkjobs, a.genKey(namespace, name))
	default:
		return
	}
	for key, pod := range a.pods {
		if !isMatched(namespace, name, pod) {
			continue
		}
		delete(a.pods, key)
	}
}

func (a *ArenaCache) DeleteK8sJob(namespace, name string, podFilter func(pod *v1.Pod) bool) {
	a.locker.Lock()
	defer a.locker.Unlock()
	delete(a.k8sjobs, a.genKey(namespace, name))
	for key, pod := range a.pods {
		if !podFilter(pod) {
			continue
		}
		delete(a.pods, key)
	}
}

// GetAllTFJobs returns all tfjobs and their pods
func (a *ArenaCache) FilterTFJobs(filter func(tfjob *tfv1.TFJob) bool) (map[string]*tfv1.TFJob, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*tfv1.TFJob{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.tfjobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsTensorFlowPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// FilterPytorchJobs returns all pytorchjobs and their pods
func (a *ArenaCache) FilterPytorchJobs(filter func(pyjob *pytorchv1.PyTorchJob) bool) (map[string]*pytorchv1.PyTorchJob, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*pytorchv1.PyTorchJob{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.pyjobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsPyTorchPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// FilterMPIJobs returns all mpijobs and their pods
func (a *ArenaCache) FilterMPIJobs(filter func(pyjob *v1alpha1.MPIJob) bool) (map[string]*v1alpha1.MPIJob, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*v1alpha1.MPIJob{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.mpijobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsMPIPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// FilterVolcanoJobs returns all volcanojobs and their pods
func (a *ArenaCache) FilterVolcanoJobs(filter func(job *volv1alpha1.Job) bool) (map[string]*volv1alpha1.Job, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*volv1alpha1.Job{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.volcanojobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsVolcanoPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// FilterETJobs returns all etjobs and their pods
func (a *ArenaCache) FilterETJobs(filter func(pyjob *etv1alpha1.TrainingJob) bool) (map[string]*etv1alpha1.TrainingJob, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*etv1alpha1.TrainingJob{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.etjobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsETPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

func (a *ArenaCache) FilterSparkJobs(filter func(job *sparkv1beta1.SparkApplication) bool) (map[string]*sparkv1beta1.SparkApplication, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*sparkv1beta1.SparkApplication{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.sparkjobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsSparkPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// FilterK8SJobs returns all mpijobs and their pods
func (a *ArenaCache) FilterK8sJobs(jobFilter func(job *batchv1.Job) bool, podFilter func(job *batchv1.Job, pod *v1.Pod) bool) (map[string]*batchv1.Job, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*batchv1.Job{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.k8sjobs {
		if !jobFilter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !podFilter(job, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// FilterDeployments filters deployments
func (a *ArenaCache) FilterDeployments(filter func(*appv1.Deployment) bool, podFilter func(*appv1.Deployment, *v1.Pod) bool) (map[string]*appv1.Deployment, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	deployments := map[string]*appv1.Deployment{}
	pods := map[string][]*v1.Pod{}
	for key, deployment := range a.deployments {
		if !filter(deployment) {
			continue
		}
		deployments[key] = deployment.DeepCopy()
		pods[key] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !podFilter(deployment, pod) {
				continue
			}
			pods[key] = append(pods[key], pod.DeepCopy())
		}
	}
	return deployments, pods
}

// FilterServices filters services
func (a *ArenaCache) FilterServices(filter func(*v1.Service) bool) []*v1.Service {
	a.locker.RLock()
	defer a.locker.RUnlock()
	services := []*v1.Service{}
	for _, s := range a.services {
		if !filter(s) {
			continue
		}
		services = append(services, s.DeepCopy())
	}
	return services
}

// GetTFJob returns the tfjob
func (a *ArenaCache) GetTFJob(namespace, name string) (*tfv1.TFJob, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.tfjobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsTensorFlowPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

// get the pytorch job
func (a *ArenaCache) GetPytorchJob(namespace, name string) (*pytorchv1.PyTorchJob, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.pyjobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsPyTorchPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

// get the mpi job
func (a *ArenaCache) GetMPIJob(namespace, name string) (*v1alpha1.MPIJob, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.mpijobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsMPIPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

// get the volcano job
func (a *ArenaCache) GetVolcanoJob(namespace, name string) (*volv1alpha1.Job, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.volcanojobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsVolcanoPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

// get the et job
func (a *ArenaCache) GetETJob(namespace, name string) (*etv1alpha1.TrainingJob, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.etjobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsETPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

func (a *ArenaCache) GetSparkJob(namespace, name string) (*sparkv1beta1.SparkApplication, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.sparkjobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsSparkPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

// get the k8s job
func (a *ArenaCache) GetK8sJob(namespace, name string, podFilter func(pod *v1.Pod) bool) (*batchv1.Job, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.k8sjobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !podFilter(pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

// get the k8s deployment
func (a *ArenaCache) GetDeployment(namespace, name string, podFilter func(pod *v1.Pod) bool) (*appv1.Deployment, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	deployment := a.deployments[a.genKey(namespace, name)]
	if deployment == nil {
		return deployment, pods
	}
	for _, pod := range a.pods {
		if !podFilter(pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return deployment.DeepCopy(), pods
}

// get the k8s service
func (a *ArenaCache) GetService(namespace, name string, podFilter func(pod *v1.Pod) bool) (*v1.Service, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	service := a.services[a.genKey(namespace, name)]
	if service == nil {
		return service, pods
	}
	for _, pod := range a.pods {
		if !podFilter(pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return service.DeepCopy(), pods
}
