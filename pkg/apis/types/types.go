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
	// Unknwon defines the unknown format
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
