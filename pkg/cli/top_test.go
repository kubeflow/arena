package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestTopJobCommand_InvalidFormat(t *testing.T) {
	// Save and restore the output format variable
	origFormat := topOutputFormat
	defer func() { topOutputFormat = origFormat }()

	// Set invalid format directly (bypassing flag parsing)
	topOutputFormat = "invalid"

	// Call RunE directly to test format validation
	err := topJobCmd.RunE(topJobCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Contains(t, err.Error(), "invalid")
}

func TestTopCmd_Help(t *testing.T) {
	assert.Equal(t, "top", topCmd.Use)
	assert.NotEmpty(t, topCmd.Short)
}

func TestTopJobCmd_Registered(t *testing.T) {
	commands := topCmd.Commands()
	var names []string
	for _, cmd := range commands {
		names = append(names, cmd.Name())
	}
	assert.Contains(t, names, "job")
}

func TestTopCmd_RegisteredOnRoot(t *testing.T) {
	commands := rootCmd.Commands()
	var names []string
	for _, cmd := range commands {
		names = append(names, cmd.Name())
	}
	assert.Contains(t, names, "top")
}

func TestExtractGPURequested_SingleRole(t *testing.T) {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
			"spec": map[string]interface{}{
				"pytorchReplicaSpecs": map[string]interface{}{
					"Worker": map[string]interface{}{
						"replicas": int64(2),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "pytorch",
										"image": "pytorch:1.13",
										"resources": map[string]interface{}{
											"requests": map[string]interface{}{
												"nvidia.com/gpu": "2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	gpu := extractGPURequested(job)
	assert.Equal(t, 4, gpu, "2 workers x 2 GPUs each = 4 total")
}

func TestExtractGPURequested_MultipleRoles(t *testing.T) {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "TFJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
			"spec": map[string]interface{}{
				"tfReplicaSpecs": map[string]interface{}{
					"Chief": map[string]interface{}{
						"replicas": int64(1),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "tensorflow",
										"image": "tf:2.15",
										"resources": map[string]interface{}{
											"requests": map[string]interface{}{
												"nvidia.com/gpu": "1",
											},
										},
									},
								},
							},
						},
					},
					"Worker": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "tensorflow",
										"image": "tf:2.15",
										"resources": map[string]interface{}{
											"requests": map[string]interface{}{
												"nvidia.com/gpu": "2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	gpu := extractGPURequested(job)
	assert.Equal(t, 7, gpu, "1 chief x 1 GPU + 3 workers x 2 GPUs = 7 total")
}

func TestExtractGPURequested_NoGPUs(t *testing.T) {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
			"spec": map[string]interface{}{
				"pytorchReplicaSpecs": map[string]interface{}{
					"Worker": map[string]interface{}{
						"replicas": int64(1),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "pytorch",
										"image": "pytorch:1.13",
										"resources": map[string]interface{}{
											"requests": map[string]interface{}{
												"cpu": "4",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	gpu := extractGPURequested(job)
	assert.Equal(t, 0, gpu, "CPU-only job should report 0 GPUs")
}

func TestExtractGPURequested_EmptySpec(t *testing.T) {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
		},
	}
	gpu := extractGPURequested(job)
	assert.Equal(t, 0, gpu)
}

func TestTopJobCmd_FrameworkLabelFromCRD(t *testing.T) {
	// Verify that when a CRD has the arena.io/framework label, it takes
	// precedence over kindToFramework mapping.
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v2beta1",
			"kind":       "MPIJob",
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": "default",
				"labels": map[string]interface{}{
					FrameworkLabel: "deepspeed",
				},
			},
		},
	}
	labels := job.GetLabels()
	fw, ok := labels[FrameworkLabel]
	assert.True(t, ok)
	assert.Equal(t, "deepspeed", fw)
}

func TestExtractJobPhase_Suspended(t *testing.T) {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
			"spec": map[string]interface{}{
				"runPolicy": map[string]interface{}{
					"suspend": true,
				},
			},
		},
	}
	phase := extractJobPhase(job)
	assert.Equal(t, "Suspended", phase)
}

func TestExtractJobPhase_Pending(t *testing.T) {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
		},
	}
	phase := extractJobPhase(job)
	assert.Equal(t, "Pending", phase)
}

func TestExtractJobPhase_ReverseScan(t *testing.T) {
	// Kubeflow appends conditions chronologically. A completed job has
	// Created=True followed by Succeeded=True. The reverse scan should
	// return "Succeeded", not "Created".
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata":   map[string]interface{}{"name": "test", "namespace": "default"},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{"type": "Created", "status": "True"},
					map[string]interface{}{"type": "Running", "status": "False"},
					map[string]interface{}{"type": "Succeeded", "status": "True"},
				},
			},
		},
	}
	phase := extractJobPhase(job)
	assert.Equal(t, "Succeeded", phase)
}

func TestExtractJobStatus_Full(t *testing.T) {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":              "my-job",
				"namespace":         "default",
				"creationTimestamp": metav1.NewTime(time.Now().Add(-2 * time.Hour)).Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"pytorchReplicaSpecs": map[string]interface{}{
					"Worker": map[string]interface{}{
						"replicas": int64(3),
					},
				},
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{"type": "Running", "status": "True"},
				},
				"replicaStatuses": map[string]interface{}{
					"Worker": map[string]interface{}{
						"active": int64(2),
					},
				},
			},
		},
	}
	status := extractJobStatus(job, "PyTorchJob")
	assert.Equal(t, "my-job", status.Name)
	assert.Equal(t, "default", status.Namespace)
	assert.Equal(t, "Running", status.Status)
	assert.Equal(t, 3, status.Replicas)
	assert.Equal(t, 2, status.Ready)
}

func TestFormatAge_Units(t *testing.T) {
	tests := []struct {
		name     string
		created  time.Time
		contains string
	}{
		{"zero time", time.Time{}, "<unknown>"},
		{"seconds ago", time.Now().Add(-30 * time.Second), "s"},
		{"minutes ago", time.Now().Add(-15 * time.Minute), "m"},
		{"hours ago", time.Now().Add(-3 * time.Hour), "h"},
		{"days ago", time.Now().Add(-48 * time.Hour), "d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.created)
			assert.Contains(t, got, tt.contains)
		})
	}
}
