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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	errInvalidSinceSecond = errors.New("failed to parse since seconds,invalid format,should like: ['1h','1m','1s','1h1m1s'...]")
)

type LogArgsBuilder struct {
	args        *types.LogArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewLogArgsBuilder(args *types.LogArgs) ArgsBuilder {
	return &LogArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
}

func (l *LogArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*l)), ".")
	return items[len(items)-1]
}

func (l *LogArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		l.subBuilders[b.GetName()] = b
	}
	return l
}

func (l *LogArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range l.subBuilders {
		l.subBuilders[name].AddArgValue(key, value)
	}
	l.argValues[key] = value
	return l
}

func (l *LogArgsBuilder) PreBuild() error {
	for name := range l.subBuilders {
		if err := l.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (l *LogArgsBuilder) Build() error {
	for name := range l.subBuilders {
		if err := l.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	return l.transfer()
}

func (l *LogArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range l.subBuilders {
		l.subBuilders[name].AddCommandFlags(command)
	}
	var since string
	var sinceTime string
	var tail int64
	command.Flags().BoolVarP(&l.args.Follow, "follow", "f", false, "Specify if the logs should be streamed.")

	command.Flags().StringVar(&since, "since", "", "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVar(&sinceTime, "since-time", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().Int64VarP(&tail, "tail", "t", -1, "Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines otherwise 10, if a selector is provided.")
	command.Flags().BoolVar(&l.args.Timestamps, "timestamps", false, "Include timestamps on each line in the log output")
	// command.Flags().StringVar(&printer.pod, "instance", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVarP(&l.args.InstanceName, "instance", "i", "", "Specify the task instance to get log")
	command.Flags().StringVarP(&l.args.ContainerName, "container", "c", "", "Specify the container name of instance to get log")

	l.AddArgValue("since", &since).
		AddArgValue("since-time", &sinceTime).
		AddArgValue("tail", &tail)
}

func (l *LogArgsBuilder) transfer() error {
	// parse tail lines
	if value, ok := l.argValues["tail"]; ok {
		tail := value.(*int64)
		if *tail > int64(0) {
			l.args.Tail = tail
		}
	}
	// parse since time
	if value, ok := l.argValues["since-time"]; ok {
		sinceTime := value.(*string)
		st, err := ParseSinceTime(*sinceTime)
		if err != nil {
			return err
		}
		l.args.SinceTime = st
	}
	if value, ok := l.argValues["since"]; ok {
		sinceSeconds := value.(*string)
		ss, err := ParseSinceSeconds(*sinceSeconds)
		if err != nil {
			return err
		}
		l.args.SinceSeconds = ss
	}
	return nil
}

func ParseSinceTime(sinceTime string) (*metav1.Time, error) {
	if sinceTime == "" {
		return nil, nil
	}
	parsedTime, err := time.Parse(time.RFC3339, sinceTime)
	if err != nil {
		return nil, err
	}
	meta1Time := metav1.NewTime(parsedTime)
	return &meta1Time, nil

}

func ParseSinceSeconds(since string) (*int64, error) {
	if since == "" {
		return nil, nil
	}
	totalSeconds := int64(0)
	items := []string{}
	for i := 0; i < len(since); i++ {
		switch string(since[i]) {
		case "h":
			hour, err := strconv.ParseInt(strings.Join(items, ""), 10, 64)
			if err != nil {
				return nil, errInvalidSinceSecond
			}
			totalSeconds = totalSeconds + hour*3600
			items = []string{}
		case "m":
			m, err := strconv.ParseInt(strings.Join(items, ""), 10, 64)
			if err != nil {
				return nil, errInvalidSinceSecond
			}
			totalSeconds = totalSeconds + m*60
			items = []string{}
		case "s":
			second, err := strconv.ParseInt(strings.Join(items, ""), 10, 64)
			if err != nil {
				return nil, errInvalidSinceSecond
			}
			totalSeconds = totalSeconds + second
			items = []string{}
		default:
			items = append(items, string(since[i]))
		}
	}
	if len(items) != 0 {
		return nil, errInvalidSinceSecond
	}
	return &totalSeconds, nil
}
