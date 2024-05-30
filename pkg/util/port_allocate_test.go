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

package util

import (
	"testing"
)

func TestSelectAvailablePort(t *testing.T) {
	clientset := GetClientSetForTest(t)
	if clientset == nil {
		t.Skip("kubeclient not setup")
	}
	port1, err := SelectAvailablePort(clientset)
	if err != nil {
		t.Errorf("failed to SelectAvailablePort, %++v", err)
	}
	t.Logf("port is %d", port1)

	port2, err := SelectAvailablePort(clientset)
	if err != nil {
		t.Errorf("failed to SelectAvailablePort, %++v", err)
	}
	t.Logf("port is %d", port2)
	if port2 != port1+1 {
		t.Errorf("Port should be %d, when latest port is %d", port1+1, port1)
	}

	k8sClusterUsedPorts = []int{20000, 20001}
	port3, err := SelectAvailablePort(clientset)
	if err != nil {
		t.Errorf("failed to SelectAvailablePort, %++v", err)
	}
	t.Logf("port is %d", port3)
	if port3 != 20002 {
		t.Errorf("Port should be 30002, when 30000,30001 is used")
	}
	port4, err := SelectAvailablePortWithDefault(clientset, 0)
	if err != nil {
		t.Errorf("failed to SelectAvailablePortWithDefault, %++v", err)
	}
	t.Logf("port is %d", port4)
	if port4 == port3 {
		t.Errorf("If default port is used, chose another one")
	}
}
