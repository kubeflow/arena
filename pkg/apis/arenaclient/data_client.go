package arenaclient

import (
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/datahouse"
)

type DataClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

// NewDataClient creates a ServingJobClient
func NewDataClient(namespace string, configer *config.ArenaConfiger) *DataClient {
	return &DataClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Namespace sets the namespace,this operation does not change the default namespace
func (d *DataClient) Namespace(namespace string) *DataClient {
	copyDataClient := &DataClient{
		namespace: namespace,
		configer:  d.configer,
	}
	return copyDataClient
}

func (d *DataClient) ListAndPrintDataVolumes(namespace string, allNamespaces bool) error {
	return datahouse.DisplayDataVolumes(namespace, allNamespaces)
}
