package types

type SubmitETJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	CommonSubmitArgs `yaml:",inline"`
	// SubmitTensorboardArgs stores tensorboard information
	SubmitTensorboardArgs `yaml:",inline"`
	// SubmitSyncCodeArgs stores syncing code information
	SubmitSyncCodeArgs  `yaml:",inline"`
	MaxWorkers          int               `yaml:"maxWorkers"`
	MinWorkers          int               `yaml:"minWorkers"`
	LauncherSelectors   map[string]string `yaml:"launcherSelectors"`   // --launcher-selector
	JobRestartPolicy    string            `yaml:"jobRestartPolicy"`    // --job-restart-policy
	WorkerRestartPolicy string            `yaml:"workerRestartPolicy"` // --worker-restart-policy
	JobBackoffLimit     int               `yaml:"jobBackoffLimit"`     // --job-backoff-limit

}

type ScaleETJobArgs struct {
	//--name string     required, et job name
	Name string `yaml:"etName"`
	// TrainingType stores the trainingType
	JobType TrainingJobType `yaml:"-"`
	// Namespace  stores the namespace of job,match option --namespace
	Namespace string `yaml:"-"`
	//--timeout int     timeout of callback scaler script.
	Timeout int `yaml:"timeout"`
	//--retry int       retry times.
	Retry int `yaml:"retry"`
	//--count int       the nums of you want to add or delete worker.
	Count int `yaml:"count"`
	//--script string        script of scaling.
	Script string `yaml:"script"`
	//-e, --env stringArray      the environment variables
	Envs map[string]string `yaml:"envs"`
}

type ScaleInETJobArgs struct {
	// common args
	ScaleETJobArgs `yaml:",inline"`
}

type ScaleOutETJobArgs struct {
	// common args
	ScaleETJobArgs `yaml:",inline"`
}
