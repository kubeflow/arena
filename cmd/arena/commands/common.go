package commands

import (
	"os"

	"github.com/kubeflow/arena/types"
	log "github.com/sirupsen/logrus"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Global variables
var (
	restConfig   *rest.Config
	clientConfig clientcmd.ClientConfig
	clientset    *kubernetes.Clientset
	// To reduce client-go API call, for 'arena list' secnario
	allPods        []v1.Pod
	allJobs        []batchv1.Job
	name           string
	namespace      string
	arenaNamespace string // the system namespace of arena
)

func initKubeClient() (*kubernetes.Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}
	var err error
	restConfig, err = clientConfig.ClientConfig()
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

func setupKubeconfig() {
	// rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if len(loadingRules.ExplicitPath) == 0 {
		if len(os.Getenv("KUBECONFIG")) > 0 {
			loadingRules.ExplicitPath = os.Getenv("KUBECONFIG")
		}
	}

	if len(loadingRules.ExplicitPath) > 0 {
		if _, err := os.Stat(loadingRules.ExplicitPath); err != nil {
			log.Warnf("Illegal kubeconfig file: %s", loadingRules.ExplicitPath)
		} else {
			log.Infof("Use specified kubeconfig file %s", loadingRules.ExplicitPath)
			types.KubeConfig = loadingRules.ExplicitPath
			os.Setenv("KUBECONFIG", loadingRules.ExplicitPath)
		}
	}
}
