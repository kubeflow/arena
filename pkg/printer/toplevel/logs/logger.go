package logs

import (

	//"github.com/kubeflow/arena/cmd/arena/commands"

	podlogger "github.com/kubeflow/arena/pkg/printer/base/logs"
)

type TopLevelPodLogPrinter struct {
	PodLogger *podlogger.PodLogPrinter
}

/*
func NewTopLevelPodLogPrinter(job commands.TrainingJob, logArgs *podlogs.OuterRequestArgs) (*TopLevelPodLogPrinter, error) {
	names := []string{}
	for _, pod := range job.AllPods() {
		names = append(names, path.Base(pod.ObjectMeta.SelfLink))
	}

	podLogPrinter, err := podlogger.NewPodLogPrinter(names, logArgs)
	if err != nil {
		return nil, err
	}
	return &TopLevelPodLogPrinter{
		PodLogger: podLogPrinter,
	}, nil
}

func (tllp *TopLevelPodLogPrinter) Print() (int, error) {
	return tllp.PodLogger.Print()
}*/
