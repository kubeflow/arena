package printer

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	app_v1 "k8s.io/api/apps/v1"

	atypes "github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// define the  serving job printer
type ServingJobPrinter struct {
	atypes.Serving  `yaml:",inline"`
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
func NewServingJobPrinter(client *kubernetes.Clientset, deploy app_v1.Deployment, allPods []v1.Pod) ServingJobPrinter {
	servingJob := atypes.NewServingJob(client, deploy, allPods)
	simplePodList := []SimplePod{}
	for ind, pod := range servingJob.AllPods() {
		hostIP := pod.Status.HostIP
		if hostIP == "" {
			hostIP = "N/A"
		}
		if debugPod, err := yaml.Marshal(pod); err == nil {
			log.Debugf("Pod %d:\n%s", ind, string(debugPod))
		} else {
			log.Errorf("failed to marshal pod,reason: %s", err.Error())
		}
		status, totalContainers, restarts, readyCount := atypes.DefinePodPhaseStatus(pod)
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
		Desired:         servingJob.DesiredInstances(),
		Available:       servingJob.AvailableInstances(),
		Age:             servingJob.GetAge(),
		EndpointAddress: servingJob.GetClusterIP(),
		EndpointPorts:   servingJob.GetPorts(),
		Pods:            simplePodList,
		Serving:         servingJob,
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

// is matched the given namespace?
func (sjp ServingJobPrinter) IsMatchedGivenNamespace(namespace string) bool {
	if namespace == "" || namespace == sjp.Serving.Namespace {
		return true
	}
	return false
}

// is matched the given verison ?
func (sjp ServingJobPrinter) IsMatchedGivenVersion(version string) bool {
	if version == "" || version == sjp.Serving.Version {
		return true
	}
	return false
}

// is match the given type?
func (sjp ServingJobPrinter) IsMatchedGivenType(servingType string) bool {
	if servingType == "" || servingType == string(sjp.Serving.ServeType) {
		return true
	}
	return false
}
