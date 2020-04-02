package logs

import (
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/pkg/jobs/serving"
	servejob "github.com/kubeflow/arena/pkg/jobs/serving"
	"github.com/kubeflow/arena/pkg/podlogs"
	"github.com/kubeflow/arena/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func LogPrint(client *kubernetes.Clientset, ns, servingName, servingTypeKey, version string, args *podlogs.OuterRequestArgs) int {
	if err := serving.CheckServingTypeIsOk(servingTypeKey); err != nil {
		log.Errorf(err.Error())
		return 1
	}
	job, helpInfo, err := servejob.GetOnlyOneJob(client, ns, servingName, servingTypeKey, version)
	if err != nil {
		if err == types.ErrTooManyJobs {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, helpInfo)
			w.Flush()
			return 1
		}
		log.Errorf(err.Error())
		return 1
	}
	logPrinter, err := NewServingPodLogPrinter(job, args)
	if err != nil {
		log.Error(err.Error())
		return 1
	}
	code, err := logPrinter.Print()
	if err != nil {
		log.Errorf("%s, %s", err.Error(), "please use \"runai serve get\" to get more information.")
	}
	return code
}
