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
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"

	"github.com/kubeflow/arena/cmd/arena/commands"
	log "github.com/sirupsen/logrus"
)

const (
	indexOfMPI    = 2
	indexOfSubmit = 1
)

func main() {
	setSubmitMpiCommandIfNeeded()

	if isPProfEnabled() {
		cpuf, err := os.Create("/tmp/cpu_profile")
		if err != nil {
			log.Fatal(err)
		}
		defer cpuf.Close()

		runtime.SetCPUProfileRate(getProfileHZ())
		pprof.StartCPUProfile(cpuf)
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

// This method is used in order to hide "runai submit mpi" command - it was added this way in order to not change the way we run runai jobs "runai submit"
// Probably in the future this part will be removed.
func setSubmitMpiCommandIfNeeded() {
	if len(os.Args) >= indexOfMPI && os.Args[indexOfSubmit] == "submit" && os.Args[indexOfMPI] == "mpi" {
		os.Args[1] = commands.SubmitMpiCommand
		copy(os.Args[indexOfMPI:], os.Args[indexOfMPI+1:]) // Shift a[i+1:] left one index.
		os.Args[len(os.Args)-1] = ""                       // Erase last element (write zero value).
		os.Args = os.Args[:len(os.Args)-1]                 // Truncate slice.
	}
}
