package client

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	client *Client
)

type Client struct {
	clientset  *kubernetes.Clientset
	restConfig *restclient.Config
	namespace  string
}

func GetClient() (*Client, error) {
	if client != nil {
		return client, nil
	}

	getter := genericclioptions.NewConfigFlags(true)
	factory := cmdutil.NewFactory(getter)
	namespace, _, err := factory.ToRawKubeConfigLoader().Namespace()

	if err != nil {
		return nil, err
	}

	clientConfig := factory.ToRawKubeConfigLoader()
	restConfig, err := clientConfig.ClientConfig()

	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		namespace:  namespace,
		restConfig: restConfig,
		clientset:  clientset,
	}, nil
}

func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

func (c *Client) GetRestConfig() *restclient.Config {
	return c.restConfig
}

func (c *Client) GetDefaultNamespace() string {
	return c.namespace
}

func (c *Client) SetDefaultNamespace(namespace string) error {
	configAccess := clientcmd.DefaultClientConfig.ConfigAccess()
	config, err := configAccess.GetStartingConfig()

	if err != nil {
		return err
	}

	context := config.Contexts[config.CurrentContext]
	context.Namespace = namespace

	err = clientcmd.ModifyConfig(configAccess, *config, true)
	return err
}
