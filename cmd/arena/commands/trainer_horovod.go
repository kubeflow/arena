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

	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Horovod Job Information
type HorovodJob struct {
	*JobInfo
}

// Get the chief Pod of the Job.
func (hj *HorovodJob) ChiefPod() v1.Pod {
	return hj.jobPod
}

// Get the name of the Training Job
// func (hj *HorovodJob) Name() string {
// 	return
// }

// Get all the pods of the Training Job
func (hj *HorovodJob) AllPods() []v1.Pod {
	return hj.pods
}

// Get Dashboard url of the job
func (hj *HorovodJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
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

	spec := hj.jobPod.Spec
	job := hj.job
	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		job.Namespace,
		hj.jobPod.Name,
		spec.Containers[0].Name,
		job.Namespace)

	urls = append(urls, url)

	return urls, nil
}

// Get the hostIP of the chief Pod
func (hj *HorovodJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if hj.GetStatus() == "RUNNING" {
		hostIP = hj.jobPod.Status.HostIP
	}

	return hostIP
}

// Get PriorityClass
func (hj *HorovodJob) GetPriorityClass() string {
	return ""
}

// Horovod Job trainer
type HorovodJobTrainer struct {
	client      *kubernetes.Clientset
	trainerType string
}

// Create HorovodJob Trainer
func NewHorovodJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init Horovod job trainer")

	return &HorovodJobTrainer{
		client:      client,
		trainerType: "horovodjob",
	}
}

// check if it's Horovod job
func (m *HorovodJobTrainer) IsSupported(name, ns string) bool {
	isHorovod := false

	if useCache {
		for _, job := range allJobs {
			if isHorovodJob(name, ns, job) {
				isHorovod = true
				log.Debugf("the job %s for %s in namespace %s is found.", job.Name, name, ns)
				break
			}
		}
	} else {
		jobList, err := m.client.BatchV1().Jobs(ns).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: fmt.Sprintf("release=%s", name),
		})
		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(jobList.Items) > 0 {
			isHorovod = true
		}
	}

	return isHorovod
}

func (m *HorovodJobTrainer) Type() string {
	return m.trainerType
}

func (m *HorovodJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	if useCache {
		tj, err = m.getTrainingJobFromCache(name, namespace)
	} else {
		tj, err = m.getTrainingJob(name, namespace)
	}

	return tj, err
}

func (m *HorovodJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		jobPod v1.Pod
		job    batchv1.Job
		latest metav1.Time
	)

	// 1. Get the batchJob of training Job
	pods := []v1.Pod{}
	jobList, err := m.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})
	if err != nil {
		return nil, err
	}

	if len(jobList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		job = jobList.Items[0]
	}

	// 2. Find the pod list, and determine the pod of the job
	podList, err := m.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, err
	}

	for _, item := range podList.Items {
		meta := item.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				log.Debugf("find job pod %v, break", item)
				break
			}
		}

		if !isJob {
			pods = append(pods, item)
			log.Debugf("add pod %v to pods", item)
		} else {
			if jobPod.Name == "" {
				latest = item.CreationTimestamp
				jobPod = item
				log.Debugf("set pod %s as first jobpod, and it's time is %v", jobPod.Name, jobPod.CreationTimestamp)
			} else {
				log.Debugf("current jobpod %s , and it's time is %v", jobPod.Name, latest)
				log.Debugf("candidate jobpod %s , and it's time is %v", item.Name, item.CreationTimestamp)
				current := item.CreationTimestamp
				if latest.Before(&current) {
					jobPod = item
					latest = current
					log.Debugf("replace")
				} else {
					log.Debugf("no replace")
				}
			}
		}
	}

	pods = append(pods, jobPod)

	return &HorovodJob{
		JobInfo: &JobInfo{
			BasicJobInfo: &BasicJobInfo{
				name:      name,
				resources: podResources(pods),
			},
			job:         job,
			jobPod:      jobPod,
			pods:        pods,
			trainerType: m.Type(),
		},
	}, nil

}

// Get the training job from Cache
func (m *HorovodJobTrainer) getTrainingJobFromCache(name, ns string) (TrainingJob, error) {

	var (
		jobPod v1.Pod
		job    batchv1.Job
		latest metav1.Time
	)

	pods := []v1.Pod{}

	// 1. Find the batch job
	for _, item := range allJobs {
		if isHorovodJob(name, ns, item) {
			job = item
			break
		}
	}

	// 2. Find the pods, and determine the pod of the job
	for _, item := range allPods {

		if !isHorovodPod(name, ns, item) {
			continue
		}

		meta := item.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				log.Debugf("find job pod %v, break", item)
				break
			}
		}

		if !isJob {
			// for non-job pod, add it into the pod list
			pods = append(pods, item)
			log.Debugf("add pod %v to pods", item)
		} else {
			if jobPod.Name == "" {
				latest = item.CreationTimestamp
				jobPod = item
				log.Debugf("set pod %s as first jobpod, and it's time is %v", jobPod.Name, jobPod.CreationTimestamp)
			} else {
				log.Debugf("current jobpod %s , and it's time is %v", jobPod.Name, latest)
				log.Debugf("candidate jobpod %s , and it's time is %v", item.Name, item.CreationTimestamp)
				current := item.CreationTimestamp
				if latest.Before(&current) {
					jobPod = item
					latest = current
					log.Debugf("replace")
				} else {
					log.Debugf("no replace")
				}
			}
		}
	}

	pods = append(pods, jobPod)

	return &HorovodJob{
		JobInfo: &JobInfo{
			BasicJobInfo: &BasicJobInfo{
				name:      name,
				resources: podResources(pods),
			},
			job:         job,
			jobPod:      jobPod,
			pods:        pods,
			trainerType: m.Type(),
		},
	}, nil
}

/**
* List Training jobs
 */
func (hj *HorovodJobTrainer) ListTrainingJobs() (jobs []TrainingJob, err error) {
	jobs = []TrainingJob{}
	jobInfos := []types.TrainingJobInfo{}
innerLoop:
	for _, horovodJob := range allJobs {
		jobInfo := types.TrainingJobInfo{}

		log.Debugf("find horovodJob %s in %s", horovodJob.Name, horovodJob.Namespace)
		if val, ok := horovodJob.Labels["release"]; ok && (horovodJob.Name == fmt.Sprintf("%s-tf-horovod-job", val)) {
			log.Debugf("the horovodJob %s with labels %s found in List", horovodJob.Name, val)
			jobInfo.Name = val
		} else {
			log.Debugf("the jobs %s with labels %s is not horovodJob in List", horovodJob.Name, val)
			continue innerLoop
		}

		jobInfo.Namespace = horovodJob.Namespace
		jobInfos = append(jobInfos, jobInfo)
		// jobInfos = append(jobInfos, types.TrainingJobInfo{Name: horovodJob.})
	}
	log.Debugf("jobInfos %v", jobInfos)

	for _, jobInfo := range jobInfos {
		job, err := hj.getTrainingJobFromCache(jobInfo.Name, jobInfo.Namespace)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func isHorovodJob(name, ns string, item batchv1.Job) bool {

	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the job %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "tf-horovod") {
		log.Debugf("the job %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func isHorovodPod(name, ns string, item v1.Pod) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the pod %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "tf-horovod") {
		log.Debugf("the pod %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}
