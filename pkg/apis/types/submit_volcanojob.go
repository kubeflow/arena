package types

type SubmitVolcanoJobArgs struct {
	// Name stores the job name
	Name string
	// Namespace stores the namespace of job
	Namespace string
	// TrainingType is used to accept job type
	TrainingType TrainingJobType
	// Command defines the job command
	Command string
	// The MinAvailable available pods to run for this Job
	MinAvailable int `yaml:"minAvailable"`
	// Specifies the queue that will be used in the scheduler, "default" queue is used this leaves empty.
	Queue string `yaml:"queue"`
	// SchedulerName is the default value of `tasks.template.spec.schedulerName`.
	SchedulerName string `yaml:"schedulerName"`
	// TaskName specifies the name of task
	TaskName string `yaml:"taskName"`
	// TaskImages specifies the task image
	TaskImages []string `yaml:"taskImages"`
	// TaskReplicas specifies the replicas of this Task in Job
	TaskReplicas int `yaml:"taskReplicas"`
	// TaskCPU specifies the cpu resource required for each replica of Task in Job. default is 250m
	TaskCPU string `yaml:"taskCPU"`
	// TaskMemory specifies the memory resource required for each replica of Task in Job. default is 128Mi
	TaskMemory string `yaml:"taskMemory"`
	// TaskPort specifies the task port
	TaskPort int `yaml:"taskPort"`
}
