package util

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
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
