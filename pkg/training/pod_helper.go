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

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

// acquire all active pods from all namespaces
func acquireAllActivePods(client *kubernetes.Clientset) ([]v1.Pod, error) {
	allPods := []v1.Pod{}

	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return allPods, err
	}
	nodeNonTerminatedPodsList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return allPods, err
	}

	for _, pod := range nodeNonTerminatedPodsList.Items {
		allPods = append(allPods, pod)
	}
	return allPods, nil
}

func acquireAllPods(client *kubernetes.Clientset, namespace string, allNamespaces bool) ([]v1.Pod, error) {
	allPods := []v1.Pod{}
	ns := namespace
	if allNamespaces {
		ns = metav1.NamespaceAll
	}
	podList, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return allPods, err
	}
	for _, pod := range podList.Items {
		allPods = append(allPods, pod)
	}
	return allPods, nil
}

func acquireAllJobs(client *kubernetes.Clientset, namespace string, allNamespaces bool) ([]batchv1.Job, error) {
	allJobs := []batchv1.Job{}
	ns := namespace
	if allNamespaces {
		ns = metav1.NamespaceAll
	}
	jobList, err := client.BatchV1().Jobs(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return allJobs, err
	}
	for _, job := range jobList.Items {
		allJobs = append(allJobs, job)
	}
	return allJobs, nil
}

func getPodNameFromJob(client kubernetes.Interface, namespace, name string) (podName string, err error) {
	pods, err := jobPods(client, namespace, name)
	if err != nil {
		return "", err
	}
	if len(pods) == 0 {
		return "", fmt.Errorf("Failed to find the pod for job %s, maybe you need to set --namespace", name)
	}

	for _, pod := range pods {
		meta := pod.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				break
			}
		}
		if isJob {
			return pod.Name, nil
		}
	}
	return "", fmt.Errorf("getPodNameFromJob: Failed to find the pod of job")
}

// Get the latest pod from the Kubernetes job
func getPodFromJob(client kubernetes.Interface, jobName, namespace, releaseName string) (jobPod v1.Pod, err error) {
	pods, err := jobPods(client, namespace, releaseName)
	if err != nil {
		return jobPod, err
	}
	if len(pods) == 0 {
		return jobPod, fmt.Errorf("getPodFromJob: Failed to find the pod for job %s, maybe you need to set --namespace", jobName)
	}
	var latest metav1.Time
	for _, pod := range pods {
		meta := pod.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind != "Job" {
				continue
			}
			isJob = true
			break
		}
		if !isJob {
			continue
		}
		// return pod, nil
		if jobPod.Name == "" {
			latest = pod.CreationTimestamp
			jobPod = pod
			log.Debugf("set pod %s as first jobpod, and it's time is %v", jobPod.Name, jobPod.CreationTimestamp)
			continue
		}
		log.Debugf("current jobpod %s , and it's time is %v", jobPod.Name, latest)
		log.Debugf("candidate jobpod %s , and it's time is %v", pod.Name, pod.CreationTimestamp)
		current := pod.CreationTimestamp
		if !latest.Before(&current) {
			log.Debugf("no replace")
			continue
		}
		jobPod = pod
		latest = current
		log.Debugf("replace")
	}
	if jobPod.Name == "" {
		err = fmt.Errorf("Not able to job with release %s in pods %v", releaseName, pods)
	}
	return jobPod, err
}

// List all the pods which associate to the arena jobs, including the pods in the statefulset and the job
func listAllPodsForJob(client kubernetes.Interface, jobName, namespace string, releaseName string) (pods []v1.Pod, err error) {
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", releaseName),
	})
	if err != nil {
		return nil, err
	}
	pods = []v1.Pod{}
	for _, item := range podList.Items {
		meta := item.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind != "Job" {
				continue
			}
			isJob = true
			log.Debugf("find job pod %v, break", item)
			break
		}
		if !isJob {
			pods = append(pods, item)
			log.Debugf("add pod %v to pods", item)
		}
	}
	jobPod, err := getPodFromJob(client, jobName, namespace, releaseName)
	if err != nil {
		return nil, err
	}
	pods = append(pods, jobPod)
	return pods, err
}

func jobPods(client kubernetes.Interface, namespace string, releaseName string) ([]v1.Pod, error) {
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", releaseName),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, err
}

// Sort the pod condition by time.
type SortPodConditionByLastTransitionTime []v1.PodCondition

func (s SortPodConditionByLastTransitionTime) Len() int      { return len(s) }
func (s SortPodConditionByLastTransitionTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortPodConditionByLastTransitionTime) Less(i, j int) bool {
	// return s[i].CreatedAt.Before(s[j].CreatedAt)
	return s[i].LastTransitionTime.Time.After(s[j].LastTransitionTime.Time)
}

func makePodConditionsSortedByTime(conditions []v1.PodCondition) []v1.PodCondition {
	newCondtions := make(SortPodConditionByLastTransitionTime, 0, len(conditions))
	for _, c := range conditions {
		newCondtions = append(newCondtions, c)
	}
	sort.Sort(newCondtions)
	return []v1.PodCondition(newCondtions)
}

func getPodLatestCondition(pod *v1.Pod) v1.PodCondition {
	cond := v1.PodCondition{}
	conditions := makePodConditionsSortedByTime(pod.Status.Conditions)
	if len(conditions) <= 0 {
		log.Debugf("the pod %s's conditions %v is empty", pod.Name, conditions)
		return cond
	}
	cond = conditions[0]
	log.Debugf("the pod %s's conditions %v is not empty", pod.Name, conditions)
	return cond
}
