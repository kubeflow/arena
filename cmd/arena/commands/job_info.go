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
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"time"
)

type JobInfo struct {
	*BasicJobInfo
	job          batchv1.Job
	pods         []v1.Pod // all the pods including statefulset and job
	jobPod       v1.Pod   // the pod of job
	gpuCount     int64
	requestedGPU int64
	allocatedGPU int64
	trainerType  string // return trainer type: MPI, STANDALONE, TENSORFLOW
}

func (ji *JobInfo) Name() string {
	return ji.name
}

func (ji *JobInfo) Uid() string {
	return string(ji.job.UID)
}

func (ji *JobInfo) Trainer() string {
	return ji.trainerType
}

// Get the chief Pod of the Job.
func (ji *JobInfo) ChiefPod() v1.Pod {
	return ji.jobPod
}

// Get all the pods of the Training Job
func (ji *JobInfo) AllPods() []v1.Pod {
	return ji.pods
}

// Get the hostIP of the chief Pod
func (ji *JobInfo) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if ji.GetStatus() == "RUNNING" {
		hostIP = ji.jobPod.Status.HostIP
	}

	return hostIP
}

// Requested GPU count of the Job
func (ji *JobInfo) RequestedGPU() int64 {
	if ji.requestedGPU > 0 {
		return ji.requestedGPU
	}
	for _, pod := range ji.pods {
		ji.requestedGPU += gpuInPod(pod)
	}
	return ji.requestedGPU
}

// Requested GPU count of the Job
func (ji *JobInfo) AllocatedGPU() int64 {
	if ji.allocatedGPU > 0 {
		return ji.allocatedGPU
	}
	for _, pod := range ji.pods {
		ji.allocatedGPU += gpuInActivePod(pod)
	}
	return ji.allocatedGPU
}

func (ji *JobInfo) Age() time.Duration {
	job := ji.job
	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}
	return metav1.Now().Sub(job.Status.StartTime.Time)
}

// Get the Job Training Duration
func (ji *JobInfo) Duration() time.Duration {
	job := ji.job

	if job.Status.StartTime == nil ||
		job.Status.StartTime.IsZero() {
		return 0
	}
	if job.Status.CompletionTime != nil {
		return job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time)
	}

	if ji.GetStatus() == "FAILED" {
		cond := getPodLatestCondition(ji.ChiefPod())
		if !cond.LastTransitionTime.IsZero() {
			return cond.LastTransitionTime.Time.Sub(job.Status.StartTime.Time)
		} else {
			log.Debugf("the latest condition's time is zero of pod %s", ji.ChiefPod().Name)
		}
	}

	return metav1.Now().Sub(job.Status.StartTime.Time)
}

func (ji *JobInfo) StartTime() *metav1.Time {
	return ji.job.Status.StartTime
}

// Get the Status of the Job: RUNNING, PENDING, SUCCEEDED, FAILED
func (ji *JobInfo) GetStatus() (status string) {
	job := ji.job
	pod := ji.jobPod
	if job.Status.Active > 0 {
		status = "RUNNING"
	} else if job.Status.Succeeded > 0 {
		status = "SUCCEEDED"
	} else if job.Status.Failed > 0 {
		status = "FAILED"
	}

	if status == "RUNNING" {
		hostIP := pod.Status.HostIP
		if hostIP == "" {
			status = "PENDING"
		} else if pod.Status.Phase == v1.PodPending {
			status = "PENDING"
		}
	}
	return status
}

func (ji *JobInfo) Namespace() string {
	return ji.job.Namespace
}
