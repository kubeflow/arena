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
package version

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
)

// Version information set by link flags during build. We fall back to these sane
// default values when we build outside the Makefile context (e.g. go build or go test).
var (
	buildDate   = "1970-01-01T00:00:00Z" // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
	gitCommit   = ""                     // output from `git rev-parse HEAD`
	gitTag      = ""                     // output from `git describe --exact-match --tags HEAD` (if clean tree state)
	versionFile = ""
)

// Version contains Arena version information
type Version struct {
	Version   string
	BuildDate string
	GitCommit string
	GitTag    string
	GoVersion string
	Compiler  string
	Platform  string
}

func (v Version) String() string {
	return v.Version
}

// GetVersion returns the version information
func GetVersion() Version {
	versionStr := ""
	file, err := os.Open(versionFile)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			versionStr = scanner.Text()
		}
	}

	return Version{
		Version:   versionStr,
		BuildDate: buildDate,
		GitCommit: gitCommit,
		GitTag:    gitTag,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
