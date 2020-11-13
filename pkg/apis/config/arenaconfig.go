package config

import (
	"fmt"
	"sync"

	"os"

	config "github.com/kubeflow/arena/pkg/util/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"

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

type ArenaConfiger struct {
	restConfig     *rest.Config
	clientConfig   clientcmd.ClientConfig
	clientset      *kubernetes.Clientset
	namespace      string
	arenaNamespace string
	configs        map[string]string
	isDaemonMode   bool
}

func newArenaConfiger(args types.ArenaClientArgs) (*ArenaConfiger, error) {
	arenaConfigs, err := loadArenaConifg()
	if err != nil {
		return nil, err
	}
	clientConfig, restConfig, clientSet, err := initKubeClient(args.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return &ArenaConfiger{
		restConfig:     restConfig,
		clientConfig:   clientConfig,
		clientset:      clientSet,
		namespace:      args.Namespace,
		arenaNamespace: args.ArenaNamespace,
		configs:        arenaConfigs,
		isDaemonMode:   args.IsDaemonMode,
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

// loadArenaConifg returns configs in map
func loadArenaConifg() (map[string]string, error) {
	arenaConfigs := map[string]string{}
	log.Debugf("init arena config")
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
