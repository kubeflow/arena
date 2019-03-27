package commands

import (
	"os"

	"github.com/labstack/gommon/log"
	log "github.com/sirupsen/logrus"
)

const (
	RecommendedConfigPathEnvVar = "ARENA_CONFIG"
	DefaultArenaConfigPath      = "~/.arena/config"
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

}

func readEnvConfigFile(filename string) (configs map[string]string) {
	configs = make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		log.Debugf("Illegal arena config file: %s due to %v", filename, err)
		return configs
	}
	defer file.Close()
}
