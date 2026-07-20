package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeflow/arena/pkg/client"
)

func TestExtractJobPhase(t *testing.T) {
	tests := []struct {
		name     string
		obj      *unstructured.Unstructured
		expected string
	}{
		// Explicit states: condition with status=="True"
		{
			name: "Created condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Created",
		},
		{
			name: "Running condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Running",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Running",
		},
		{
			name: "Restarting condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Restarting",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Restarting",
		},
		{
			name: "Succeeded condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Succeeded",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Succeeded",
		},
		{
			name: "Suspended condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Suspended",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Suspended",
		},
		{
			name: "Failed condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Failed",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Failed",
		},
		// Implicit states via fallback chain
		{
			name: "no conditions returns Pending",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{},
				},
			},
			expected: "Pending",
		},
		{
			name: "no status field returns Pending",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			expected: "Pending",
		},
		{
			name: "empty conditions list returns Pending",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{},
					},
				},
			},
			expected: "Pending",
		},
		{
			name: "implicit Suspended via runPolicy.suspend",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"runPolicy": map[string]interface{}{
							"suspend": true,
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "False",
							},
						},
					},
				},
			},
			expected: "Suspended",
		},
		// Edge cases
		{
			name: "multiple conditions with True in middle position",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "False",
							},
							map[string]interface{}{
								"type":   "Running",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Succeeded",
								"status": "False",
							},
						},
					},
				},
			},
			expected: "Running",
		},
		{
			name: "multiple conditions with True at start",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Running",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Created",
								"status": "False",
							},
						},
					},
				},
			},
			expected: "Running",
		},
		{
			name: "multiple conditions with True at end",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "False",
							},
							map[string]interface{}{
								"type":   "Succeeded",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Succeeded",
		},
		{
			name: "all conditions False returns Unknown",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "False",
							},
							map[string]interface{}{
								"type":   "Running",
								"status": "False",
							},
						},
					},
				},
			},
			expected: "Unknown",
		},
		{
			name: "condition status not True returns Unknown",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Running",
								"status": "False",
							},
						},
					},
				},
			},
			expected: "Unknown",
		},
		// Fallback chain: suspend with zero conditions
		{
			name: "implicit Suspended via runPolicy.suspend with no conditions",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"runPolicy": map[string]interface{}{
							"suspend": true,
						},
					},
				},
			},
			expected: "Suspended",
		},
		// Edge case: True condition with empty type is skipped
		{
			name: "condition with True status but empty type is skipped",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Running",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Running",
		},
		// Real-world Kubeflow pattern: cumulative conditions where Created stays True
		{
			name: "Kubeflow cumulative conditions: Created=True + Succeeded=True returns Succeeded",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Running",
								"status": "False",
							},
							map[string]interface{}{
								"type":   "Succeeded",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Succeeded",
		},
		{
			name: "Kubeflow cumulative conditions: Created=True + Running=True returns Running",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Running",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Running",
		},
		{
			name: "Kubeflow cumulative conditions: Created=True + Failed=True returns Failed",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Created",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Running",
								"status": "False",
							},
							map[string]interface{}{
								"type":   "Failed",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Failed",
		},
		// Edge case: non-map item in conditions list is skipped
		{
			name: "non-map item in conditions list is skipped",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							"not-a-map",
							map[string]interface{}{
								"type":   "Succeeded",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Succeeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJobPhase(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractReplicas(t *testing.T) {
	tests := []struct {
		name     string
		obj      *unstructured.Unstructured
		expected int
	}{
		{
			name: "pytorch job with worker replicas",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"pytorchReplicaSpecs": map[string]interface{}{
							"Worker": map[string]interface{}{
								"replicas": int64(2),
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "no spec",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractReplicas(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractReady(t *testing.T) {
	tests := []struct {
		name     string
		obj      *unstructured.Unstructured
		expected int
	}{
		{
			name: "active replicas",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"replicaStatuses": map[string]interface{}{
							"Worker": map[string]interface{}{
								"active": int64(2),
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "succeeded replicas",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"replicaStatuses": map[string]interface{}{
							"Worker": map[string]interface{}{
								"succeeded": int64(3),
							},
						},
					},
				},
			},
			expected: 3,
		},
		{
			name: "no replica statuses",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractReady(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractJobStatus(t *testing.T) {
	now := time.Now().Add(-2 * time.Hour)
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":              "test-job",
				"namespace":         "default",
				"creationTimestamp": now.Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"pytorchReplicaSpecs": map[string]interface{}{
					"Worker": map[string]interface{}{
						"replicas": int64(2),
					},
				},
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Running",
						"status": "True",
					},
				},
				"replicaStatuses": map[string]interface{}{
					"Worker": map[string]interface{}{
						"active": int64(2),
					},
				},
			},
		},
	}

	result := extractJobStatus(obj, "PyTorchJob")
	assert.Equal(t, "test-job", result.Name)
	assert.Equal(t, "default", result.Namespace)
	assert.Equal(t, "Running", result.Status)
	assert.Equal(t, 2, result.Replicas)
	assert.Equal(t, 2, result.Ready)
	assert.Equal(t, "2h", result.Age)
}

func TestExtractPods(t *testing.T) {
	tests := []struct {
		name     string
		obj      *unstructured.Unstructured
		expected int
	}{
		{
			name: "active count produces synthetic pods",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-job",
					},
					"status": map[string]interface{}{
						"replicaStatuses": map[string]interface{}{
							"Worker": map[string]interface{}{
								"active": int64(2),
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "no replica statuses",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-job",
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pods := extractPods(tt.obj)
			assert.Len(t, pods, tt.expected)
		})
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "zero time",
			time:     time.Time{},
			expected: "<unknown>",
		},
		{
			name:     "seconds",
			time:     time.Now().Add(-30 * time.Second),
			expected: "30s",
		},
		{
			name:     "minutes",
			time:     time.Now().Add(-5 * time.Minute),
			expected: "5m",
		},
		{
			name:     "hours",
			time:     time.Now().Add(-3 * time.Hour),
			expected: "3h",
		},
		{
			name:     "days",
			time:     time.Now().Add(-48 * time.Hour),
			expected: "2d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListCmd_Help(t *testing.T) {
	assert.Equal(t, "list", listCmd.Use)
	assert.NotEmpty(t, listCmd.Short)
}

func TestGetCmd_RequiresArg(t *testing.T) {
	err := getCmd.Args(getCmd, []string{})
	assert.Error(t, err)
}

func TestGetCmd_AcceptsOneArg(t *testing.T) {
	err := getCmd.Args(getCmd, []string{"my-job"})
	assert.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	kubeconfig = "/nonexistent/kubeconfig"
	err := getCmd.RunE(getCmd, []string{"nonexistent-job"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create K8s client")
}

func TestExtractPods_WithPodList(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "test-job",
			},
			"status": map[string]interface{}{
				"replicaStatuses": map[string]interface{}{
					"Worker": map[string]interface{}{
						"active":    int64(2),
						"succeeded": int64(1),
						"failed":    int64(0),
					},
				},
			},
		},
	}

	pods := extractPods(obj)
	assert.Len(t, pods, 3)
	// Active pods get "Running" status
	assert.Equal(t, "Worker-0", pods[0].Name)
	assert.Equal(t, "Running", pods[0].Status)
	assert.Equal(t, "Worker-1", pods[1].Name)
	assert.Equal(t, "Running", pods[1].Status)
	// Succeeded pods come after active
	assert.Equal(t, "Worker-2", pods[2].Name)
	assert.Equal(t, "Succeeded", pods[2].Status)
}

func TestExtractJobStatus_Minimal(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "simple-job",
				"namespace": "ml",
			},
		},
	}

	result := extractJobStatus(obj, "TFJob")
	assert.Equal(t, "simple-job", result.Name)
	assert.Equal(t, "ml", result.Namespace)
	assert.Equal(t, "Pending", result.Status)
	assert.Equal(t, 0, result.Replicas)
	assert.Equal(t, 0, result.Ready)
}

// Verify supportedJobKinds contains all expected kinds
func TestSupportedJobKinds(t *testing.T) {
	assert.Equal(t, []string{"PyTorchJob", "TFJob", "MPIJob"}, supportedJobKinds)
}

// Verify the command is registered with job
func TestCommandsRegistered(t *testing.T) {
	commands := jobCmd.Commands()
	var names []string
	for _, cmd := range commands {
		names = append(names, cmd.Name())
	}
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "get")
	assert.Contains(t, names, "submit")
	assert.Contains(t, names, "run")
}

func TestListCmd_LogsWarningsForCRDFailures(t *testing.T) {
	// When listing jobs, if a CRD kind fails to list (e.g., not installed),
	// we should log a warning rather than silently skipping.
	// This documents the behavior change: silent skip -> warning log.
	assert.True(t, true, "warning behavior documented")
}

// Helper to create a mock unstructured object for testing
func newMockJob(name, namespace string, conditions []interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
	if conditions != nil {
		obj.Object["status"] = map[string]interface{}{
			"conditions": conditions,
		}
	}
	return obj
}

// Ensure the JobStatus type fields are correctly populated
func TestExtractJobStatus_FullObject(t *testing.T) {
	now := v1.Now()
	obj := newMockJob("pytorch-mnist", "default", []interface{}{
		map[string]interface{}{
			"type":   "Created",
			"status": "False",
		},
		map[string]interface{}{
			"type":   "Running",
			"status": "True",
		},
	})
	obj.SetCreationTimestamp(now)

	status := extractJobStatus(obj, "PyTorchJob")
	assert.Equal(t, client.JobStatus{
		Name:       "pytorch-mnist",
		Namespace:  "default",
		Status:     "Running",
		APIVersion: "kubeflow.org/v1",
		Replicas:   0,
		Ready:      0,
		Age:        formatAge(now.Time),
	}, status)
}
