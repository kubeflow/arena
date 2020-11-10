package commands

// SearchTrainingJob is used for arena-go-sdk
func SearchTrainingJob(kubeconfig, logLevel, ns, jobName, jobType string) (TrainingJob, error) {
	if err := InitCommonConfig(kubeconfig, logLevel, ns); err != nil {
		return nil, err
	}
	return searchTrainingJob(jobName, jobType, ns)
}
