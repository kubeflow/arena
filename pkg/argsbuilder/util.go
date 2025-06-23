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

package argsbuilder

import (
	"errors"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
)

func transformSliceToMap(sets []string, split string) (valuesMap map[string]string) {
	valuesMap = map[string]string{}
	for _, member := range sets {
		splits := strings.SplitN(member, split, 2)
		if len(splits) == 2 {
			valuesMap[splits[0]] = splits[1]
		}
	}

	return valuesMap
}

func parseTolerationString(toleration string) (*types.TolerationArgs, error) {
	if strings.Contains(toleration, "=") && strings.Contains(toleration, ":") && strings.Contains(toleration, ",") {
		key, toleration := split(toleration, "=")
		value, toleration := split(toleration, ":")
		effect, operator := split(toleration, ",")
		return &types.TolerationArgs{
			Key:      key,
			Value:    value,
			Effect:   effect,
			Operator: operator,
		}, nil
	} else {
		return nil, errors.New("invalid toleration format")
	}
}

func split(value, sep string) (string, string) {
	index := strings.Index(value, sep)
	return value[:index], value[index+1:]
}
