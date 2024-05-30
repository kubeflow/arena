// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serving

import (
	"fmt"
	"io"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
)

func PrintLine(w io.Writer, fields ...string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

// print the help info  when jobs more than one
func moreThanOneJobHelpInfo(infos []ServingJob) string {
	header := fmt.Sprintf("There is %d jobs have been found:", len(infos))
	tableHeader := "NAME\tTYPE\tVERSION"
	lines := []string{tableHeader}
	footer := "please use '--type' or '--version' to filter."
	for _, info := range infos {
		line := fmt.Sprintf("%s\t%s\t%s",
			info.Name(),
			info.Type(),
			info.Version(),
		)
		lines = append(lines, line)
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(lines, "\n"), footer)
}

func moreThanOneInstanceHelpInfo(instances []types.ServingInstance) string {
	header := fmt.Sprintf("There is %d instances have been found:", len(instances))
	lines := []string{}
	footer := "please use '-i' or '--instance' to filter."
	for _, i := range instances {
		lines = append(lines, fmt.Sprintf("%v", i.Name))
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(lines, "\n"), footer)

}

func CheckJobIsOwnedByProcesser(labels map[string]string) bool {
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsIsolateUserInNamespace() && labels[types.UserNameIdLabel] != arenaConfiger.GetUser().GetId() {
		return false
	}
	return true
}

func ValidateJobsBeforeSubmiting(jobs []ServingJob, name string) error {
	if len(jobs) == 0 {
		log.Debugf("not found serving job %v,we will submit it", name)
		return nil
	}
	knownJobs := []ServingJob{}
	unknownJobs := []ServingJob{}
	for _, s := range jobs {
		var labels map[string]string
		if ksjob, ok := s.(*kserveJob); ok {
			labels = ksjob.inferenceService.Labels
		} else {
			labels = s.Deployment().Labels
		}
		if CheckJobIsOwnedByProcesser(labels) {
			knownJobs = append(knownJobs, s)
		} else {
			unknownJobs = append(unknownJobs, s)
		}
	}
	log.Debugf("total known jobs: %v,total unknown jobs: %v", len(knownJobs), len(unknownJobs))
	if len(knownJobs) != 0 {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena serve delete %s'", name, name)
	}
	if len(unknownJobs) != 0 {
		return fmt.Errorf("the job %v is already exist, but its' owner is not you", name)
	}
	return nil
}
