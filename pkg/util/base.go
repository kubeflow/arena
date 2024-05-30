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
	"crypto/md5"
	"fmt"
	"os"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClientSetForTest(t *testing.T) *kubernetes.Clientset {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		t.Logf("failed to initclient, %++v", err)
	}
	if restConfig == nil {
		t.Logf("Kube Client is not setup")
		return nil
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		t.Errorf("failed to NewForConfig, %++v", err)
	}
	return clientset
}

func Md5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}
