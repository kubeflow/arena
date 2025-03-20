//
// Copyright 2025 The Kubeflow Authors.
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
//

package training

import (
	"testing"

	commonv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestGetStatus(t *testing.T) {
	testcases := []struct {
		status   commonv1.JobStatus
		expected string
	}{
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "PENDING",
		},
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobRunning,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "RUNNING",
		},
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobRunning,
						Status: corev1.ConditionFalse,
					},
					{
						Type:   commonv1.JobSucceeded,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "SUCCEEDED",
		},
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobRunning,
						Status: corev1.ConditionFalse,
					},
					{
						Type:   commonv1.JobFailed,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "FAILED",
		},
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobConditionType("Queueing"),
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "PENDING",
		},
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobConditionType("Queueing"),
						Status: corev1.ConditionFalse,
					},
					{
						Type:   commonv1.JobRunning,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "RUNNING",
		},
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobConditionType("Queueing"),
						Status: corev1.ConditionFalse,
					},
					{
						Type:   commonv1.JobRunning,
						Status: corev1.ConditionFalse,
					},
					{
						Type:   commonv1.JobSucceeded,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "SUCCEEDED",
		},
		{
			status: commonv1.JobStatus{
				Conditions: []commonv1.JobCondition{
					{
						Type:   commonv1.JobCreated,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobConditionType("Queueing"),
						Status: corev1.ConditionFalse,
					},
					{
						Type:   commonv1.JobRunning,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   commonv1.JobFailed,
						Status: corev1.ConditionTrue,
					},
				},
			},
			expected: "FAILED",
		},
	}

	for _, tc := range testcases {
		actual := getStatus(tc.status)
		if actual != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, actual)
		}

	}
}

func TestHasCondition(t *testing.T) {
	conditions := []commonv1.JobCondition{
		{
			Type:   commonv1.JobCreated,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   commonv1.JobRunning,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   commonv1.JobSucceeded,
			Status: corev1.ConditionTrue,
		},
	}

	testcases := []struct {
		status   commonv1.JobStatus
		condType commonv1.JobConditionType
		expected bool
	}{
		{
			status: commonv1.JobStatus{
				Conditions: conditions,
			},
			condType: commonv1.JobCreated,
			expected: true,
		},
		{
			status: commonv1.JobStatus{
				Conditions: conditions,
			},
			condType: commonv1.JobRunning,
			expected: true,
		},
		{
			status: commonv1.JobStatus{
				Conditions: conditions,
			},
			condType: commonv1.JobSucceeded,
			expected: true,
		},
		{
			status: commonv1.JobStatus{
				Conditions: conditions,
			},
			condType: commonv1.JobRestarting,
			expected: false,
		},
		{
			status: commonv1.JobStatus{
				Conditions: conditions,
			},
			condType: commonv1.JobFailed,
			expected: false,
		},
	}

	for _, tc := range testcases {
		actual := hasCondition(tc.status, tc.condType)
		if actual != tc.expected {
			t.Errorf("expected: %v, got: %v", tc.expected, actual)
		}
	}
}
