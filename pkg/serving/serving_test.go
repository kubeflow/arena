// Copyright 2026 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serving

import (
	"testing"

	"github.com/kubeflow/arena/pkg/apis/types"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	lws_v1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

func TestInt32PtrVal(t *testing.T) {
	if val := int32PtrVal(nil, 1); val != 1 {
		t.Errorf("expected 1, got %d", val)
	}
	var v int32 = 5
	if val := int32PtrVal(&v, 1); val != 5 {
		t.Errorf("expected 5, got %d", val)
	}
}

func TestServingJobNilDeployment(t *testing.T) {
	job := &servingJob{
		name:        "test-job",
		namespace:   "default",
		servingType: types.CustomServingJob,
		version:     "v1",
		deployment:  nil,
	}

	if job.Uid() != "" {
		t.Errorf("expected empty string for Uid when deployment is nil, got %s", job.Uid())
	}
	if job.Age() != 0 {
		t.Errorf("expected 0 for Age when deployment is nil, got %v", job.Age())
	}
	if job.StartTime() == nil {
		t.Errorf("expected non-nil StartTime when deployment is nil")
	}
	if job.RequestCPUs() != 0 {
		t.Errorf("expected 0 for RequestCPUs when deployment is nil, got %f", job.RequestCPUs())
	}
	if job.RequestGPUs() != 0 {
		t.Errorf("expected 0 for RequestGPUs when deployment is nil, got %f", job.RequestGPUs())
	}
	if job.RequestGPUMemory() != 0 {
		t.Errorf("expected 0 for RequestGPUMemory when deployment is nil, got %d", job.RequestGPUMemory())
	}
	if job.RequestGPUCore() != 0 {
		t.Errorf("expected 0 for RequestGPUCore when deployment is nil, got %d", job.RequestGPUCore())
	}
	if job.AvailableInstances() != 0 {
		t.Errorf("expected 0 for AvailableInstances when deployment is nil, got %d", job.AvailableInstances())
	}
	if job.DesiredInstances() != 0 {
		t.Errorf("expected 0 for DesiredInstances when deployment is nil, got %d", job.DesiredInstances())
	}
	labels := job.GetLabels()
	if labels == nil || len(labels) != 0 {
		t.Errorf("expected empty map for GetLabels when deployment is nil, got %v", labels)
	}
}

func TestServingJobNilReplicas(t *testing.T) {
	job := &servingJob{
		name:        "test-job",
		namespace:   "default",
		servingType: types.CustomServingJob,
		version:     "v1",
		deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job-v1",
				Namespace: "default",
				Labels: map[string]string{
					servingNameLabelKey: "test-job",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: nil, // nil Replicas pointer
			},
		},
	}

	if job.GetLabels()[servingNameLabelKey] != "test-job" {
		t.Errorf("expected label servingName to be test-job, got %v", job.GetLabels())
	}
	// Verify calculations don't panic and default replicas to 1
	if job.RequestCPUs() != 0 {
		t.Errorf("expected 0 CPUs for container without CPU limits, got %f", job.RequestCPUs())
	}
}

func TestLwsJobNilLws(t *testing.T) {
	job := &lwsJob{
		lws: nil,
		servingJob: &servingJob{
			name:        "lws-test",
			namespace:   "default",
			servingType: types.DistributedServingJob,
			version:     "v1",
			deployment:  nil,
		},
	}

	if job.GetLabels() == nil || len(job.GetLabels()) != 0 {
		t.Errorf("expected empty map for GetLabels when lws is nil, got %v", job.GetLabels())
	}
	if job.Uid() != "" {
		t.Errorf("expected empty string for Uid when lws is nil, got %s", job.Uid())
	}
	if job.Age() != 0 {
		t.Errorf("expected 0 for Age when lws is nil, got %v", job.Age())
	}
	if job.RequestCPUs() != 0 {
		t.Errorf("expected 0 for RequestCPUs when lws is nil, got %f", job.RequestCPUs())
	}
	if job.RequestGPUs() != 0 {
		t.Errorf("expected 0 for RequestGPUs when lws is nil, got %f", job.RequestGPUs())
	}
	if job.DesiredInstances() != 0 {
		t.Errorf("expected 0 for DesiredInstances when lws is nil, got %d", job.DesiredInstances())
	}
	if job.AvailableInstances() != 0 {
		t.Errorf("expected 0 for AvailableInstances when lws is nil, got %d", job.AvailableInstances())
	}
}

func TestLwsJobValidLws(t *testing.T) {
	lws := &lws_v1.LeaderWorkerSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lws-job",
			Namespace: "default",
			UID:       "uid-1234",
			Labels: map[string]string{
				servingNameLabelKey: "lws-job",
			},
		},
		Spec: lws_v1.LeaderWorkerSetSpec{
			Replicas: nil, // nil replicas
		},
		Status: lws_v1.LeaderWorkerSetStatus{
			Replicas:      2,
			ReadyReplicas: 1,
		},
	}

	job := &lwsJob{
		lws: lws,
		servingJob: &servingJob{
			name:        "lws-job",
			namespace:   "default",
			servingType: types.DistributedServingJob,
			version:     "v1",
			deployment:  nil,
		},
	}

	if job.Uid() != "uid-1234" {
		t.Errorf("expected Uid uid-1234, got %s", job.Uid())
	}
	if job.GetLabels()[servingNameLabelKey] != "lws-job" {
		t.Errorf("expected label servingName to be lws-job, got %v", job.GetLabels())
	}
	if job.DesiredInstances() != 2 {
		t.Errorf("expected DesiredInstances 2, got %d", job.DesiredInstances())
	}
	if job.AvailableInstances() != 1 {
		t.Errorf("expected AvailableInstances 1, got %d", job.AvailableInstances())
	}
}
