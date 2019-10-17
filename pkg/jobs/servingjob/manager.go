package serving
import (
	"strings"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kubeflow/arena/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type ServingJobManager interface {
	GetAllServingJob() []ServingJob
	GetTargetServingJob(filter ServingJobFilterArgs) ([]ServingJob)
	Printer
}
type ServingJobFilterArgs struct {
	Namespace string
	Type string
	Version string 
	Name string  
}

type manager struct  {
	jobs []ServingJob
}

func NewServingJobManager(client *kubernetes.Clientset, ns string)  (ServingJobManager,error) {
	// get all deployments with label "serviceName"
	jobs := []ServingJob{}
	deployments, err := client.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: "serviceName",
	})
	if err != nil {
		return nil, fmt.Errorf("Failed due to %v", err)
	}
	// get all pods with label "serviceName"
	podListObject, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: "serviceName",
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get pods by label serviceName=,reason=%s", err.Error())
	}
	// get all service with label "serviceName"
	serviceList, err := client.CoreV1().Services(ns).List(metav1.ListOptions{
		LabelSelector: "servingName",
	})
	if err != nil {
		return nil,fmt.Errorf("Failed to list services due to %v", err)
	}
	pods := podListObject.Items
	svcs := serviceList.Items
	for _, deploy := range deployments.Items {
		jobs = append(jobs,NewServingJob(deploy,pods,svcs))
	}
	return &manager{jobs: jobs}, nil
}

func (m *manager) GetAllServingJob() []ServingJob {
	return m.jobs
}

func (m *manager) GetTargetServingJob(filter ServingJobFilterArgs) ([]ServingJob) {
	jobs := []ServingJob{}
	for _,job := range m.jobs {
		if !job.IsMatchedTargetType(filter.Type) {
			continue
		}
		if !job.IsMatchedTargetVersion(filter.Version) {
			continue
		}
		if !job.IsMatchedTargetNamespace(filter.Namespace) {
			continue
		}
		if !job.IsMatchedTargetName(filter.Name) {
			continue
		}
		jobs = append(jobs,job) 
	}
	return jobs
}
// TODO: return manager information with json format
func (m *manager) GetJsonPrintString() (string, error) {
	return "",nil
}
// TODO: return mamanger information with yaml format
func (m *manager) GetYamlPrintString() (string, error) {
	return "",nil
}
// TODO: return manager information with wide format 
func (m *manager) GetWidePrintString() (string, error) {
	return "",nil
}
// implement interface Printer
func (m *manager) GetHelpInfo(objs ...interface{}) (string, error) {
	if len(objs) < 1 {
		return "",fmt.Errorf("you should give args for function GetHelpInfo")
	}
	obj := objs[0]
	jobs := obj.([]ServingJob)
	if len(jobs) == 0 {
		return "",types.ErrNotFoundJobs
	}
	header := fmt.Sprintf("There is %d jobs have been found:", len(jobs))
	tableHeader := "NAME\tTYPE\tVERSION"
	printLines := []string{tableHeader}
	footer := fmt.Sprintf("please use \"--type\" or \"--version\" to filter.")
	for _, job := range jobs {
		line := fmt.Sprintf("%s\t%s\t%s",
			job.GetName(),
			string(job.GetType()),
			job.GetVersion(),
		)
		printLines = append(printLines, line)
	}
	helpInfo := fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(printLines, "\n"), footer)
	return helpInfo,nil
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	//fmt.Fprintf(w, helpInfo)
	//w.Flush()
}
