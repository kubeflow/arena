// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package config

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func initKubeClient(kubeconfig string) (clientcmd.ClientConfig, *rest.Config, *kubernetes.Clientset, error) {
	var err error
	kubeconfig, err = setupKubeconfig(kubeconfig)
	if err != nil {
		return nil, nil, nil, err
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
	// create rest config
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, nil, err
	}
	restConfig.QPS = 10
	restConfig.Burst = 20
	// create the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	return clientConfig, restConfig, clientset, nil
}

func setupKubeconfig(kubeconfig string) (string, error) {
	// if kubeconfig is null and env "KUBECONFIG" is not null
	// read kubeconfig from env
	currentUser, err := user.Current()
	if err != nil {
		return kubeconfig, err
	}
	switch {
	case kubeconfig != "":
		break
	case os.Getenv("KUBECONFIG") != "":
		kubeconfig = os.Getenv("KUBECONFIG")
	default:
		// if default kubeconfig is invalid,set kubeconfig null value and return it
		defaultKubeconfig := path.Join(currentUser.HomeDir, ".kube", "config")
		_, err = os.Stat(defaultKubeconfig)
		if err != nil {
			return kubeconfig, nil
		}
		kubeconfig = defaultKubeconfig
	}
	// normalize path
	kubeconfig = filepath.Clean(kubeconfig)
	// change ~ to user home dir
	kubeconfig = strings.ReplaceAll(kubeconfig, "~", currentUser.HomeDir)
	_ = os.Setenv("KUBECONFIG", kubeconfig)
	// set env
	_, err = os.Stat(kubeconfig)
	return kubeconfig, err
}
