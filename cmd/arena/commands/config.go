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
	arenaConfigs           map[string]string
	alreadyInitArenaConfig bool
)

// loadArenaConifg returns configs in map
func loadArenaConifg() {
	if alreadyInitArenaConfig {
		return
	}
	log.Debugf("Init arena config")
	alreadyInitArenaConfig = true
	arenaConfigs = make(map[string]string)
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

	log.Debugf("Load arena config file %s", envVarFileName)

	arenaConfigs = config.ReadConfigFile(envVarFileName)
	log.Debugf("arena configs: %v", arenaConfigs)
}
