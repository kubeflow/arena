package client

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

var (
	clientset    *kubernetes.Clientset
	clientConfig clientcmd.ClientConfig
)

func GetClientConfig() (*restclient.Config, error) {
	if clientConfig != nil {
		return clientConfig.ClientConfig()
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)

	return clientConfig.ClientConfig()
}

func GetClientSet() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}

	var err error
	restConfig, err := GetClientConfig()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return clientset, nil
}
