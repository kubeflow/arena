package cli

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestExtractGPURequested(t *testing.T) {
	tests := []struct {
		name     string
		obj      *unstructured.Unstructured
		expected int
	}{
		{
			name: "PyTorchJob with GPU",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"pytorchReplicaSpecs": map[string]interface{}{
							"Worker": map[string]interface{}{
								"replicas": int64(4),
								"template": map[string]interface{}{
									"spec": map[string]interface{}{
										"containers": []interface{}{
											map[string]interface{}{
												"resources": map[string]interface{}{
													"requests": map[string]interface{}{
														"nvidia.com/gpu": int64(8),
													},
												},
											},
										},
									},
								},
							},
							"Master": map[string]interface{}{
								"replicas": int64(1),
								"template": map[string]interface{}{
									"spec": map[string]interface{}{
										"containers": []interface{}{
											map[string]interface{}{
												"resources": map[string]interface{}{
													"requests": map[string]interface{}{
														"nvidia.com/gpu": int64(8),
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
			},
			expected: 40, // 4*8 + 1*8
		},
		{
			name: "PyTorchJob with GPU as string (K8s format)",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"pytorchReplicaSpecs": map[string]interface{}{
							"Worker": map[string]interface{}{
								"replicas": int64(2),
								"template": map[string]interface{}{
									"spec": map[string]interface{}{
										"containers": []interface{}{
											map[string]interface{}{
												"resources": map[string]interface{}{
													"requests": map[string]interface{}{
														"nvidia.com/gpu": "8",
													},
												},
											},
										},
									},
								},
							},
							"Master": map[string]interface{}{
								"replicas": int64(1),
								"template": map[string]interface{}{
									"spec": map[string]interface{}{
										"containers": []interface{}{
											map[string]interface{}{
												"resources": map[string]interface{}{
													"requests": map[string]interface{}{
														"nvidia.com/gpu": "8",
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
			},
			expected: 24, // 2*8 + 1*8
		},
		{
			name: "No GPU requested",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"pytorchReplicaSpecs": map[string]interface{}{
							"Worker": map[string]interface{}{
								"replicas": int64(2),
								"template": map[string]interface{}{
									"spec": map[string]interface{}{
										"containers": []interface{}{
											map[string]interface{}{
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
			},
			expected: 0,
		},
		{
			name:     "Empty spec",
			obj:      &unstructured.Unstructured{Object: map[string]interface{}{}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGPURequested(tt.obj)
			if result != tt.expected {
				t.Errorf("extractGPURequested() = %d, want %d", result, tt.expected)
			}
		})
	}
}
