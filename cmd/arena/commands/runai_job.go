package commands

import (
	"k8s.io/client-go/kubernetes"
)

type RunaiJob struct {
	*JobInfo
}

// Get Dashboard
func (rj *RunaiJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	return []string{}, nil
}

// The priority class name of the training job
func (rj *RunaiJob) GetPriorityClass() string {
	return ""
}
