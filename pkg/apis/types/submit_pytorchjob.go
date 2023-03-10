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

	// ActiveDeadlineSeconds Specifies the duration (in seconds) since startTime during which the job can remain active
	// before it is terminated
	ActiveDeadlineSeconds int64 `yaml:"activeDeadlineSeconds,omitempty"`

	// Defines the TTL for cleaning up finished PytorchJobs. Defaults to infinite.
	TTLSecondsAfterFinished int32 `yaml:"ttlSecondsAfterFinished,omitempty"`
}
