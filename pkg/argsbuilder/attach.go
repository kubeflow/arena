package argsbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/podexec"
	"github.com/spf13/cobra"
)

type AttachArgsBuilder struct {
	args        *podexec.AttachPodArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewAttachPodArgsBuilder(args *podexec.AttachPodArgs) ArgsBuilder {
	return &AttachArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
}

func (l *AttachArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*l)), ".")
	return items[len(items)-1]
}

func (l *AttachArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		l.subBuilders[b.GetName()] = b
	}
	return l
}

func (l *AttachArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range l.subBuilders {
		l.subBuilders[name].AddArgValue(key, value)
	}
	l.argValues[key] = value
	return l
}

func (l *AttachArgsBuilder) PreBuild() error {
	for name := range l.subBuilders {
		if err := l.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (l *AttachArgsBuilder) Build() error {
	for name := range l.subBuilders {
		if err := l.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	return nil
}

func (l *AttachArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range l.subBuilders {
		l.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().StringVarP(&l.args.Options.PodName, "instance", "i", "", "Job instance name")
	command.Flags().StringVarP(&l.args.Options.ContainerName, "container", "c", "", "Container name. If omitted, the first container in the instance will be chosen")
}
