package util
import (
	"os"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// create k8s client set
func CreateK8sClientSet(configFile string) (*kubernetes.Clientset,error) {
    config,err := CreateK8sConfig(configFile)
    if err != nil {
        return nil,err
    }
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }
    return clientset,nil
}

func CreateK8sClientSetWithConfig(config *rest.Config) (*kubernetes.Clientset,error) {
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }
    return clientset,nil
}

// create k8s client config
func CreateK8sConfig(confFile string) (*rest.Config, error) {
	var config *rest.Config
	var err error
	switch {
	// build client from rest config
	case confFile == "", !CheckFileExist(confFile):
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	// build client from client config file
	case CheckFileExist(confFile):
		config, err = clientcmd.BuildConfigFromFlags("", confFile)
		if err != nil {
			return nil, err
		}
	}
	return config,nil
}

func CheckFileExist(filename string) bool {
	var exist = true
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		exist = false
	}
	return exist
}
