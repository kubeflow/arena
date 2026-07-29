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
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func initKubeClient(kubeconfig string) (clientcmd.ClientConfig, *rest.Config, *kubernetes.Clientset, error) {
	loadingRules, err := buildClientConfigLoadingRules(kubeconfig)
	if err != nil {
		return nil, nil, nil, err
	}
	if loadingRules.ExplicitPath != "" {
		// Keep KUBECONFIG in sync so Arena's kubectl subprocesses inherit the
		// explicitly selected kubeconfig.
		if err := os.Setenv("KUBECONFIG", loadingRules.ExplicitPath); err != nil {
			return nil, nil, nil, err
		}
	}
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

func buildClientConfigLoadingRules(kubeconfig string) (*clientcmd.ClientConfigLoadingRules, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	// When kubeconfig is empty, leave ExplicitPath unset so client-go merges
	// the files listed in KUBECONFIG using its native loading precedence.
	if kubeconfig != "" {
		explicitPath, err := setupExplicitKubeconfig(kubeconfig)
		if err != nil {
			return nil, err
		}
		loadingRules.ExplicitPath = explicitPath
	}

	return loadingRules, nil
}

func setupExplicitKubeconfig(kubeconfig string) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return kubeconfig, err
	}

	if len(kubeconfig) >= 2 && kubeconfig[:2] == "~/" {
		kubeconfig = filepath.Join(currentUser.HomeDir, kubeconfig[2:])
	} else if kubeconfig == "~" {
		kubeconfig = currentUser.HomeDir
	} else {
		kubeconfig = filepath.Clean(kubeconfig)
	}

	if _, err = os.Stat(kubeconfig); err != nil {
		return kubeconfig, err
	}
	return kubeconfig, nil
}
