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
	a.args.Options.StreamOptions.IOStreams = stream
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
