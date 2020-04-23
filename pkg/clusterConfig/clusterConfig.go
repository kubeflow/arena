package clusterConfig

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterConfig struct {
	Name        string
	Description string
	Values      string
	IsDefault   bool
}

type ClusterConfigs struct {
	clientset kubernetes.Interface
}

var (
	runaiNamespace         = "runai"
	runaiConfigLabel       = "runai/template"
	runaiNameLabel         = "runai/name"
	runaiDefaultAnnotation = "runai/default"
)

func NewClusterConfigs(clientset kubernetes.Interface) ClusterConfigs {
	return ClusterConfigs{
		clientset: clientset,
	}
}

func (cg *ClusterConfigs) ListClusterConfigs() ([]ClusterConfig, error) {
	configsList, err := cg.clientset.CoreV1().ConfigMaps(runaiNamespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", runaiConfigLabel),
	})

	if err != nil {
		return []ClusterConfig{}, err
	}

	log.Debugf("Found %d templates", len(configsList.Items))

	clusterConfigs := []ClusterConfig{}

	for _, config := range configsList.Items {
		clusterConfig := ClusterConfig{}

		if config.Annotations != nil {
			clusterConfig.IsDefault = config.Annotations[runaiDefaultAnnotation] == "true"
		}

		clusterConfig.Name = config.Data["name"]
		clusterConfig.Description = config.Data["description"]
		clusterConfig.Values = config.Data["values"]
		clusterConfigs = append(clusterConfigs, clusterConfig)
	}

	return clusterConfigs, nil
}

func (cg *ClusterConfigs) GetClusterConfig(name string) (*ClusterConfig, error) {
	configs, err := cg.ListClusterConfigs()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.Name == name {
			return &config, err
		}
	}

	return nil, nil
}

func (cg *ClusterConfigs) GetClusterDefaultConfig() (*ClusterConfig, error) {
	configs, err := cg.ListClusterConfigs()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.IsDefault {
			return &config, err
		}
	}

	return nil, nil
}
