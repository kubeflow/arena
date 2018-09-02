package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/kubeflow/arena/cmd/arena/commands"
	log "github.com/sirupsen/logrus"
)

func main() {

	if isPProfEnabled() {
		cpuf, err := os.Create("/tmp/cpu_profile")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(cpuf)
		log.Infof("Dump cpu profile file into /tmp/cpu_profile")
		defer pprof.StopCPUProfile()
	}

	if err := commands.NewCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func isPProfEnabled() (enable bool) {
	for _, arg := range os.Args {
		if arg == "--pprof" {
			enable = true
			break
		}
	}

	return
}
