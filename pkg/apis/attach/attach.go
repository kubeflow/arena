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

package attach

import (
	"os"

	"github.com/kubeflow/arena/pkg/argsbuilder"

	"github.com/kubeflow/arena/pkg/podexec"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type AttachBuilder struct {
	args      *podexec.AttachPodArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewAttachArgsBuilder() *AttachBuilder {
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	args := &podexec.AttachPodArgs{
		Options: &podexec.ExecOptions{
			StreamOptions: podexec.StreamOptions{
				IOStreams: ioStreams,
			},
			Executor: &podexec.DefaultRemoteExecutor{},
		},
		Command:          []string{"sh"},
		CmdArgsLenAtDash: 0,
	}
	return &AttachBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewAttachPodArgsBuilder(args),
	}
}

func (a *AttachBuilder) PodName(name string) *AttachBuilder {
	a.args.Options.PodName = name
	return a
}

func (a *AttachBuilder) ContainerName(name string) *AttachBuilder {
	a.args.Options.ContainerName = name
	return a
}

func (a *AttachBuilder) IOStreams(stream genericclioptions.IOStreams) *AttachBuilder {
	a.args.Options.IOStreams = stream
	return a
}

func (a *AttachBuilder) Command(command []string) *AttachBuilder {
	if len(command) != 0 {
		a.args.Command = command
	}
	return a
}

func (a *AttachBuilder) CmdArgsLenAtDash(length int) *AttachBuilder {
	a.args.CmdArgsLenAtDash = length
	return a
}

func (a *AttachBuilder) Build() (*podexec.AttachPodArgs, error) {
	for key, value := range a.argValues {
		a.AddArgValue(key, value)
	}
	if err := a.PreBuild(); err != nil {
		return nil, err
	}
	if err := a.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return a.args, nil
}
