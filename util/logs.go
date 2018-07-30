package util

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// SetLogLevel sets the logrus logging level
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Fatalf("Unknown level: %s", level)
	}
}
