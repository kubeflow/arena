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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrintFormatStyle defines the format of output
// it only used in cmd
type FormatStyle string

const (
	// Wide defines the wide format
	WideFormat FormatStyle = "wide"
	// Json defines the json format
	JsonFormat FormatStyle = "json"
	// Yaml defines the yaml format
	YamlFormat FormatStyle = "yaml"
	// Unknown defines the unknown format
	UnknownFormat FormatStyle = "unknown"
)

type ArenaClientArgs struct {
	Kubeconfig     string
	Namespace      string
	ArenaNamespace string
	IsDaemonMode   bool
	LogLevel       string
}

type K8sObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}
