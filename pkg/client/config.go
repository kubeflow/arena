package client

import (
	"github.com/kubeflow/arena/pkg/constants"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LoadRestConfig builds a *rest.Config using the standard kubeconfig resolution chain.
// If kubeconfig is empty, it uses the default loading rules (KUBECONFIG env, ~/.kube/config).
// If kubeContext is non-empty, it overrides the current-context.
// QPS and Burst are set to sensible defaults for CLI workloads.
func LoadRestConfig(kubeconfig, kubeContext string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	overrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		overrides.CurrentContext = kubeContext
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	config.QPS = float32(constants.DefaultQPS)
	config.Burst = constants.DefaultBurst
	return config, nil
}

// ResolveNamespace determines the effective namespace for the CLI.
// Priority: CLI flag > kubeconfig context namespace > "default".
func ResolveNamespace(kubeconfig, kubeContext, cliNamespace string) string {
	if cliNamespace != "" {
		return cliNamespace
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	overrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		overrides.CurrentContext = kubeContext
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	ns, _, err := clientConfig.Namespace()
	if err == nil && ns != "" {
		return ns
	}
	return "default"
}
