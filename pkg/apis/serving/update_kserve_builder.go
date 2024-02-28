package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type UpdateKServeJobBuilder struct {
	args      *types.UpdateKServeArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewUpdateKServeJobBuilder() *UpdateKServeJobBuilder {
	args := &types.UpdateKServeArgs{
		CommonUpdateServingArgs: types.CommonUpdateServingArgs{
			Replicas: 1,
		},
	}
	return &UpdateKServeJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewUpdateKServeArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *UpdateKServeJobBuilder) Name(name string) *UpdateKServeJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Namespace is used to set job namespace,match option --namespace
func (b *UpdateKServeJobBuilder) Namespace(namespace string) *UpdateKServeJobBuilder {
	if namespace != "" {
		b.args.Namespace = namespace
	}
	return b
}

// Version is used to set serving job version, match the option --version
func (b *UpdateKServeJobBuilder) Version(version string) *UpdateKServeJobBuilder {
	if version != "" {
		b.args.Version = version
	}
	return b
}

// Command is used to set job command
func (b *UpdateKServeJobBuilder) Command(args []string) *UpdateKServeJobBuilder {
	b.args.Command = strings.Join(args, " ")
	return b
}

// Image is used to set job image,match the option --image
func (b *UpdateKServeJobBuilder) Image(image string) *UpdateKServeJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *UpdateKServeJobBuilder) Envs(envs map[string]string) *UpdateKServeJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// Annotations is used to add annotations for job pods,match option --annotation
func (b *UpdateKServeJobBuilder) Annotations(annotations map[string]string) *UpdateKServeJobBuilder {
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
func (b *UpdateKServeJobBuilder) Labels(labels map[string]string) *UpdateKServeJobBuilder {
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
func (b *UpdateKServeJobBuilder) Replicas(count int) *UpdateKServeJobBuilder {
	if count > 0 {
		b.args.Replicas = count
	}
	return b
}

// ModelFormat the ModelFormat being served.
func (b *UpdateKServeJobBuilder) ModelFormat(modelFormat *types.ModelFormat) *UpdateKServeJobBuilder {
	if modelFormat != nil {
		b.args.ModelFormat = modelFormat
	}
	return b
}

// Runtime specific ClusterServingRuntime/ServingRuntime name to use for deployment.
func (b *UpdateKServeJobBuilder) Runtime(runtime string) *UpdateKServeJobBuilder {
	if runtime != "" {
		b.args.Runtime = runtime
	}
	return b
}

// StorageUri is used to set storage uri,match the option --storage-uri
func (b *UpdateKServeJobBuilder) StorageUri(uri string) *UpdateKServeJobBuilder {
	if uri != "" {
		b.args.StorageUri = uri
	}
	return b
}

// RuntimeVersion of the predictor docker image
func (b *UpdateKServeJobBuilder) RuntimeVersion(runtimeVersion string) *UpdateKServeJobBuilder {
	if runtimeVersion != "" {
		b.args.RuntimeVersion = runtimeVersion
	}
	return b
}

// ProtocolVersion use by the predictor (i.e. v1 or v2 or grpc-v1 or grpc-v2)
func (b *UpdateKServeJobBuilder) ProtocolVersion(protocolVersion string) *UpdateKServeJobBuilder {
	if protocolVersion != "" {
		b.args.ProtocolVersion = protocolVersion
	}
	return b
}

// MinReplicas number of replicas, defaults to 1 but can be set to 0 to enable scale-to-zero.
func (b *UpdateKServeJobBuilder) MinReplicas(minReplicas int) *UpdateKServeJobBuilder {
	if minReplicas >= 0 {
		b.args.MinReplicas = minReplicas
	}
	return b
}

// MaxReplicas number of replicas for autoscaling.
func (b *UpdateKServeJobBuilder) MaxReplicas(maxReplicas int) *UpdateKServeJobBuilder {
	if maxReplicas > 0 {
		b.args.MaxReplicas = maxReplicas
	}
	return b
}

// ScaleTarget number of replicas for autoscaling.
func (b *UpdateKServeJobBuilder) ScaleTarget(scaleTarget int) *UpdateKServeJobBuilder {
	if scaleTarget > 0 {
		b.args.ScaleTarget = scaleTarget
	}
	return b
}

// ScaleMetric watched by autoscaler. possible values are concurrency, rps, cpu, memory. concurrency, rps are supported via KPA
func (b *UpdateKServeJobBuilder) ScaleMetric(scaleMetric string) *UpdateKServeJobBuilder {
	if scaleMetric != "" {
		b.args.ScaleMetric = scaleMetric
	}
	return b
}

// ContainerConcurrency specifies how many requests can be processed concurrently
func (b *UpdateKServeJobBuilder) ContainerConcurrency(containerConcurrency int64) *UpdateKServeJobBuilder {
	if containerConcurrency > 0 {
		b.args.ContainerConcurrency = containerConcurrency
	}
	return b
}

// TimeoutSeconds specifies the number of seconds to wait before timing out a request to the component.
func (b *UpdateKServeJobBuilder) TimeoutSeconds(timeoutSeconds int64) *UpdateKServeJobBuilder {
	if timeoutSeconds > 0 {
		b.args.TimeoutSeconds = timeoutSeconds
	}
	return b
}

// CanaryTrafficPercent defines the traffic split percentage between the candidate revision and the last ready revision
func (b *UpdateKServeJobBuilder) CanaryTrafficPercent(canaryTrafficPercent int64) *UpdateKServeJobBuilder {
	if canaryTrafficPercent > 0 {
		b.args.CanaryTrafficPercent = canaryTrafficPercent
	}
	return b
}

// Port the port of tcp listening port, default is 8080 in kserve
func (b *UpdateKServeJobBuilder) Port(port int) *UpdateKServeJobBuilder {
	if port > 0 {
		b.args.Port = port
	}
	return b
}

// Build is used to build the job
func (b *UpdateKServeJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.KServeJob, b.args), nil
}
