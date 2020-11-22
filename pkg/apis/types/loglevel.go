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
	Namespace    string
	JobName      string
	InstanceName string
	Follow       bool
	SinceSeconds *int64
	SinceTime    *metav1.Time
	Tail         *int64
	Timestamps   bool
	RetryCnt     int
	RetryTimeout time.Duration
	WriterCloser io.WriteCloser
}
