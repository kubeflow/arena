package types

type SubmitMPIJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// for tensorboard
	SubmitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`

	// enable gpu topology scheduling
	GPUTopology        bool   `yaml:"gputopology"`
	GPUTopologyReplica string `yaml:"gputopologyreplica"`
	MountsOnLauncher   bool   `yaml:"mountsOnLauncher"`

	// clean-task-policy
	CleanPodPolicy string `yaml:"cleanPodPolicy"`
}
