package commands

import (
	"fmt"
	"strconv"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_handleSharedGPUsIfNeeded(t *testing.T) {
	interactiveTrue := true
	interactiveFalse := false
	elasticJobTrue := true
	elasticJobFalse := false
	fractionalGPU := 0.2
	wholeGPU := float64(1)
	namespace := "default"

	type args struct {
		jobName    string
		submitArgs submitRunaiJobArgs
	}
	tests := []struct {
		name                      string
		args                      args
		wantErr                   bool
		shouldRunFractionalGPUJob bool
	}{
		{
			name: "Valid fractional GPU job",
			args: args{
				jobName: "job1",
				submitArgs: submitRunaiJobArgs{
					Interactive: &interactiveTrue,
					GPU:         &fractionalGPU,
					Elastic:     &elasticJobFalse,
				},
			},
			wantErr:                   false,
			shouldRunFractionalGPUJob: true,
		},
		{
			name: "Valid whole GPU job",
			args: args{
				jobName: "job2",
				submitArgs: submitRunaiJobArgs{
					Interactive: &interactiveTrue,
					GPU:         &wholeGPU,
					Elastic:     &elasticJobFalse,
				},
			},
			wantErr: false,
		},
		{
			name: "Non interactive fractional GPU job",
			args: args{
				jobName: "job3",
				submitArgs: submitRunaiJobArgs{
					Interactive: &interactiveFalse,
					GPU:         &fractionalGPU,
					Elastic:     &elasticJobFalse,
				},
			},
			wantErr: true,
		},
		{
			name: "Elastic fractional GPU job",
			args: args{
				jobName: "job4",
				submitArgs: submitRunaiJobArgs{
					Interactive: &interactiveTrue,
					GPU:         &fractionalGPU,
					Elastic:     &elasticJobTrue,
				},
			},
			wantErr: true,
		},
	}

	clientMock := fake.NewSimpleClientset()
	configMapMock := clientMock.CoreV1().ConfigMaps("default")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSubmitArgs := tt.args.submitArgs
			if err := handleSharedGPUsIfNeeded(clientMock, tt.args.jobName, namespace, &testSubmitArgs); (err != nil) != tt.wantErr {
				t.Errorf("Miss amtch between error and expected error error = %v, wantErr %v", err, tt.wantErr)
			}

			gpuFraction, err := strconv.ParseFloat(testSubmitArgs.GPUFraction, 64)
			if err != nil && tt.shouldRunFractionalGPUJob {
				t.Errorf("handleSharedGPUsIfNeeded() failed to parse gpuFraction %v, while expecting it to manage", err)
			}

			if gpuFraction != *testSubmitArgs.GPU && tt.shouldRunFractionalGPUJob {
				t.Errorf("gpuFraction: %v, *testSubmitArgs.GPU: %v, miss match", gpuFraction, *testSubmitArgs.GPU)
			}

			gpuFractionFixed, err := strconv.ParseFloat(testSubmitArgs.GPUFractionFixed, 64)
			if err != nil && tt.shouldRunFractionalGPUJob {
				t.Errorf("handleSharedGPUsIfNeeded() failed to parse gpuFraction %v, while expecting it to manage", err)
			}

			expectingFixedGPU := *testSubmitArgs.GPU * factorForGPUFraction
			if gpuFractionFixed != expectingFixedGPU && tt.shouldRunFractionalGPUJob {
				t.Errorf("gpuFractionFixed: %v, expectingFixedGPU: %v, miss match", gpuFractionFixed, expectingFixedGPU)
			}

			configMapName := fmt.Sprintf("%v-%v", tt.args.jobName, runaiFractionGPUSuffix)
			configMapResult, _ := configMapMock.Get(configMapName, metav1.GetOptions{})

			// Not expecting configmap and configmap really doesn't exist
			if configMapResult == nil && !tt.shouldRunFractionalGPUJob {
				return
			}

			if configMapResult != nil && tt.shouldRunFractionalGPUJob {
				if _, found := configMapResult.Data[runaiVisibleDevices]; !found {
					t.Errorf("Failed to find runaiVisibleDevices in configMap")
				}
			} else {
				t.Errorf("Failed to get configmap for job, or wasn't expecting one, tt.shouldRunFractionalGPUJob: %v, configMapResult: %v", tt.shouldRunFractionalGPUJob, configMapResult)
			}
		})
	}
}
