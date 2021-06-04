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
package arena

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/helm"
)

// Version information set by link flags during build. We fall back to these sane
// default values when we build outside the Makefile context (e.g. go build or go test).
var (
	version      = "0.0.0"                // value from VERSION file
	buildDate    = "1970-01-01T00:00:00Z" // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
	gitCommit    = ""                     // output from `git rev-parse HEAD`
	gitTag       = ""                     // output from `git describe --exact-match --tags HEAD` (if clean tree state)
	gitTreeState = ""                     // determined from `git status --porcelain`. either 'clean' or 'dirty'
)

// Version contains Arena version information
type Version struct {
	Version      string
	BuildDate    string
	GitCommit    string
	GitTag       string
	GitTreeState string
	GoVersion    string
	Compiler     string
	Platform     string
	ChartsInfo   ChartsInfo
}

type ChartsInfo struct {
	ChartsVersion map[string]string
	ChartsHome    string
}

func (v Version) String() string {
	return v.Version
}

// GetVersion returns the version information
func GetVersion() Version {
	var versionStr string
	if gitCommit != "" && gitTag != "" && gitTreeState == "clean" {
		// if we have a clean tree state and the current commit is tagged,
		// this is an official release.
		versionStr = gitTag
	} else {
		// otherwise formulate a version string based on as much metadata
		// information we have available.
		versionStr = "v" + version
		if len(gitCommit) >= 7 {
			versionStr += "+" + gitCommit[0:7]
			if gitTreeState != "clean" {
				versionStr += ".dirty"
			}
		} else {
			versionStr += "+unknown"
		}
	}
	return Version{
		Version:      versionStr,
		BuildDate:    buildDate,
		GitCommit:    gitCommit,
		GitTag:       gitTag,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		ChartsInfo:   getChartsInfo(),
	}
}

func getChartsInfo() ChartsInfo {
	chartsFolder := util.GetChartsFolder()
	charts := []string{}
	chartFolder, err := ioutil.ReadDir(chartsFolder)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, c := range chartFolder {
			if c.IsDir() && c.Name() != "kubernetes-artifacts" {
				if !util.StringInSlice(c.Name(), charts) {
					charts = append(charts, c.Name())
				}
			}
		}
	}

	chartMap := make(map[string]string, len(charts))
	for _, c := range charts {
		chart := chartsFolder + "/" + c
		chartName := helm.GetChartName(chart)
		chartVersion, err := helm.GetChartVersion(chart)
		if err != nil {
			chartMap[chartName] = ""
		} else {
			chartMap[chartName] = chartVersion
		}
	}
	return ChartsInfo{
		ChartsVersion: chartMap,
		ChartsHome:    chartsFolder,
	}
}
