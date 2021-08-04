package config

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"os"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	config "github.com/kubeflow/arena/pkg/util/config"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	extclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	RecommendedConfigPathEnvVar = "ARENA_CONFIG"
	DefaultArenaConfigPath      = "~/.arena/config"
)

var arenaClient *ArenaConfiger
var errInitArenaClient error

var once sync.Once

// InitArenaConfiger initilize
func InitArenaConfiger(args types.ArenaClientArgs) (*ArenaConfiger, error) {
	once.Do(func() {
		arenaClient, errInitArenaClient = newArenaConfiger(args)
	})
	return arenaClient, errInitArenaClient
}

// GetArenaConfiger returns the arena configer,it must be invoked after invoking function InitArenaConfiger(...)
func GetArenaConfiger() *ArenaConfiger {
	if arenaClient == nil {
		err := fmt.Errorf("ArenaClient is not initilized,but you want to get it")
		log.Errorf("Arena Client is not initilized,please use function InitArenaClient(...) to init it")
		panic(err)
	}
	return arenaClient
}

type tokenRetriever struct {
	rountTripper http.RoundTripper
	token        string
}

func (t *tokenRetriever) RoundTrip(req *http.Request) (*http.Response, error) {
	header := req.Header.Get("authorization")
	switch {
	case strings.HasPrefix(header, "Bearer "):
		t.token = strings.ReplaceAll(header, "Bearer ", "")
	}
	return t.rountTripper.RoundTrip(req)
}

type User struct {
	name string
	id   string
}

func (u User) GetName() string {
	return u.name
}

func (u User) GetId() string {
	return u.id
}

type ArenaConfiger struct {
	restConfig             *rest.Config
	clientConfig           clientcmd.ClientConfig
	clientset              *kubernetes.Clientset
	apiExtensionClientset  *extclientset.Clientset
	user                   User
	namespace              string
	arenaNamespace         string
	configs                map[string]string
	isDaemonMode           bool
	clusterInstalledCRDs   []string
	isolateUserInNamespace bool
	tokenRetriever         *tokenRetriever
}

func newArenaConfiger(args types.ArenaClientArgs) (*ArenaConfiger, error) {
	tr := &tokenRetriever{}
	arenaConfigs, err := loadArenaConifg()
	if err != nil {
		return nil, err
	}
	clientConfig, restConfig, clientSet, err := initKubeClient(args.Kubeconfig)
	if err != nil {
		return nil, err
	}
	restConfig.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		tr.rountTripper = rt
		return tr
	})
	apiExtensionClientSet, err := extclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	/*
		crdNames, err := getClusterInstalledCRDs(apiExtensionClientSet)
		if err != nil {
			return nil, err
		}
	*/
	namespace := updateNamespace(args.Namespace, arenaConfigs, clientConfig)
	log.Debugf("current namespace is %v", namespace)
	userName, err := getUserName(namespace, clientConfig, restConfig, clientSet, tr)
	if err != nil {
		return nil, err
	}
	log.Debugf("succeed to get user name: %v from client-go", *userName)
	userId := util.Md5(*userName)
	log.Debugf("the user id is %v", userId)
	i, err := isolateUserInNamespace(namespace, clientSet)
	if err != nil {
		return nil, err
	}
	log.Debugf("enable isolate user in namespace %v: %v", namespace, i)
	return &ArenaConfiger{
		restConfig:             restConfig,
		clientConfig:           clientConfig,
		clientset:              clientSet,
		apiExtensionClientset:  apiExtensionClientSet,
		namespace:              namespace,
		arenaNamespace:         args.ArenaNamespace,
		configs:                arenaConfigs,
		isDaemonMode:           args.IsDaemonMode,
		clusterInstalledCRDs:   []string{},
		user:                   User{name: *userName, id: userId},
		isolateUserInNamespace: i,
		tokenRetriever:         tr,
	}, nil

}

// GetClientConfig returns the kubernetes ClientConfig
func (a *ArenaConfiger) GetClientConfig() clientcmd.ClientConfig {
	return a.clientConfig
}

// GetRestConfig returns the kubernetes RestConfig
func (a *ArenaConfiger) GetRestConfig() *rest.Config {
	return a.restConfig
}

// GetClientSet returns the kubernetes ClientSet
func (a *ArenaConfiger) GetClientSet() *kubernetes.Clientset {
	return a.clientset
}

func (a *ArenaConfiger) GetAPIExtensionClientSet() *extclientset.Clientset {
	return a.apiExtensionClientset
}

// GetArenaNamespace returns the kubernetes namespace which some operators exists in
func (a *ArenaConfiger) GetArenaNamespace() string {
	return a.arenaNamespace
}

// GetNamespace returns the namespace of user assigns
func (a *ArenaConfiger) GetNamespace() string {
	return a.namespace
}

// GetConfigsFromConfigFile returns the configs read from config file
func (a *ArenaConfiger) GetConfigsFromConfigFile() map[string]string {
	return a.configs
}

func (a *ArenaConfiger) IsDaemonMode() bool {
	return a.isDaemonMode
}

func (a *ArenaConfiger) GetClusterInstalledCRDs() []string {
	return a.clusterInstalledCRDs
}

func (a *ArenaConfiger) GetUser() User {
	return a.user
}

func (a *ArenaConfiger) IsIsolateUserInNamespace() bool {
	return a.isolateUserInNamespace
}

// loadArenaConifg returns configs in map
func loadArenaConifg() (map[string]string, error) {
	arenaConfigs := map[string]string{}
	log.Debugf("start to init arena config")
	validateFile := func(file string) bool {
		if file == "" {
			return false
		}
		_, err := os.Stat(file)
		if err != nil {
			log.Debugf("failed to get state of file %v,reason: %v,skip to handle it", file, err)
			return false
		}
		return true
	}
	configFileName := os.Getenv(RecommendedConfigPathEnvVar)
	defaultConfigFile, err := homedir.Expand(DefaultArenaConfigPath)
	if err != nil {
		return arenaConfigs, err
	}
	// if config file path read from env is invalid,read it from default path
	if !validateFile(configFileName) {
		configFileName = defaultConfigFile
	}
	// if config file is invalid,return null
	if !validateFile(configFileName) {
		return arenaConfigs, nil
	}
	arenaConfigs = config.ReadConfigFile(configFileName)
	log.Debugf("arena configs: %v", arenaConfigs)
	return arenaConfigs, nil
}

func updateNamespace(namespace string, arenaConfigs map[string]string, clientConfig clientcmd.ClientConfig) string {
	if namespace != "" {
		return namespace
	}
	log.Debugf("we need to update the namespace")
	if n, ok := arenaConfigs["namespace"]; ok {
		log.Debugf("read namespace %v from arena configuration file", n)
		return n
	}
	n, _, err := clientConfig.Namespace()
	if err == nil {
		log.Debugf("read namespace %v from kubeconfig", n)
		return n
	}
	log.Debugf("failed to read namespace from kubeconfig,we set the default namespace with 'default'")
	return "default"
}

func isolateUserInNamespace(namespaceName string, clientSet *kubernetes.Clientset) (bool, error) {
	namespace, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return namespace.Labels[types.MultiTenantIsolationLabel] == "true", nil
}

func getClusterInstalledCRDs(client *extclientset.Clientset) ([]string, error) {
	selectorListOpts := metav1.ListOptions{}

	list, err := client.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), selectorListOpts)
	if err != nil {
		return nil, err
	}
	crds := []string{}
	for _, crd := range list.Items {
		crds = append(crds, crd.Name)
	}
	return crds, nil
}
