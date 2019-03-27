package commands

import (
	"os"

	config "github.com/kubeflow/arena/pkg/util/config"
	log "github.com/sirupsen/logrus"
)

const (
	RecommendedConfigPathEnvVar = "ARENA_CONFIG"
	DefaultArenaConfigPath      = "~/.arena/config"
)

var (
	arenaConfigs map[string]string
)

// LoadArenaClientConifg returns configs in map
func LoadArenaClientConifg() (configs map[string]string) {
	configs = make(map[string]string)
	envVarFileName := os.Getenv(RecommendedConfigPathEnvVar)
	_, err := os.Stat(envVarFileName)
	if err != nil {
		log.Debugf("Illegal arena config file: %s due to %v", envVarFileName, err)
		envVarFileName = DefaultArenaConfigPath
	}

	_, err = os.Stat(envVarFileName)
	if err != nil {
		log.Debugf("Illegal arena config file: %s due to %v", envVarFileName, err)
		return
	}

	return config.ReadConfigFile(envVarFileName)

}
