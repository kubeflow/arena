// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/commands"
)

func main() {

	if isPProfEnabled() {
		cpuf, err := os.Create("/tmp/cpu_profile")
		if err != nil {
			log.Fatal(err)
		}
		defer cpuf.Close()

		runtime.SetCPUProfileRate(getProfileHZ())
		err = pprof.StartCPUProfile(cpuf)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Dump cpu profile file into /tmp/cpu_profile")
		defer pprof.StopCPUProfile()
	}

	// debug latency issue
	if isTraceEnabled() {
		tracef, err := os.Create("/tmp/trace.log")
		if err != nil {
			log.Fatal(err)
		}
		defer tracef.Close()

		err = trace.Start(tracef)
		if err != nil {
			log.Fatal(err)
		}
		defer trace.Stop()
	}

	if err := commands.NewCommand().Execute(); err != nil {
		utils.PrintErrorMessage(err.Error())
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

func getProfileHZ() int {
	profileRate := 1000
	if s, err := strconv.Atoi(os.Getenv("PROFILE_RATE")); err == nil {
		profileRate = s
	}
	return profileRate
}

func isTraceEnabled() (enable bool) {
	for _, arg := range os.Args {
		if arg == "--trace" {
			enable = true
			break
		}
	}

	return
}
