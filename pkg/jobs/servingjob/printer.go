package serving

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
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

type Printer interface {
	GetJsonPrintString() (string, error)
	GetYamlPrintString() (string, error)
	GetWidePrintString() (string, error)
	GetHelpInfo(obj ...interface{}) (string, error)
}

// define the  serving job printer
type servingJobPrinter struct {
	Name            string            `yaml:"name" json:"name"`
	Namespace       string            `yaml:"namespace" json:"namespace"`
	ServingType     types.ServingType `yaml:"serving_type" json:"serving_type"`
	Version         string            `yaml:"version" json:"version"`
	Desired         int32             `yaml:"desired" json:"desired"`
	Available       int32             `yaml:"available" json:"available"`
	Age             string            `yaml:"serving_duration" json:"serving_duration"`
	EndpointAddress string            `yaml:"endpoint_address" json:"endpoint_address"`
	EndpointPorts   string            `yaml:"endpoint_ports" json:"endpoint_ports"`
	Pods            []simplePod       `yaml:"instances" json:"instances"`
}

// define the simple pod information
type simplePod struct {
	PodName string `yaml:"pod_name" json:"pod_name"` // selfLink
	// how long the pod is running
	Age string `yaml:"age" json:"age"`
	// pod' status,there is "Running" and "Pending"
	Status string `yaml:"status" json:"status"`
	// the node ip
	HostIP       string `yaml:"host_ip" json:"host_ip"`
	Ready        string `yaml:"ready" json:"ready"`
	RestartCount string `yaml:"restart_count" json:"restart_count"`
}

func NewServingJobPrinter(s ServingJob) Printer {
	printer := servingJobPrinter{
		Name:            s.GetName(),
		Namespace:       s.GetNamespace(),
		ServingType:     s.GetType(),
		Version:         s.GetVersion(),
		Desired:         s.DesiredInstances(),
		Available:       s.AvailableInstances(),
		Age:             s.GetAge(),
		EndpointAddress: s.GetClusterIP(),
		EndpointPorts:   s.GetPorts(),
	}
	simples := []simplePod{}
	for _, pod := range s.GetAllPods() {
		simples = append(simples, NewSimplePod(pod))
	}
	printer.Pods = simples
	return printer
}

func (s servingJobPrinter) GetJsonPrintString() (string, error) {
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (s servingJobPrinter) GetYamlPrintString() (string, error) {
	yamlBytes, err := yaml.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}
func (s servingJobPrinter) GetWidePrintString() (string, error) {
	printer := s
	podInfoStringArray := []string{subtableHeader}
	for _, pod := range printer.Pods {
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
	wide := fmt.Sprintf(
		tablePrintTemplate,
		printer.Name,
		printer.Namespace,
		printer.Version,
		printer.Desired,
		printer.Available,
		string(printer.ServingType),
		printer.EndpointAddress,
		printer.EndpointPorts,
		printer.Age,
		strings.Join(podInfoStringArray, "\n"),
	)
	return wide, nil
}
func (s servingJobPrinter) GetHelpInfo(obj ...interface{}) (string, error) {
	header := fmt.Sprintf("There is %d instances(pods) have been found:", len(s.Pods))
	printLines := []string{}
	footer := fmt.Sprintf("please use \"--instance\" or \"-p\" to assign one.")
	for _, pod := range s.Pods {
		line := fmt.Sprintf("\t%s",
			pod.PodName,
		)
		printLines = append(printLines, line)
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(printLines, "\n"), footer),nil 
}

func NewSimplePod(pod v1.Pod) simplePod {
	hostIP := pod.Status.HostIP
	if hostIP == "" {
		hostIP = "N/A"
	}
	status, totalContainers, restarts, readyCount := DefinePodPhaseStatus(pod)
	age := util.ShortHumanDuration(time.Now().Sub(pod.ObjectMeta.CreationTimestamp.Time))
	return simplePod{
		PodName:      path.Base(pod.ObjectMeta.SelfLink),
		Age:          age,
		Status:       status,
		HostIP:       hostIP,
		RestartCount: fmt.Sprintf("%d", restarts),
		Ready:        fmt.Sprintf("%d/%d", readyCount, totalContainers),
	}
}
