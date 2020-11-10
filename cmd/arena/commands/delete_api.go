package commands

// DeleteTrainingJob is used for arena-go-sdk
func DeleteTrainingJob(kubeconfig, logLevel, ns, jobType string, jobNames ...string) error {
	if err := InitCommonConfig(kubeconfig, logLevel, ns); err != nil {
		return err
	}
	for _, jobName := range jobNames {
		if err := deleteTrainingJob(jobName, jobType); err != nil {
			return err
		}
	}
	return nil
}
