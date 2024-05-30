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

func (l *LoggerBuilder) Container(name string) *LoggerBuilder {
	l.args.ContainerName = name
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
