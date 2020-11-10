package commands

// ListTrainingJobs is used for arena-go-sdk
func ListTrainingJobs(kubeconfig, logLevel, ns string) ([]TrainingJob, error) {
	if err := InitCommonConfig(kubeconfig, logLevel, ns); err != nil {
		return nil, err
	}
	useCache = false
	var err error
	allPods, err = acquireAllPods(clientset)
	if err != nil {
		return nil, err
	}
	allJobs, err = acquireAllJobs(clientset)
	if err != nil {
		return nil, err
	}
	jobs := []TrainingJob{}
	trainers := NewTrainers(clientset)
	for _, trainer := range trainers {
		trainingJobs, err := trainer.ListTrainingJobs()
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, trainingJobs...)
	}

	jobs = makeTrainingJobOrderdByAge(jobs)
	return jobs, nil
}
