package util

import (
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"os"
	"k8s.io/client-go/kubernetes"
)

func getClientSet(t *testing.T) *kubernetes.Clientset {
	// The "usual" clientcmd/kubectl flags
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		t.Errorf("failed to initclient, %++v", err)
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		t.Errorf("failed to NewForConfig, %++v", err)
	}
	return clientset
}

func TestGetAvailablePort(t *testing.T)  {
	clientset := getClientSet(t)
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
	if port2 != port1 + 1 {
		t.Errorf("Port should be %s, when latest port is %d", port1 + 1, port1)
	}

	k8sClusterUsedPorts = []int {30000, 30001}
	port3, err := SelectAvailablePort(clientset)
	if err != nil {
		t.Errorf("failed to SelectAvailablePort, %++v", err)
	}
	t.Logf("port is %d", port3)
	if port3 != 30002 {
		t.Errorf("Port should be 30002, when 30000,30001 is used")
	}
	port4, err := SelectAvailablePortWithDefault(clientset, port3)
	if err != nil {
		t.Errorf("failed to SelectAvailablePortWithDefault, %++v", err)
	}
	t.Logf("port is %d", port4)
	if port4 == port3 {
		t.Errorf("If default port is used, chose another one")
	}
}