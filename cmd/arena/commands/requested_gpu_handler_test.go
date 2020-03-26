package commands

import (
	"strconv"
	"testing"
)

func TestGPUSharingManager(t *testing.T) {
	interactiveTrue := true
	interactiveFalse := false
	elasticJobTrue := true
	elasticJobFalse := false
	fractionalGPU := 0.2
	wholeGPU := float64(1)

	type args struct {
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
				submitArgs: submitRunaiJobArgs{
					Interactive: &interactiveTrue,
					GPU:         &fractionalGPU,
					Elastic:     &elasticJobTrue,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSubmitArgs := tt.args.submitArgs
			if err := handleRequestedGPUs(&testSubmitArgs); (err != nil) != tt.wantErr {
				t.Errorf("Miss match between error and expected error, error: %v, wantErr: %v", err, tt.wantErr)
			}

			gpuFraction, err := strconv.ParseFloat(testSubmitArgs.GPUFraction, 64)
			if err != nil {
				if tt.shouldRunFractionalGPUJob {
					t.Errorf("handleSharedGPUsIfNeeded() failed to parse gpuFraction %v, while expecting it to manage", err)
				} else if !tt.wantErr && float64(*testSubmitArgs.GPUInt) != *tt.args.submitArgs.GPU {
					t.Errorf("GPUInt: %v, tt.args.submitArgs.GPU: %v", *testSubmitArgs.GPUInt, *tt.args.submitArgs.GPU)
				}
			}

			if gpuFraction != *tt.args.submitArgs.GPU && tt.shouldRunFractionalGPUJob {
				t.Errorf("gpuFraction: %v, *testSubmitArgs.GPU: %v, miss match", gpuFraction, *testSubmitArgs.GPU)
			}
		})
	}
}
