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
	"sort"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// Sort the pod condition by time.
type SortPodConditionByLastTransitionTime []corev1.PodCondition

func (s SortPodConditionByLastTransitionTime) Len() int      { return len(s) }
func (s SortPodConditionByLastTransitionTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortPodConditionByLastTransitionTime) Less(i, j int) bool {
	// return s[i].CreatedAt.Before(s[j].CreatedAt)
	return s[i].LastTransitionTime.After(s[j].LastTransitionTime.Time)
}

func makePodConditionsSortedByTime(conditions []corev1.PodCondition) []corev1.PodCondition {
	newCondtions := make(SortPodConditionByLastTransitionTime, 0, len(conditions))
	for _, c := range conditions {
		newCondtions = append(newCondtions, c)
	}
	sort.Sort(newCondtions)
	return []corev1.PodCondition(newCondtions)
}

func getPodLatestCondition(pod *corev1.Pod) corev1.PodCondition {
	cond := corev1.PodCondition{}
	conditions := makePodConditionsSortedByTime(pod.Status.Conditions)
	if len(conditions) <= 0 {
		log.Debugf("the pod %s's conditions %v is empty", pod.Name, conditions)
		return cond
	}
	cond = conditions[0]
	log.Debugf("the pod %s's conditions %v is not empty", pod.Name, conditions)
	return cond
}
