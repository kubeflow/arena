package commands

import (
	"os"

	config "github.com/kubeflow/arena/pkg/util/config"
	homedir "github.com/mitchellh/go-homedir"
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
	log.Debugf("init arena config")
	if alreadyInitArenaConfig {
		return
	}
	alreadyInitArenaConfig = true
	arenaConfigs = make(map[string]string)
	envVarFileName := os.Getenv(RecommendedConfigPathEnvVar)
	_, err := os.Stat(envVarFileName)
	if err != nil {
		log.Debugf("illegal arena config file: %s due to %v", envVarFileName, err)
		envVarFileName, err = homedir.Expand(DefaultArenaConfigPath)
		if err != nil {
			log.Debugf("fail to get arena config file: %s due to %v", envVarFileName, err)
			return
		}
	}

	_, err = os.Stat(envVarFileName)
	if err != nil {
		log.Debugf("illegal arena config file: %s due to %v", envVarFileName, err)
		return
	}

	log.Debugf("load arena config file %s", envVarFileName)

	arenaConfigs = config.ReadConfigFile(envVarFileName)
	log.Debugf("arena configs: %v", arenaConfigs)
}
