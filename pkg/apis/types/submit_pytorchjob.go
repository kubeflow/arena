package types

type SubmitPyTorchJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// for tensorboard
	SubmitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`

	// worker init pytorch image, default "alpine:3.10";
	// TODO jiaqianjing: user can set init-pytorch container image by param "--worker-init-pytorch-image"
	// WorkerInitPytorchImage string `yaml: workerInitPytorchImage`

	// clean-task-policy
	CleanPodPolicy string `yaml:"cleanPodPolicy"`
}
