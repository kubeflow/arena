// Copyright 2021 The Kubeflow Authors
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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/kubeflow/arena/pkg/operators/training-operator/client/clientset/versioned/typed/kubeflow.org/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeKubeflowV1 struct {
	*testing.Fake
}

func (c *FakeKubeflowV1) MPIJobs(namespace string) v1.MPIJobInterface {
	return &FakeMPIJobs{c, namespace}
}

func (c *FakeKubeflowV1) MXJobs(namespace string) v1.MXJobInterface {
	return &FakeMXJobs{c, namespace}
}

func (c *FakeKubeflowV1) PaddleJobs(namespace string) v1.PaddleJobInterface {
	return &FakePaddleJobs{c, namespace}
}

func (c *FakeKubeflowV1) PyTorchJobs(namespace string) v1.PyTorchJobInterface {
	return &FakePyTorchJobs{c, namespace}
}

func (c *FakeKubeflowV1) TFJobs(namespace string) v1.TFJobInterface {
	return &FakeTFJobs{c, namespace}
}

func (c *FakeKubeflowV1) XGBoostJobs(namespace string) v1.XGBoostJobInterface {
	return &FakeXGBoostJobs{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeKubeflowV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
