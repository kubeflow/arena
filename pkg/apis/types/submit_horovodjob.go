package types

type SubmitHorovodJobArgs struct {
	SSHPort int    `yaml:"sshPort"`
	Cpu     string `yaml:"cpu"`    // --cpu
	Memory  string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// for tensorboard
	SubmitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`
}
