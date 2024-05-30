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

package env

import (
	"bufio"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ReadEnvFile returns configs map
func ReadConfigFile(filename string) (configs map[string]string) {
	configs = make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		log.Debugf("load config file: %s due to error %v", filename, err)
		return
	}

	for _, line := range lines {
		if !canIgnore(line) {
			var key, value string
			splitString := strings.SplitN(line, "=", 2)
			if len(splitString) != 2 {
				continue
			}
			key = strings.Trim(splitString[0], " ")
			value = strings.Trim(splitString[1], " ")
			configs[key] = value
		}
	}

	return
}

func canIgnore(line string) bool {
	trimmedLine := strings.Trim(line, " \n\t")
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
