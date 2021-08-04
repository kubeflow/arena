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
	// Annotations defines pod annotations of job,match option --annotation
	Annotations map[string]string `yaml:"annotations"`
	// Labels specify the job labels and it is work for pods
	Labels map[string]string `yaml:"labels"`
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
