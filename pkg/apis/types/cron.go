package types

// CronType defines the supporting job type
type CronType string

const (
	// CronTFTrainingJob defines the cron tfjob
	CronTFTrainingJob CronType = "tfjob"
)

// ConcurrencyPolicy describes how the job will be handled.
// Only one of the following concurrent policies may be specified.
// If none of the following policies is specified, the default one
// is AllowConcurrent.
type ConcurrencyPolicy string

const (
	ConcurrencyAllow   ConcurrencyPolicy = "Allow"
	ConcurrencyForbid  ConcurrencyPolicy = "Forbid"
	ConcurrencyReplace ConcurrencyPolicy = "Replace"
)

// JobConditionType defines all kinds of types of JobStatus.
type JobConditionType string

const (
	// JobCreated means the job has been accepted by the system,
	// but one or more of the pods/services has not been started.
	// This includes time before pods being scheduled and launched.
	JobCreated JobConditionType = "Created"

	// JobRunning means all sub-resources (e.g. services/pods) of this job
	// have been successfully scheduled and launched.
	// The training is running without error.
	JobRunning JobConditionType = "Running"

	// JobRestarting means one or more sub-resources (e.g. services/pods) of this job
	// reached phase failed but maybe restarted according to it's restart policy
	// which specified by user in v1.PodTemplateSpec.
	// The training is freezing/pending.
	JobRestarting JobConditionType = "Restarting"

	// JobSucceeded means all sub-resources (e.g. services/pods) of this job
	// reached phase have terminated in success.
	// The training is complete without error.
	JobSucceeded JobConditionType = "Succeeded"

	// JobFailed means one or more sub-resources (e.g. services/pods) of this job
	// reached phase failed with no restarting.
	// The training has failed its execution.
	JobFailed JobConditionType = "Failed"
)

type CronInfo struct {
	Name string `json:"name" yaml:"name"`

	Namespace string `json:"namespace" yaml:"namespace"`

	// Type is the job type, like TFjob、PyTorchJob
	Type string `json:"type" yaml:"type"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule" yaml:"schedule"`

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy string `json:"concurrencyPolicy" yaml:"concurrencyPolicy"` // --concurrency-policy

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend bool `json:"suspend" yaml:"suspend"` // --suspend

	// Deadline is the timestamp that a cron job can keep scheduling util then.
	Deadline string `json:"deadline" yaml:"deadline"` // --deadline

	// The number of finished job history to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	HistoryLimit int64 `json:"historyLimit" yaml:"historyLimit"` // --history-limit

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime string `json:"lastScheduleTime" yaml:"lastScheduleTime"`

	// CreationTimestamp stores the creation timestamp of job
	CreationTimestamp string `json:"creationTimestamp" yaml:"creationTimestamp"`

	History []CronHistoryInfo `json:"cronHistory" yaml:"cronHistory"`
}

type CronHistoryInfo struct {
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	Group      string `json:"group" yaml:"group"`
	Kind       string `json:"kind" yaml:"kind"`
	Status     string `json:"status" yaml:"status"`
	CreateTime string `json:"createTime" yaml:"createTime"`
	FinishTime string `json:"finishTime" yaml:"finishTime"`
}

type CommonCronArgs struct {
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `yaml:"schedule"` // --schedule

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy string `yaml:"concurrencyPolicy"` // --concurrency-policy

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend bool `yaml:"suspend"` // --suspend

	// Deadline is the timestamp that a cron job can keep scheduling util then.
	Deadline string `yaml:"deadline"` // --deadline

	// The number of finished job history to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	HistoryLimit int `yaml:"historyLimit"` // --history-limit
}

type CronTFJobArgs struct {
	CommonCronArgs  `yaml:"cron"`
	SubmitTFJobArgs `yaml:"tfjob"`
}
