package get

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	servejob "github.com/kubeflow/arena/pkg/jobs/serving"
	printer "github.com/kubeflow/arena/pkg/printer/serving"
	"github.com/kubeflow/arena/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/kubelet/kubeletconfig/util/log"
)

var (
	tablePrintTemplate = `NAME:             %s
NAMESPACE:        %s	
VERSION:          %s
DESIRED:          %d
AVAILABLE:        %d
SERVING TYPE:     %s
ENDPOINT ADDRESS: %s
ENDPOINT PORTS:   %s
AGE:              %s

%s
`
	// table header
	subtableHeader = "INSTANCE\tSTATUS\tAGE\tREADY\tRESTARTS\tNODE"
)

// entry function for "serve get"
func ServingGetExecute(client *kubernetes.Clientset, servingName, namespace, stype, version, format string) {
	// check some conditions are ok
	if err := PrepareCheck(stype); err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	if code := GetPrint(client, servingName, namespace, stype, version, format); code != 0 {
		os.Exit(code)
	}
}

// make some checks firstly
func PrepareCheck(stype string) error {
	err := servejob.CheckServingTypeIsOk(stype)
	if err != nil {
		return err
	}
	return nil
}
func GetPrint(client *kubernetes.Clientset, servingName, namespace, stype, version, format string) int {
	job, helpInfo, err := servejob.GetOnlyOneJob(client, namespace, servingName, stype, version)
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
	printJob := printer.NewServingJobPrinter(job)
	printInfoToBytes, err := FormatServingJobs(format, printJob)
	if err != nil {
		log.Errorf(err.Error())
		return 1
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, string(printInfoToBytes))
	w.Flush()
	return 0
}

// format the serving jobs information
func FormatServingJobs(format string, job printer.ServingJobPrinter) ([]byte, error) {
	switch format {
	case "json":
		return job.GetJson()
	case "yaml":
		return job.GetYaml()
	default:
		return []byte(customFormat(job)), nil
	}
}

// if format type is "wide",define our printable string
// 	subtableHeader = "INSTANCE\tSTATUS\tAGE\tRESTARTS\tNODE"

func customFormat(job printer.ServingJobPrinter) string {
	podInfoStringArray := []string{subtableHeader}
	for _, pod := range job.Pods {
		podInfoStringLine := fmt.Sprintf("%s\t%v\t%s\t%s\t%s\t%s",
			pod.PodName,
			pod.Status,
			pod.Age,
			pod.Ready,
			pod.RestartCount,
			pod.HostIP,
		)
		podInfoStringArray = append(podInfoStringArray, podInfoStringLine)
	}
	jobPrintString := fmt.Sprintf(
		tablePrintTemplate,
		job.GetName(),
		job.GetNamespace(),
		job.GetVersion(),
		job.Desired,
		job.Available,
		job.GetType(),
		job.EndpointAddress,
		job.EndpointPorts,
		job.Age,
		strings.Join(podInfoStringArray, "\n"),
	)
	return jobPrintString
}
