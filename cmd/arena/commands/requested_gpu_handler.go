package commands

import (
	"fmt"
)

const (
	runaiGPUFraction = "gpu-fraction"
	runaiGPUIndex    = "runai-gpu"
)

func handleRequestedGPUs(submitArgs *submitRunaiJobArgs) error {
	if submitArgs.GPU == nil {
		return nil
	}

	if float64(int(*submitArgs.GPU)) == *submitArgs.GPU {
		gpu := int(*submitArgs.GPU)
		submitArgs.GPUInt = &gpu

		return nil
	}

	err := validateFractionalGPUTask(submitArgs)
	if err != nil {
		return err
	}

	submitArgs.GPUFraction = fmt.Sprintf("%v", *submitArgs.GPU)
	return nil
}

func validateFractionalGPUTask(submitArgs *submitRunaiJobArgs) error {
	if submitArgs.Interactive == nil || *submitArgs.Interactive == false {
		return fmt.Errorf("Jobs that require a fractional number of GPUs must be interactive. Run the job with flag '--interactive'")
	}

	if submitArgs.Elastic != nil && *submitArgs.Elastic == true {
		return fmt.Errorf("Jobs that require a fractional number of GPUs can't be elastic jobs. Run the job without flag '--elastic'")
	}

	if *submitArgs.GPU > 1 {
		return fmt.Errorf("Jobs that require a fractional number of GPUs must require less than 1 GPU")
	}

	return nil
}
