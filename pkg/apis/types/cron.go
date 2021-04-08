package types

// TrainingJobType defines the supporting training job type
type CronTaskType string

const (
	// CronTFTrainingJob defines the cron tfjob
	CronTFTrainingJob CronTaskType = "tfjob"
)

type ConcurrencyPolicy string

const (
	ConcurrencyAllow   ConcurrencyPolicy = "Allow"
	ConcurrencyForbid  ConcurrencyPolicy = "Forbid"
	ConcurrencyReplace ConcurrencyPolicy = "Replace"
)

type CronTaskInfo struct {
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
	JobsHistoryLimit int `yaml:"jobsHistoryLimit"` // --jobs-history-limit
}

type CronTFJobArgs struct {
	CommonCronArgs
	SubmitTFJobArgs
}
