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
	"sync"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util/kubectl"
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
			NewVolcanoJobTrainer,
			NewSparkJobTrainer,
			NewDeepSpeedJobTrainer,
			NewRayJobTrainer,
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

func GetTrainingJobLabels(jobType types.TrainingJobType) string {
	l := fmt.Sprintf("app=%v,release", jobType)
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsAdminUser() || !arenaConfiger.IsIsolateUserInNamespace() {
		log.Debugf("list training jobs by labels: %v", l)
		return l
	}
	l = fmt.Sprintf("%v,%v=%v", l, types.UserNameIdLabel, arenaConfiger.GetUser().GetId())
	log.Debugf("list training jobs by label: %v", l)
	return l
}

func CheckJobIsOwnedByTrainer(labels map[string]string) error {
	arenaConfiger := config.GetArenaConfiger()
	// if not enabled isolate namespace feature,return nil
	if !arenaConfiger.IsIsolateUserInNamespace() {
		return nil
	}
	// if current user is admin user,return nil
	if arenaConfiger.IsAdminUser() {
		return nil
	}
	// if current user is matched the job user,return nil
	if labels[types.UserNameIdLabel] == arenaConfiger.GetUser().GetId() {
		return nil
	}
	return types.ErrNoPrivilegesToOperateJob
}

// CompatibleJobCRD Compatible with training-operator CRD.
func CompatibleJobCRD(crdName, fieldToCheck string) bool {
	arenaConfiger := config.GetArenaConfiger()

	tfCRD, err := arenaConfiger.GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crdName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Get tensorflow crd failed, error: %s", err)
		return false
	}

	compatible := false
	for _, version := range tfCRD.Spec.Versions {
		if _, ok := version.Schema.OpenAPIV3Schema.Properties["spec"].Properties[fieldToCheck]; ok {
			compatible = true
			break
		}
	}

	return compatible
}
