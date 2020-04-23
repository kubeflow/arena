package commands

import (
	"strconv"
	"testing"
)

type testArgs struct {
	interactive *bool
	elastic     *bool
	gpu         *float64
}

func TestGPUSharingManager(t *testing.T) {
	interactiveTrue := true
	interactiveFalse := false
	elasticJobTrue := true
	elasticJobFalse := false
	fractionalGPU := 0.2
	wholeGPU := float64(1)

	tests := []struct {
		name                      string
		wantErr                   bool
		shouldRunFractionalGPUJob bool
		args                      *testArgs
	}{
		{
			name: "Valid fractional GPU job",
			args: &testArgs{
				interactive: &interactiveTrue,
				gpu:         &fractionalGPU,
				elastic:     &elasticJobFalse,
			},
			wantErr:                   false,
			shouldRunFractionalGPUJob: true,
		},
		{
			name: "Valid whole GPU job",
			args: &testArgs{
				interactive: &interactiveTrue,
				gpu:         &wholeGPU,
				elastic:     &elasticJobFalse,
			},
			wantErr: false,
		},
		{
			name: "Non interactive fractional GPU job",
			args: &testArgs{
				interactive: &interactiveFalse,
				gpu:         &fractionalGPU,
				elastic:     &elasticJobFalse,
			},
			wantErr: true,
		},
		{
			name: "Elastic fractional GPU job",
			args: &testArgs{
				interactive: &interactiveTrue,
				gpu:         &fractionalGPU,
				elastic:     &elasticJobTrue,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submitArgs := setSubmitArgs(tt.args)
			testSubmitArgs := *submitArgs
			if err := handleRequestedGPUs(&testSubmitArgs); (err != nil) != tt.wantErr {
				t.Errorf("Miss match between error and expected error, error: %v, wantErr: %v", err, tt.wantErr)
			}

			gpuFraction, err := strconv.ParseFloat(testSubmitArgs.GPUFraction, 64)
			if err != nil {
				if tt.shouldRunFractionalGPUJob {
					t.Errorf("handleSharedGPUsIfNeeded() failed to parse gpuFraction %v, while expecting it to manage", err)
				} else if !tt.wantErr && float64(*testSubmitArgs.GPUInt) != *submitArgs.GPU {
					t.Errorf("GPUInt: %v, submitArgs.gpu: %v", *testSubmitArgs.GPUInt, *submitArgs.GPU)
				}
			}

			if gpuFraction != *submitArgs.GPU && tt.shouldRunFractionalGPUJob {
				t.Errorf("gpuFraction: %v, *testSubmitArgs.gpu: %v, miss match", gpuFraction, *testSubmitArgs.GPU)
			}
		})
	}
}

func setSubmitArgs(args *testArgs) *submitRunaiJobArgs {
	submitArgs := submitRunaiJobArgs{}
	submitArgs.GPU = args.gpu
	submitArgs.Interactive = args.interactive
	submitArgs.Elastic = args.elastic
	return &submitArgs
}
