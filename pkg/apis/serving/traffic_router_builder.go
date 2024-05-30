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

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type TrafficRouterBuilder struct {
	args      *types.TrafficRouterSplitArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewTrafficRouterBuilder() *TrafficRouterBuilder {
	args := &types.TrafficRouterSplitArgs{
		Namespace: "default",
	}
	return &TrafficRouterBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewTrafficRouterArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *TrafficRouterBuilder) Name(name string) *TrafficRouterBuilder {
	if name != "" {
		b.args.ServingName = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *TrafficRouterBuilder) Namespace(namespace string) *TrafficRouterBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// VersionWeight is used to set version weight
func (b *TrafficRouterBuilder) VersionWeight(weights []types.ServingVersionWeight) *TrafficRouterBuilder {
	if len(weights) != 0 {
		versionWeithts := []string{}
		for _, v := range weights {
			versionWeithts = append(versionWeithts, fmt.Sprintf("%v:%v", v.Version, v.Weight))
		}
		b.argValues["version-weight"] = &versionWeithts
	}
	return b
}

// Build is used to build the traffic router split args
func (b *TrafficRouterBuilder) Build() (*types.TrafficRouterSplitArgs, error) {
	if b.args.Namespace == "" {
		return nil, fmt.Errorf("not set namespace,please set it")
	}
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return b.args, nil
}
