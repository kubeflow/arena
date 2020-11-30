package training

import (
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type SparkJobBuilder struct {
	args      *types.SubmitSparkJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewSparkJobBuilder() *SparkJobBuilder {
	args := &types.SubmitSparkJobArgs{
		Image:     "registry.aliyuncs.com/acs/spark:v2.4.0",
		MainClass: "org.apache.spark.examples.SparkPi",
		Jar:       "local:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar",
		Driver: types.Driver{
			CPURequest:    1,
			MemoryRequest: "500m",
		},
		Executor: types.Executor{
			Replicas:      1,
			CPURequest:    1,
			MemoryRequest: "500m",
		},
	}
	return &SparkJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitSparkJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *SparkJobBuilder) Name(name string) *SparkJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

func (b *SparkJobBuilder) Image(image string) *SparkJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

func (b *SparkJobBuilder) ExecutorReplicas(replicas int) *SparkJobBuilder {
	if replicas > 0 {
		b.args.Executor.Replicas = replicas
	}
	return b
}

func (b *SparkJobBuilder) MainClass(mainClass string) *SparkJobBuilder {
	if mainClass != "" {
		b.args.MainClass = mainClass
	}
	return b
}

func (b *SparkJobBuilder) Jar(jar string) *SparkJobBuilder {
	if jar != "" {
		b.args.Jar = jar
	}
	return b
}

func (b *SparkJobBuilder) DriverCPURequest(request int) *SparkJobBuilder {
	if request > 0 {
		b.args.Driver.CPURequest = request
	}
	return b
}

func (b *SparkJobBuilder) DriverMemoryRequest(memory string) *SparkJobBuilder {
	if memory != "" {
		b.args.Driver.MemoryRequest = memory
	}
	return b
}

func (b *SparkJobBuilder) ExecutorCPURequest(request int) *SparkJobBuilder {
	if request > 0 {
		b.args.Executor.CPURequest = request
	}
	return b
}

func (b *SparkJobBuilder) ExecutorMemoryRequest(memory string) *SparkJobBuilder {
	if memory != "" {
		b.args.Executor.MemoryRequest = memory
	}
	return b
}

// Build is used to build the job
func (b *SparkJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.SparkTrainingJob, b.args), nil
}
