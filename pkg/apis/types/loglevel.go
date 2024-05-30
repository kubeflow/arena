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

package types

import (
	"io"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LogLevel string

const (
	LogDebug   LogLevel = "debug"
	LogInfo    LogLevel = "info"
	LogWarning LogLevel = "warn"
	LogError   LogLevel = "error"
	LogUnknown LogLevel = "unknown"
)

type LogArgs struct {
	Namespace     string
	JobName       string
	InstanceName  string
	ContainerName string
	Follow        bool
	SinceSeconds  *int64
	SinceTime     *metav1.Time
	Tail          *int64
	Timestamps    bool
	RetryCnt      int
	RetryTimeout  time.Duration
	WriterCloser  io.WriteCloser
}
