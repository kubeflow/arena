package types

type SubmitSparkJobArgs struct {
	Name         string          `yaml:"-"`
	Namespace    string          `yaml:"-"`
	TrainingType TrainingJobType `yaml:"-"`
	Image        string          `yaml:"Image"`
	MainClass    string          `yaml:"MainClass"`
	Jar          string          `yaml:"Jar"`
	Executor     *Executor       `yaml:"Executor"`
	Driver       *Driver         `yaml:"Driver"`
}

type Driver struct {
	CPURequest    int    `yaml:"CPURequest"`
	MemoryRequest string `yaml:"MemoryRequest"`
}

type Executor struct {
	Replicas      int    `yaml:"Replicas"`
	CPURequest    int    `yaml:"CPURequest"`
	MemoryRequest string `yaml:"MemoryRequest"`
}
