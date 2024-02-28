package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type UpdateCustomServingJobBuilder struct {
	args      *types.UpdateCustomServingArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewUpdateCustomServingJobBuilder() *UpdateCustomServingJobBuilder {
	args := &types.UpdateCustomServingArgs{
		CommonUpdateServingArgs: types.CommonUpdateServingArgs{
			Replicas: 1,
		},
	}
	return &UpdateCustomServingJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewUpdateCustomServingArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *UpdateCustomServingJobBuilder) Name(name string) *UpdateCustomServingJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *UpdateCustomServingJobBuilder) Namespace(namespace string) *UpdateCustomServingJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Version is used to set serving job version, match the option --version
func (b *UpdateCustomServingJobBuilder) Version(version string) *UpdateCustomServingJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Command is used to set job command
func (b *UpdateCustomServingJobBuilder) Command(args []string) *UpdateCustomServingJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// Image is used to set job image,match the option --image
func (b *UpdateCustomServingJobBuilder) Image(image string) *UpdateCustomServingJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *UpdateCustomServingJobBuilder) Envs(envs map[string]string) *UpdateCustomServingJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// Tolerations are used to set tolerations for tolerate nodes, match option --toleration
func (b *UpdateCustomServingJobBuilder) Tolerations(tolerations []string) *UpdateCustomServingJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// NodeSelectors is used to set node selectors for scheduling job, match option --selector
func (b *UpdateCustomServingJobBuilder) NodeSelectors(selectors map[string]string) *UpdateCustomServingJobBuilder {
	if len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["selector"] = &selectorsSlice
	}
	return b
}

// Annotations is used to add annotations for job pods,match option --annotation
func (b *UpdateCustomServingJobBuilder) Annotations(annotations map[string]string) *UpdateCustomServingJobBuilder {
	if len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["annotation"] = &s
	}
	return b
}

// Labels is used to add labels for job
func (b *UpdateCustomServingJobBuilder) Labels(labels map[string]string) *UpdateCustomServingJobBuilder {
	if len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["label"] = &s
	}
	return b
}

// Replicas is used to set serving job replicas,match the option --replicas
func (b *UpdateCustomServingJobBuilder) Replicas(count int) *UpdateCustomServingJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// Build is used to build the job
func (b *UpdateCustomServingJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.CustomServingJob, b.args), nil
}
