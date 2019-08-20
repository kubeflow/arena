package logs

import (
	"path"

	servejob "github.com/kubeflow/arena/pkg/jobs/serving"
	"github.com/kubeflow/arena/pkg/podlogs"
	podlogger "github.com/kubeflow/arena/pkg/printer/base/logs"
)

type ServingPodLogPrinter struct {
	PodLogger *podlogger.PodLogPrinter
}

func NewServingPodLogPrinter(job servejob.Serving, logArgs *podlogs.OuterRequestArgs) (*ServingPodLogPrinter, error) {
	names := []string{}
	for _, pod := range job.AllPods() {
		names = append(names, path.Base(pod.ObjectMeta.SelfLink))
	}

	podLogPrinter, err := podlogger.NewPodLogPrinter(names, logArgs)
	if err != nil {
		return nil, err
	}
	return &ServingPodLogPrinter{
		PodLogger: podLogPrinter,
	}, nil
}

func (slp *ServingPodLogPrinter) Print() (int, error) {
	return slp.PodLogger.Print()
}
