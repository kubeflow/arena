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

package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
)

func ValidateAndParseTags(tagsStr string) (map[string]string, error) {
	tags := map[string]string{}
	if len(tagsStr) == 0 {
		return tags, nil
	}

	regexp1 := regexp.MustCompile(`^([/\w]+)=([/\w]+-?)$`)
	regexp2 := regexp.MustCompile(`^([/\w]+-?)`)

	pairs := strings.Split(tagsStr, ",")
	for _, pair := range pairs {
		if regexp1.MatchString(pair) {
			matches := regexp1.FindStringSubmatch(pair)
			tags[matches[1]] = matches[2]
			continue
		}

		if regexp2.MatchString(pair) {
			matches := regexp2.FindStringSubmatch(pair)
			tags[matches[1]] = ""
			continue
		}

		return nil, errors.New("invalid format, should be like key1,key2=value2,key3-,key4=value4-")
	}

	return tags, nil
}

func PrintRegisteredModel(rm *types.RegisteredModel) {
	fmt.Printf("%-20s %s\n", "Name:", rm.Name)
	if len(rm.LatestVersions) == 0 {
		fmt.Printf("%-20s %s\n", "LatestVersion", "-")
	} else {
		fmt.Printf("%-20s %s\n", "LatestVersion", rm.LatestVersions[0].Version)
	}
	timeFormat := "2006-01-02T15:04:05Z07:00"
	fmt.Printf("%-20s %s\n", "CreationTime:", time.Unix(0, rm.CreationTimestamp*1000000).Local().Format(timeFormat))
	fmt.Printf("%-20s %s\n", "LastUpdatedTime:", time.Unix(0, rm.LastUpdatedTimestamp*1000000).Local().Format(timeFormat))
	fmt.Println("Description:")
	fmt.Printf("  %s\n", strings.ReplaceAll(rm.Description, "\n", "\n  "))
	fmt.Println("Tags:")
	for _, tag := range rm.Tags {
		fmt.Printf("  %s: %s\n", tag.Key, tag.Value)
	}
}

func PrintRegisteredModels(registeredModels []*types.RegisteredModel) {
	format := "%-20s %-20s %-80s\n"
	fmt.Printf(format, "NAME", "LATEST_VERSION", "LAST_UPDATED_TIME")
	for _, rm := range registeredModels {
		lastUpdatedTime := time.Unix(0, rm.LastUpdatedTimestamp*1000000).Local().Format("2006-01-02T15:04:05Z07:00")
		if len(rm.LatestVersions) == 0 {
			fmt.Printf(format, rm.Name, "-", lastUpdatedTime)
		} else {
			fmt.Printf(format, rm.Name, rm.LatestVersions[len(rm.LatestVersions)-1].Version, lastUpdatedTime)
		}
	}
}

func PrintModelVersion(mv *types.ModelVersion) {
	fmt.Printf("%-20s %s\n", "Name:", mv.Name)
	fmt.Printf("%-20s %s\n", "Version:", mv.Version)
	timeFormat := "2006-01-02T15:04:05Z07:00"
	fmt.Printf("%-20s %s\n", "CreationTime:", time.Unix(0, mv.CreationTimestamp*1000000).Local().Format(timeFormat))
	fmt.Printf("%-20s %s\n", "LastUpdatedTime:", time.Unix(0, mv.LastUpdatedTimestamp*1000000).Local().Format(timeFormat))
	fmt.Printf("%-20s %s\n", "Source:", mv.Source)
	fmt.Println("Description:")
	fmt.Printf("  %s\n", strings.ReplaceAll(mv.Description, "\n", "\n  "))
	fmt.Println("Tags:")
	for _, tag := range mv.Tags {
		fmt.Printf("  %s: %s\n", tag.Key, tag.Value)
	}
}

func PrintModelVersions(modelVersions []*types.ModelVersion) {
	format := "  %-10s %s\n"
	fmt.Println("Versions:")
	fmt.Printf(format, "Version", "Source")
	fmt.Printf(format, "---", "---")
	for _, mv := range modelVersions {
		fmt.Printf(format, mv.Version, mv.Source)
	}
}

func ReadUserConfirmation(prompt string) bool {
	fmt.Println(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(response)
	return response == "yes"
}
