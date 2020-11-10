package commands

import (
	"os"

	"github.com/kubeflow/arena/pkg/util"

	"k8s.io/client-go/tools/clientcmd"
)

// InitCommonConfig is used for arena-go-sdk
func InitCommonConfig(kubeconfig, logLevel, ns string) error {
	if logLevel == "" {
		logLevel = "info"
	}
	util.SetLogLevel(logLevel)
	loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
	setupKubeconfig()
	_, err := initKubeClient()
	if err != nil {
		return err
	}
	namespace = ns
	if namespace == "" {
		namespace, _, err = clientConfig.Namespace()
		if err != nil {
			return err
		}
	}
	return nil
}
