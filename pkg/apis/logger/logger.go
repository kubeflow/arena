package logger

import (
	"io"
	"os"
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type LoggerBuilder struct {
	args      *types.LogArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewLoggerBuilder() *LoggerBuilder {
	args := &types.LogArgs{
		WriterCloser: os.Stdout,
		RetryCnt:     5,
		RetryTimeout: time.Second,
	}
	return &LoggerBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewLogArgsBuilder(args),
	}
}
func (l *LoggerBuilder) SinceSeconds(sinceSeconds string) *LoggerBuilder {
	if sinceSeconds != "" {
		l.argValues["since"] = &sinceSeconds
	}
	return l
}

func (l *LoggerBuilder) SinceTime(sinceTime string) *LoggerBuilder {
	if sinceTime != "" {
		l.argValues["since-time"] = &sinceTime
	}
	return l
}

func (l *LoggerBuilder) Instance(name string) *LoggerBuilder {
	l.args.InstanceName = name
	return l
}

func (l *LoggerBuilder) Follow() *LoggerBuilder {
	l.args.Follow = true
	return l
}

func (l *LoggerBuilder) EnableTimestamp() *LoggerBuilder {
	l.args.Timestamps = true
	return l
}

func (l *LoggerBuilder) Tail(line int) *LoggerBuilder {
	if line > 0 {
		lineInt64 := int64(line)
		l.argValues["tail"] = &lineInt64
	}
	return l
}

func (l *LoggerBuilder) WriterCloser(writerCloser io.WriteCloser) *LoggerBuilder {
	if writerCloser != nil {
		l.args.WriterCloser = writerCloser
	}
	return l
}

func (l *LoggerBuilder) Build() (*types.LogArgs, error) {
	for key, value := range l.argValues {
		l.AddArgValue(key, value)
	}
	if err := l.PreBuild(); err != nil {
		return nil, err
	}
	if err := l.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return l.args, nil
}
