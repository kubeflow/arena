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
	"sort"
	"strings"
	"sync"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/arenacache"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util/kubectl"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var trainers map[types.TrainingJobType]Trainer

var once sync.Once

func GetAllTrainers() map[types.TrainingJobType]Trainer {
	once.Do(func() {
		locker := new(sync.RWMutex)
		trainers = map[types.TrainingJobType]Trainer{}
		trainerInits := []func() Trainer{
			NewTensorFlowJobTrainer,
			NewPyTorchJobTrainer,
			NewMPIJobTrainer,
			NewETJobTrainer,
			NewHorovodJobTrainer,
			NewVolcanoJobTrainer,
			NewSparkJobTrainer,
		}
		var wg sync.WaitGroup
		for _, initFunc := range trainerInits {
			wg.Add(1)
			f := initFunc
			go func() {
				defer wg.Done()
				trainer := f()
				locker.Lock()
				trainers[trainer.Type()] = trainer
				locker.Unlock()
			}()
		}
		wg.Wait()
	})
	return trainers
}

type orderedTrainingJob []TrainingJob

func (this orderedTrainingJob) Len() int {
	return len(this)
}

func (this orderedTrainingJob) Less(i, j int) bool {
	return this[i].RequestedGPU() > this[j].RequestedGPU()
}

func (this orderedTrainingJob) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type orderedTrainingJobByAge []TrainingJob

func (this orderedTrainingJobByAge) Len() int {
	return len(this)
}

func (this orderedTrainingJobByAge) Less(i, j int) bool {
	if this[i].StartTime() == nil {
		return true
	} else if this[j].StartTime() == nil {
		return false
	}

	return this[i].StartTime().After(this[j].StartTime().Time)
}

func (this orderedTrainingJobByAge) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func makeTrainingJobOrderdByAge(jobList []TrainingJob) []TrainingJob {
	newJoblist := make(orderedTrainingJobByAge, 0, len(jobList))
	for _, v := range jobList {
		newJoblist = append(newJoblist, v)
	}
	sort.Sort(newJoblist)
	return []TrainingJob(newJoblist)
}

func makeTrainingJobOrderdByGPUCount(jobList []TrainingJob) []TrainingJob {
	newJoblist := make(orderedTrainingJob, 0, len(jobList))
	for _, v := range jobList {
		newJoblist = append(newJoblist, v)
	}
	sort.Sort(newJoblist)
	return []TrainingJob(newJoblist)
}

func getPodsOfTrainingJob(name, namespace string, podList []*v1.Pod, isTrainingJobPod func(name, namespace string, pod *v1.Pod) bool, isChiefPod func(pod *v1.Pod) bool) ([]*v1.Pod, *v1.Pod) {
	pods := []*v1.Pod{}
	var (
		pendingChiefPod     *v1.Pod
		nonePendingChiefPod *v1.Pod
	)
	for _, item := range podList {
		if !isTrainingJobPod(name, namespace, item) {
			continue
		}
		pods = append(pods, item)
		if !isChiefPod(item) {
			log.Debugf("the pod %v is not chief pod", item.Name)
			continue
		}
		if item.Status.Phase == v1.PodPending {
			if pendingChiefPod == nil {
				pendingChiefPod = item
			}
			if item.CreationTimestamp.After(pendingChiefPod.CreationTimestamp.Time) {
				pendingChiefPod = item
			}
			continue
		}
		// set the chief pod
		if nonePendingChiefPod == nil {
			nonePendingChiefPod = item
		}
		// If there are some failed chiefPod, and the new chiefPod haven't started, set the latest failed pod as chief pod
		if item.CreationTimestamp.After(nonePendingChiefPod.CreationTimestamp.Time) {
			nonePendingChiefPod = item
		}
	}
	if nonePendingChiefPod != nil {
		return pods, nonePendingChiefPod
	}
	if pendingChiefPod == nil {
		return pods, &v1.Pod{}
	}
	return pods, pendingChiefPod
}

func CheckOperatorIsInstalled(crdName string) bool {
	crdNames, err := kubectl.GetCrdNames()
	log.Debugf("get all crd names: %v", crdNames)
	if err != nil {
		log.Debugf("failed to get crd names,reason: %v", err)
		return false
	}
	for _, name := range crdNames {
		if name == crdName {
			return true
		}
	}
	return false
}

func listJobPods(k8sclient *kubernetes.Clientset, namespace string, podLabels map[string]string) ([]*v1.Pod, error) {
	pods := []*v1.Pod{}
	podList := &v1.PodList{}
	selectors, err := parseLabelSelectors(podLabels)
	if err != nil {
		return nil, err
	}
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().List(context.Background(), podList, client.InNamespace(namespace), &client.ListOptions{LabelSelector: selectors})
	} else {
		podList, err = k8sclient.CoreV1().Pods(namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: selectors.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		pods = append(pods, pod.DeepCopy())
	}
	return pods, nil
}

func listStatefulSets(k8sclient *kubernetes.Clientset, namespace string, stsLabels map[string]string) ([]*appv1.StatefulSet, error) {
	statefulsets := []*appv1.StatefulSet{}
	stsList := &appv1.StatefulSetList{}
	selectors, err := parseLabelSelectors(stsLabels)
	if err != nil {
		return nil, err
	}
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().List(context.Background(), stsList, client.InNamespace(namespace), &client.ListOptions{LabelSelector: selectors})
	} else {
		// 2. Find the pod list, and determine the pod of the job
		stsList, err = k8sclient.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: selectors.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, sts := range stsList.Items {
		statefulsets = append(statefulsets, sts.DeepCopy())
	}
	return statefulsets, nil
}

func listJobBatchJobs(k8sclient *kubernetes.Clientset, namespace string, jobLabels map[string]string) ([]*batchv1.Job, error) {
	jobs := []*batchv1.Job{}
	jobList := &batchv1.JobList{}
	selectors, err := parseLabelSelectors(jobLabels)
	if err != nil {
		return nil, err
	}
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().List(context.Background(), jobList, client.InNamespace(namespace), &client.ListOptions{LabelSelector: selectors})
	} else {
		jobList, err = k8sclient.BatchV1().Jobs(namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: selectors.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, j := range jobList.Items {
		jobs = append(jobs, j.DeepCopy())
	}
	return jobs, nil
}

func ListServices(k8sclient *kubernetes.Clientset, namespace string, svcLabels map[string]string) ([]*v1.Service, error) {
	return listServices(k8sclient, namespace, svcLabels)
}
func listServices(k8sclient *kubernetes.Clientset, namespace string, svcLabels map[string]string) ([]*v1.Service, error) {
	services := []*v1.Service{}
	serviceList := &v1.ServiceList{}
	selectors, err := parseLabelSelectors(svcLabels)
	if err != nil {
		return nil, err
	}
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().List(context.Background(), serviceList, client.InNamespace(namespace), &client.ListOptions{LabelSelector: selectors})
	} else {
		serviceList, err = k8sclient.CoreV1().Services(namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: selectors.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, svc := range serviceList.Items {
		services = append(services, svc.DeepCopy())
	}
	return services, nil
}

func ListNodes(k8sclient *kubernetes.Clientset, nodeLabels map[string]string) ([]*v1.Node, error) {
	return listNodes(k8sclient, nodeLabels)
}

func listNodes(k8sclient *kubernetes.Clientset, nodeLabels map[string]string) ([]*v1.Node, error) {
	nodeList := &v1.NodeList{}
	nodes := []*v1.Node{}
	selectors, err := parseLabelSelectors(nodeLabels)
	if err != nil {
		return nil, err
	}
	if config.GetArenaConfiger().IsDaemonMode() {
		err = arenacache.GetCacheClient().List(context.Background(), nodeList, &client.ListOptions{LabelSelector: selectors})
	} else {
		nodeList, err = k8sclient.CoreV1().Nodes().List(metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		nodes = append(nodes, node.DeepCopy())
	}
	return nodes, nil
}

func parseLabelSelectors(objectLabels map[string]string) (labels.Selector, error) {
	labelSelector := []string{}
	for key, value := range objectLabels {
		if value != "" {
			labelSelector = append(labelSelector, fmt.Sprintf("%v=%v", key, value))
			continue
		}
		labelSelector = append(labelSelector, key)
	}
	selector, err := labels.Parse(strings.Join(labelSelector, ","))
	if err != nil {
		log.Errorf("failed to parse label selectors,reason: %v", err)
		return nil, err
	}
	return selector, nil
}
