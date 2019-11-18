package printer

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/kubeflow/arena/pkg/jobs/serving"
	servejob "github.com/kubeflow/arena/pkg/jobs/serving"

	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// define the  serving job printer
type ServingJobPrinter struct {
	serving.Serving `yaml:",inline"`
	Desired         int32       `yaml:"desired" json:"desired"`
	Available       int32       `yaml:"available" json:"available"`
	Age             string      `yaml:"serving_duration" json:"serving_duration"`
	EndpointAddress string      `yaml:"endpoint_address" json:"endpoint_address"`
	EndpointPorts   string      `yaml:"endpoint_ports" json:"endpoint_ports"`
	Pods            []SimplePod `yaml:"instances" json:"instances"`
}

// define the simple pod information
type SimplePod struct {
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

// create the printer
func NewServingJobPrinter(job servejob.Serving) ServingJobPrinter {
	simplePodList := []SimplePod{}
	for ind, pod := range job.AllPods() {
		hostIP := pod.Status.HostIP
		if hostIP == "" {
			hostIP = "N/A"
		}
		if debugPod, err := yaml.Marshal(pod); err == nil {
			log.Debugf("Pod %d:\n%s", ind, string(debugPod))
		} else {
			log.Errorf("failed to marshal pod,reason: %s", err.Error())
		}
		status, totalContainers, restarts, readyCount := servejob.DefinePodPhaseStatus(pod)
		age := util.ShortHumanDuration(time.Now().Sub(pod.ObjectMeta.CreationTimestamp.Time))
		simplePod := SimplePod{
			PodName:      path.Base(pod.ObjectMeta.SelfLink),
			Age:          age,
			Status:       status,
			HostIP:       hostIP,
			RestartCount: fmt.Sprintf("%d", restarts),
			Ready:        fmt.Sprintf("%d/%d", readyCount, totalContainers),
		}
		simplePodList = append(simplePodList, simplePod)
	}
	return ServingJobPrinter{
		Desired:         job.DesiredInstances(),
		Available:       job.AvailableInstances(),
		Age:             job.GetAge(),
		EndpointAddress: job.GetEndpointIP(),
		EndpointPorts:   job.GetPorts(),
		Pods:            simplePodList,
		Serving:         job,
	}
}

// get the name of printer
func (sjp ServingJobPrinter) GetName() string {
	return sjp.Serving.Name
}

// get the namespace of printer
func (sjp ServingJobPrinter) GetNamespace() string {
	return sjp.Serving.Namespace
}

// get the version of printer
func (sjp ServingJobPrinter) GetVersion() string {
	return sjp.Serving.Version
}

// get the serving type of printer
func (sjp ServingJobPrinter) GetType() string {
	return string(sjp.Serving.ServeType)
}

// output the printer with json format
func (sjp ServingJobPrinter) GetJson() ([]byte, error) {
	return json.Marshal(sjp)
}

// output the printer with yaml format
func (sjp ServingJobPrinter) GetYaml() ([]byte, error) {
	return yaml.Marshal(sjp)
}
