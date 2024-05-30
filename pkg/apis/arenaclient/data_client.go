// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
